package cone

import "context"

func NewEvent(subject string, body []byte) (*Event, error) {
	return NewEventWithContext(context.Background(), subject, body)
}

func NewEventWithContext(ctx context.Context, subject string, body []byte) (*Event, error) {
	return &Event{Subject: subject, Body: body, Header: make(Header), ctx: ctx}, nil
}

type Event struct {
	Subject string
	Body    []byte
	Header  Header
	ctx     context.Context
}

func (e *Event) Context() context.Context {
	return e.ctx
}

func (e *Event) WithContext(ctx context.Context) *Event {
	if ctx == nil {
		panic("nil context")
	}
	e2 := new(Event)
	*e2 = *e
	e2.ctx = ctx
	return e2
}

type Header map[string][]string

func (h Header) Set(key, value string) {
	h[key] = []string{value}
}

func (h Header) Add(key, value string) {
	h[key] = append(h[key], value)
}

func (h Header) Get(key string) string {
	values, ok := h[key]
	if !ok {
		return ""
	}
	return values[0]
}

func (h Header) Values(key string) []string {
	values, ok := h[key]
	if !ok {
		return nil
	}
	return values
}
