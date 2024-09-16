package conetest

import (
	"context"

	"github.com/zapling/cone"
)

func NewEvent(subject string, body []byte) *cone.Event {
	return NewEventWithContext(context.Background(), subject, body)
}

func NewEventWithContext(ctx context.Context, subject string, body []byte) *cone.Event {
	event, err := cone.NewEventWithContext(ctx, subject, body)
	if err != nil {
		panic(err)
	}
	return event
}
