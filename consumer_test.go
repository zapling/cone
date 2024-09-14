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
}

func TestHandle(t *testing.T) {
	t.Run("Empty subject should panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("Expected panic, empty subject is not allowed")
			}
		}()

		c := cone.New(nil)
		c.Handle("", nil)
	})

	t.Run("Nil handler is ok", func(t *testing.T) {
		c := cone.New(nil)
		c.Handle("event.subject", nil)
	})

	t.Run("Same subject should override previous handler", func(t *testing.T) {
		c := cone.New(nil)
		var handlerCalled string
		var firstHandler cone.HandlerFunc = func(_ cone.Response, _ cone.Event) { handlerCalled = "first" }
		var secondHandler cone.HandlerFunc = func(_ cone.Response, _ cone.Event) { handlerCalled = "second" }
		c.Handle("event.subject", firstHandler)
		c.Handle("event.subject", secondHandler)
		r := conetest.NewRecorder()
		c.Serve(r, conetest.NewEvent("event.subject", nil))
		if handlerCalled != "second" {
			t.Errorf("Second handler should have been called, but '%s' was called", handlerCalled)
		}
	})
}

func TestHandleFunc(t *testing.T) {
	t.Run("Empty subject should panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("Expected panic, empty subject is not allowed")
			}
		}()

		c := cone.New(nil)
		c.HandleFunc("", nil)
	})

	t.Run("Nil handler is ok", func(t *testing.T) {
		c := cone.New(nil)
		c.HandleFunc("event.subject", nil)
	})

	t.Run("Same subject should override previous handler", func(t *testing.T) {
		c := cone.New(nil)
		var handlerCalled string
		var firstHandler cone.HandlerFunc = func(_ cone.Response, _ cone.Event) { handlerCalled = "first" }
		var secondHandler cone.HandlerFunc = func(_ cone.Response, _ cone.Event) { handlerCalled = "second" }
		c.HandleFunc("event.subject", firstHandler)
		c.HandleFunc("event.subject", secondHandler)
		r := conetest.NewRecorder()
		c.Serve(r, conetest.NewEvent("event.subject", nil))
		if handlerCalled != "second" {
			t.Errorf("Second handler should have been called, but '%s' was called", handlerCalled)
		}
	})
}

func TestServe(t *testing.T) {
	t.Run("Nil Response should panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("Expected panic, empty Response is not allowed")
			}
		}()
		c := cone.New(nil)
		c.Serve(nil, conetest.NewEvent("event.subject", nil))
	})

	t.Run("Nil Event should panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("Expected panic but got none, empty Event is not allowed")
			}
		}()
		c := cone.New(nil)
		c.Serve(conetest.NewRecorder(), nil)
	})

	t.Run("Unregistred event should nak", func(t *testing.T) {
		c := cone.New(nil)
		r := conetest.NewRecorder()
		c.Serve(r, conetest.NewEvent("not.wanted", nil))
		if r.Result() != conetest.Nak {
			t.Errorf("Expected %s but got: %s", conetest.Nak, r.Result())
		}
	})

	t.Run("Registred event should ack", func(t *testing.T) {
		c := cone.New(nil)
		c.HandleFunc("is.wanted", func(_ cone.Response, _ cone.Event) {})
		r := conetest.NewRecorder()
		c.Serve(r, conetest.NewEvent("is.wanted", nil))
		if r.Result() != conetest.Ack {
			t.Errorf("Expected %s but got: %s", conetest.Nak, r.Result())
		}
	})
}
