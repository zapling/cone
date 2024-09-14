package jetstream

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/zapling/cone"
)

var (
	_ cone.Source   = &Source{}
	_ cone.Event    = &event{}
	_ cone.Response = &event{}
)

func New(consumer jetstream.Consumer, opts ...jetstream.PullConsumeOpt) *Source {
	return &Source{consumer: consumer, opts: opts}
}

type Source struct {
	consumer       jetstream.Consumer
	consumeContext jetstream.ConsumeContext
	opts           []jetstream.PullConsumeOpt

	events chan *event
}

func (s *Source) Start() error {
	s.events = make(chan *event)
	consumeContext, err := s.consumer.Consume(s.messageHandler(), s.opts...)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	s.consumeContext = consumeContext

	return nil
}

func (s *Source) Stop(ctx context.Context) error {
	if s.consumeContext == nil {
		return fmt.Errorf("is not running")
	}

	s.consumeContext.Drain()

	select {
	case <-s.consumeContext.Closed():
		s.consumeContext.Stop()
	case <-ctx.Done():
		s.consumeContext.Stop()
	}

	s.consumeContext = nil
	close(s.events)

	return nil
}

func (s *Source) GetNextEvent() (cone.ResponseAndEvent, error) {
	select {
	case event := <-s.events:
		return event, nil
	case <-time.After(10 * time.Millisecond):
		return nil, nil
	}
}

func (s *Source) messageHandler() func(jetstream.Msg) {
	return func(m jetstream.Msg) {
		event := &event{m: m}
		s.events <- event
	}
}

type event struct {
	m jetstream.Msg
}

func (e *event) Subject() string {
	return e.m.Subject()
}

func (e *event) Body() []byte {
	return e.m.Data()
}

func (e *event) Ack() error {
	return e.m.Ack()
}

func (e *event) Nak() error {
	return e.m.Nak()
}
