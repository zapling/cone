package cone_test

import (
	"testing"

	"github.com/zapling/cone"
	"github.com/zapling/cone/conetest"
)

func TestNew(t *testing.T) {
	c := cone.New(nil)
	if c == nil {
		t.Fatalf("Expected Consumer got nil")
	}

	c.Handle("event.subject", nil)
	c.HandleFunc("event.subejct", nil)
}

func TestServeNilResponse(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatalf("Expected panic, empty Response is not allowed")
		}
	}()
	c := cone.New(nil)
	c.Serve(nil, nil) // This should panic but got none, empty Response is not allowed
}

func TestServeNilEvent(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatalf("Expected panic but got none, empty Event is not allowed")
		}
	}()
	c := cone.New(nil)
	c.Serve(conetest.NewRecorder(), nil) // This should panic, empty Event is not allowed
}

func TestServeNoSubjectHandlerNak(t *testing.T) {
	c := cone.New(nil)

	r := conetest.NewRecorder()

	c.Serve(r, conetest.NewEvent("event.subject", nil))

	if r.Result() != conetest.Nak {
		t.Fatalf("No subject handler should call nak the response")
	}
}

func TestServeDefaultAck(t *testing.T) {
	c := cone.New(nil)
	c.HandleFunc("event.subject", func(r cone.Response, e cone.Event) {})

	r := conetest.NewRecorder()

	c.Serve(r, conetest.NewEvent("event.subject", nil))

	if r.Result() != conetest.Ack {
		t.Fatalf("Expected Ack got: %s", r.Result())
	}
}
