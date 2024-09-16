package conetest

import (
	"github.com/zapling/cone"
)

func NewEvent(subject string, body []byte) *cone.Event {
	event, err := cone.NewEvent(subject, body)
	if err != nil {
		panic(err)
	}
	return event
}
