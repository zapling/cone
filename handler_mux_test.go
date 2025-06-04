package cone_test

import (
	"testing"

	"github.com/zapling/cone"
	"github.com/zapling/cone/conetest"
)

func TestHandle(t *testing.T) {
	t.Run("Empty subject should panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("Expected panic, empty subject is not allowed")
			}
		}()

		c := cone.NewHandlerMux()
		c.Handle("", nil)
	})

	t.Run("Nil handler is ok", func(t *testing.T) {
		c := cone.NewHandlerMux()
		c.Handle("event.subject", nil)
	})

	t.Run("Same subject should override previous handler", func(t *testing.T) {
		c := cone.NewHandlerMux()
		var handlerCalled string
		var firstHandler cone.HandlerFunc = func(_ cone.Response, _ *cone.Event) { handlerCalled = "first" }
		var secondHandler cone.HandlerFunc = func(_ cone.Response, _ *cone.Event) { handlerCalled = "second" }
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

		c := cone.NewHandlerMux()
		c.HandleFunc("", nil)
	})

	t.Run("Nil handler is ok", func(t *testing.T) {
		c := cone.NewHandlerMux()
		c.HandleFunc("event.subject", nil)
	})

	t.Run("Same subject should override previous handler", func(t *testing.T) {
		c := cone.NewHandlerMux()
		var handlerCalled string
		var firstHandler cone.HandlerFunc = func(_ cone.Response, _ *cone.Event) { handlerCalled = "first" }
		var secondHandler cone.HandlerFunc = func(_ cone.Response, _ *cone.Event) { handlerCalled = "second" }
		c.HandleFunc("event.subject", firstHandler)
		c.HandleFunc("event.subject", secondHandler)
		r := conetest.NewRecorder()
		c.Serve(r, conetest.NewEvent("event.subject", nil))
		if handlerCalled != "second" {
			t.Errorf("Second handler should have been called, but '%s' was called", handlerCalled)
		}
	})
}

func TestHandlerMiddleware(t *testing.T) {
	middleware := func(next cone.Handler) cone.HandlerFunc {
		return func(r cone.Response, e *cone.Event) {
			if e.Subject != "valid" {
				return
			}

			next.Serve(r, e)
		}
	}

	var handler cone.HandlerFunc = func(r cone.Response, e *cone.Event) {
		_ = r.Ack()
	}

	t.Run("Should not pass middleware", func(t *testing.T) {
		r := conetest.NewRecorder()
		e := conetest.NewEvent("event.subject", nil)
		middleware(handler).Serve(r, e)
		if r.Result() != "" {
			t.Fatalf("Expected no result but got: %s", r.Result())
		}
	})

	t.Run("Should pass middleware", func(t *testing.T) {
		r := conetest.NewRecorder()
		e := conetest.NewEvent("valid", nil)
		middleware(handler).Serve(r, e)
		if r.Result() != conetest.Ack {
			t.Fatalf("Expected ack but got: %s", r.Result())
		}
	})
}

// func TestMiddlewareAroundConsumer(t *testing.T) {
// 	s := conetest.NewSource()
// 	h := cone.NewHandlerMux()
// 	h.HandleFunc("event.subject", func(r cone.Response, e *cone.Event) {
// 	})
//
// 	middleware := func(next cone.Handler) cone.HandlerFunc {
// 		return func(r cone.Response, e *cone.Event) {
// 			next.Serve(r, e)
// 		}
// 	}
//
// 	c := cone.New(s, middleware(h))
// 	c.Serve()
// }

func TestServe(t *testing.T) {
	t.Run("Nil Response should panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("Expected panic, empty Response is not allowed")
			}
		}()
		c := cone.NewHandlerMux()
		c.Serve(nil, conetest.NewEvent("event.subject", nil))
	})

	t.Run("Nil Event should panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Fatalf("Expected panic but got none, empty Event is not allowed")
			}
		}()
		c := cone.NewHandlerMux()
		c.Serve(conetest.NewRecorder(), nil)
	})

	t.Run("Unregistred event should nak", func(t *testing.T) {
		c := cone.NewHandlerMux()
		r := conetest.NewRecorder()
		c.Serve(r, conetest.NewEvent("not.wanted", nil))
		if r.Result() != conetest.Nak {
			t.Errorf("Expected %s but got: %s", conetest.Nak, r.Result())
		}
	})

	t.Run("Unregistred event should ack if AckUnknownSubjects=true is set", func(t *testing.T) {
		c := cone.NewHandlerMux()
		c.AckUnknownSubjects = true
		r := conetest.NewRecorder()
		c.Serve(r, conetest.NewEvent("not.wanted", nil))
		if r.Result() != conetest.Ack {
			t.Errorf("Expected %s but got: %s", conetest.Ack, r.Result())
		}
	})

	t.Run("Registred event should ack", func(t *testing.T) {
		c := cone.NewHandlerMux()
		c.HandleFunc("is.wanted", func(_ cone.Response, _ *cone.Event) {})
		r := conetest.NewRecorder()
		c.Serve(r, conetest.NewEvent("is.wanted", nil))
		if r.Result() != conetest.Ack {
			t.Errorf("Expected %s but got: %s", conetest.Nak, r.Result())
		}
	})
}
