package conetest

import "github.com/zapling/cone"

var _ cone.Response = &ResponseRecorder{}

const (
	Ack = "ack"
	Nak = "nak"
)

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{}
}

type ResponseRecorder struct {
	response string
}

func (r *ResponseRecorder) Result() string {
	return r.response
}

func (r *ResponseRecorder) Ack() error {
	if r.response == "" {
		r.response = "ack"
	}
	return nil
}

func (r *ResponseRecorder) Nak() error {
	if r.response == "" {
		r.response = "nak"
	}
	return nil
}
