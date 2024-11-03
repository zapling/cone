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
	_ cone.Response = &responseAndEvent{}
	_ Response      = &responseAndEvent{}
)

type Response interface {
	cone.Response
	NakWithDelay(delay time.Duration) error
}

func New(consumer jetstream.Consumer, opts ...jetstream.PullConsumeOpt) *Source {
	return &Source{consumer: consumer, opts: opts}
}

type Source struct {
	consumer       jetstream.Consumer
	consumeContext jetstream.ConsumeContext
	opts           []jetstream.PullConsumeOpt

	responseAndEvents chan *responseAndEvent
}

func (s *Source) Start() error {
	s.responseAndEvents = make(chan *responseAndEvent)
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
	close(s.responseAndEvents)

	return nil
}

func (s *Source) Next() (cone.Response, *cone.Event, error) {
	select {
	case responseEvent := <-s.responseAndEvents:
		return responseEvent, responseEvent.Event, nil
	case <-time.After(10 * time.Millisecond):
		return nil, nil, nil
	}
}

func (s *Source) messageHandler() func(jetstream.Msg) {
	return func(m jetstream.Msg) {
		event, err := cone.NewEvent(m.Subject(), m.Data())
		if err != nil {
			// TODO: Somehow not lose this info?
			_ = m.Nak()
			return
		}
		event.Header = cone.Header(m.Headers())
		s.responseAndEvents <- &responseAndEvent{Event: event, m: m}
	}
}

type responseAndEvent struct {
	*cone.Event
	m            jetstream.Msg
	responseSent bool
}

func (e *responseAndEvent) Ack() error {
	if e.responseSent {
		return nil
	}
	e.responseSent = true
	return e.m.Ack()
}

func (e *responseAndEvent) Nak() error {
	if e.responseSent {
		return nil
	}
	e.responseSent = true
	return e.m.Nak()
}

func (e *responseAndEvent) NakWithDelay(delay time.Duration) error {
	if e.responseSent {
		return nil
	}
	e.responseSent = true
	return e.m.NakWithDelay(delay)
}
