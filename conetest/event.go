package conetest

import "github.com/zapling/cone"

var _ cone.Event = &Event{}

func NewEvent(subject string, body []byte) *Event {
	return &Event{subject: subject, body: body}
}

type Event struct {
	subject string
	body    []byte
}

func (e *Event) Subject() string {
	return e.subject
}

func (e *Event) Body() []byte {
	return e.body
}
