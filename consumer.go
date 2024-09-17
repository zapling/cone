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

func New(source Source, handler Handler) *Consumer {
	return &Consumer{source: source, handler: handler}
}

type Consumer struct {
	source  Source
	handler Handler

	activeHandles sync.WaitGroup

	isRunning  atomic.Bool
	inShutdown atomic.Bool
}

func (c *Consumer) Serve(r Response, e *Event) {
	c.handler.Serve(r, e)
}

func (c *Consumer) ListenAndConsume() error {
	if c.isRunning.Load() {
		return fmt.Errorf("is already running")
	}

	if c.source == nil {
		return fmt.Errorf("source is nil")
	}

	if c.handler == nil {
		return fmt.Errorf("handler is nil")
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

		response, event, err := c.source.Next()
		if err != nil {
			return err
		}

		if event == nil {
			continue
		}

		c.activeHandles.Add(1)
		go func() {
			defer c.activeHandles.Done()
			c.Serve(response, event)
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
