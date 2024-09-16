package conetest

import (
	"context"
	"sync"

	"github.com/zapling/cone"
)

var _ cone.Source = &Source{}

func NewSource() *Source {
	return &Source{
		eventsMap: make(map[int]*sourceEvent),
	}
}

type Source struct {
	mu      sync.Mutex
	counter int

	events    []*sourceEvent
	ackEvents []*sourceEvent
	nakEvents []*sourceEvent

	eventsMap map[int]*sourceEvent
}

func (s *Source) Start() error {
	return nil // We have nothing to do here
}

func (s *Source) Stop(ctx context.Context) error {
	return nil // We have nothing to do here
}

func (s *Source) Next() (cone.Response, *cone.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < len(s.events); i++ {
		if s.events[i].isProcessing {
			continue
		}

		s.events[i].isProcessing = true
		return s.events[i], &s.events[i].Event, nil
	}

	return nil, nil, nil
}

func (s *Source) AddEvent(e *cone.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := &sourceEvent{Event: *e, source: s, id: s.counter}

	s.events = append(s.events, event)
	s.eventsMap[event.id] = event

	s.counter++
}

func (s *Source) NumAckd() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.ackEvents)
}

func (s *Source) NumNakd() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.ackEvents)
}

func (s *Source) ackEvent(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.removeEventFromQueue(id)
	s.ackEvents = append(s.ackEvents, s.eventsMap[id])

	return nil
}

func (s *Source) nakEvent(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.removeEventFromQueue(id)

	s.nakEvents = append(s.nakEvents, s.eventsMap[id])
	s.eventsMap[id].isProcessing = false

	return nil
}

func (s *Source) removeEventFromQueue(id int) {
	for i := 0; i < len(s.events); i++ {
		if s.events[i].id == id {
			s.events = append(s.events[:i], s.events[i+1:]...)
			break
		}
	}
}

type sourceEvent struct {
	cone.Event
	source *Source

	id           int
	isProcessing bool
	hasResponded bool
}

func (e *sourceEvent) Ack() error {
	if e.hasResponded {
		return nil
	}
	e.hasResponded = true
	return e.source.ackEvent(e.id)
}

func (e *sourceEvent) Nak() error {
	if e.hasResponded {
		return nil
	}
	e.hasResponded = true
	return e.source.nakEvent(e.id)
}
