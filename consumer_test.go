package cone_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

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
		c.HandleFunc("is.wanted", func(_ cone.Response, _ *cone.Event) {})
		r := conetest.NewRecorder()
		c.Serve(r, conetest.NewEvent("is.wanted", nil))
		if r.Result() != conetest.Ack {
			t.Errorf("Expected %s but got: %s", conetest.Nak, r.Result())
		}
	})
}

func TestListenAndConsume(t *testing.T) {

	t.Run("Calling func while still running should error", func(t *testing.T) {
		source := conetest.NewSource()
		c := cone.New(source)

		go func() {
			_ = c.ListenAndConsume()
		}()

		time.Sleep(1 * time.Millisecond) // Give first gorutine time to start

		go func() {
			err := c.ListenAndConsume()
			if err == nil {
				panic("Expected error but got nil")
			}
		}()

		time.Sleep(5 * time.Millisecond)
	})

	t.Run("Nil source should error", func(t *testing.T) {
		c := cone.New(nil)
		err := c.ListenAndConsume()
		if err == nil {
			t.Fatal("Expected error but got nil")
		}
	})

	t.Run("Event should be served", func(t *testing.T) {
		event := conetest.NewEvent("event.subject", nil)
		source := conetest.NewSource()
		source.AddEvent(event)

		c := cone.New(source)
		c.HandleFunc("event.subject", func(r cone.Response, _ *cone.Event) {
			_ = r.Ack()
		})

		go func() {
			_ = c.ListenAndConsume()
		}()

		time.Sleep(5 * time.Millisecond)
		if source.NumAckd() != 1 {
			t.Error("Expected event to be acked, but it was not")
		}
	})
}

func TestShutdown(t *testing.T) {
	t.Run("Unstarted consumer should error", func(t *testing.T) {
		c := cone.New(nil)
		err := c.Shutdown(context.Background())
		if err == nil {
			t.Error("Expected error but got nil")
		}
	})

	t.Run("Should wait for all started handles", func(t *testing.T) {
		source := conetest.NewSource()
		source.AddEvent(conetest.NewEvent("event.subject", nil))
		c := cone.New(source)
		c.HandleFunc("event.subject", func(r cone.Response, _ *cone.Event) {
			time.Sleep(1 * time.Second)
			_ = r.Ack()
		})

		var consumerHasStopped atomic.Bool
		go func() {
			err := c.ListenAndConsume()
			if !errors.Is(err, cone.ErrConsumerStopped) {
				panic(fmt.Sprintf("Unexpected error: %s", err.Error()))
			}
			consumerHasStopped.Swap(true)
		}()

		time.Sleep(5 * time.Millisecond)

		err := c.Shutdown(context.Background())
		if err != nil {
			t.Fatalf("Expected nil but got err: %s", err.Error())
		}

		if !consumerHasStopped.Load() {
			t.Fatal("Consumer never stopped!")
		}

		numAcked := source.NumAckd()
		if numAcked != 1 {
			t.Fatalf("Expected 1 acked message, but got: %d", numAcked)
		}
	})

	t.Run("Should stop eariler ListenAndConsume", func(t *testing.T) {
		source := conetest.NewSource()
		c := cone.New(source)

		var consumerHasStopped atomic.Bool
		go func() {
			err := c.ListenAndConsume()
			if !errors.Is(err, cone.ErrConsumerStopped) {
				panic(fmt.Sprintf("Unexpected error: %s", err.Error()))
			}
			consumerHasStopped.Swap(true)
		}()

		time.Sleep(5 * time.Millisecond)
		err := c.Shutdown(context.Background())
		if err != nil {
			t.Fatalf("Expected nil but got err: %s", err.Error())
		}

		if !consumerHasStopped.Load() {
			t.Fatal("Consumer never stopped!")
		}
	})
}
