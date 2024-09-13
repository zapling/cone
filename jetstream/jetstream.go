package jetstream

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/zapling/cone"
)

var (
	_ cone.Handler  = &Consumer{}
	_ cone.Event    = &EventResponse{}
	_ cone.Response = &EventResponse{}
)

func New(consumer jetstream.Consumer, opts ...jetstream.PullConsumeOpt) *Consumer {
	return &Consumer{consumer: consumer, opts: opts}
}

type Consumer struct {
	consumer       jetstream.Consumer
	consumeContext jetstream.ConsumeContext
	opts           []jetstream.PullConsumeOpt

	handlers  map[string]cone.Handler
	isRunning bool
}

func (c *Consumer) Serve(r cone.Response, e cone.Event) {
	handler, exists := c.handlers[e.Subject()]
	if !exists {
		if err := r.Nak(); err != nil {

		}
		return
	}

	handler.Serve(r, e)
	if err := r.Ack(); err != nil {

	}
}

func (c *Consumer) Handle(subject string, handler cone.Handler) {
	if c.isRunning {
		panic("Not allowed to add event handlers on a running consumer")
	}

	if c.handlers == nil {
		c.handlers = make(map[string]cone.Handler)
	}

	c.handlers[subject] = handler
}

func (c *Consumer) HandleFunc(subject string, handlerFunc cone.HandlerFunc) {
	c.Handle(subject, handlerFunc)
}

func (c *Consumer) Consume() error {
	if c.isRunning {
		return fmt.Errorf("consumer is already running")
	}

	consumeContext, err := c.consumer.Consume(
		func(m jetstream.Msg) {
			e := &EventResponse{m: m}
			c.Serve(e, e)
		},
		c.opts...,
	)
	if err != nil {
		return err
	}

	c.consumeContext = consumeContext
	c.isRunning = true

	return nil
}

func (c *Consumer) Shutdown(ctx context.Context) error {
	if !c.isRunning {
		return nil
	}

	return nil
}

type EventResponse struct {
	m               jetstream.Msg
	hasSentResponse bool
}

func (e *EventResponse) Subject() string {
	return e.m.Subject()
}

func (e *EventResponse) Body() []byte {
	return e.m.Data()
}

func (e *EventResponse) Ack() error {
	if e.hasSentResponse {
		return nil
	}

	e.hasSentResponse = true
	return e.m.Ack()
}

func (e *EventResponse) Nak() error {
	if e.hasSentResponse {
		return nil
	}

	e.hasSentResponse = true
	return e.m.Nak()
}
