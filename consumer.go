package cone

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	ErrConsumerStopped = errors.New("consumer stopped")
)

func New(source Source) *Consumer {
	return &Consumer{
		source:   source,
		handlers: make(map[string]Handler),
	}
}

type Consumer struct {
	source   Source
	handlers map[string]Handler

	activeHandles sync.WaitGroup

	isRunning  atomic.Bool
	inShutdown atomic.Bool
}

func (c *Consumer) Handle(subject string, handler Handler) {
	err := c.register(subject, handler)
	if err != nil {
		panic(err)
	}
}

func (c *Consumer) HandleFunc(subject string, handlerFunc HandlerFunc) {
	err := c.register(subject, handlerFunc)
	if err != nil {
		panic(err)
	}
}

func (c *Consumer) register(subject string, handler Handler) error {
	if c.isRunning.Load() {
		return fmt.Errorf("not allowed to register handler while running")
	}

	if subject == "" {
		return fmt.Errorf("empty subject is not allowed")
	}

	c.handlers[subject] = handler
	return nil
}

func (c *Consumer) Serve(r Response, e Event) {
	if err := c.serveEvent(r, e); err != nil {
		panic(err)
	}
}

func (c *Consumer) serveEvent(r Response, e Event) error {
	handler, ok := c.handlers[e.Subject()]
	if !ok {
		return r.Nak()
	}

	handler.Serve(r, e)
	return r.Ack()
}

func (c *Consumer) ListenAndConsume() error {
	if c.isRunning.Load() {
		return fmt.Errorf("is already running")
	}

	if c.source == nil {
		return fmt.Errorf("source is nil")
	}

	c.isRunning.Swap(true)
	defer c.isRunning.Swap(false)

	if err := c.source.Start(); err != nil {
		return fmt.Errorf("failed to start source: %w", err)
	}

	for {
		if c.inShutdown.Load() {
			return ErrConsumerStopped
		}

		event, err := c.source.GetNextEvent()
		if err != nil {
			return err
		}

		if event == nil {
			continue
		}

		c.activeHandles.Add(1)
		go func() {
			defer c.activeHandles.Done()
			c.Serve(event, event)
		}()
	}
}

func (c *Consumer) Shutdown(ctx context.Context) error {
	if !c.isRunning.Load() {
		return fmt.Errorf("consumer is not running")
	}

	err := c.source.Stop(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop source: %w", err)
	}

	c.inShutdown.Swap(true)

	// Wait for all active handles to finish
	if waitCtx(&c.activeHandles, ctx) {
		return ctx.Err() // Context was cancelled
	}

	return nil
}

func waitCtx(wg *sync.WaitGroup, ctx context.Context) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return false // Completed normally
	case <-ctx.Done():
		return true // Context was cancelled
	}
}
