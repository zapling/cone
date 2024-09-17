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
	c := cone.New(nil, nil)
	if c == nil {
		t.Fatalf("Expected Consumer got nil")
	}
}

func TestListenAndConsume(t *testing.T) {
	t.Run("Nil source should error", func(t *testing.T) {
		c := cone.New(nil, nil)
		err := c.ListenAndConsume()
		if err == nil {
			t.Fatal("Expected error but got nil")
		}
	})

	t.Run("Nil handler should error", func(t *testing.T) {
		s := conetest.NewSource()
		c := cone.New(s, nil)
		err := c.ListenAndConsume()
		if err == nil {
			t.Fatalf("Expected error but got nil")
		}
	})

	t.Run("Event should be served", func(t *testing.T) {
		e := conetest.NewEvent("event.subject", nil)
		s := conetest.NewSource()
		s.AddEvent(e)

		h := cone.NewHandlerMux()
		h.HandleFunc("event.subject", func(r cone.Response, _ *cone.Event) {
			_ = r.Ack()
		})

		c := cone.New(s, h)

		go func() {
			_ = c.ListenAndConsume()
		}()

		time.Sleep(5 * time.Millisecond)
		if s.NumAckd() != 1 {
			t.Error("Expected event to be acked, but it was not")
		}
	})

	t.Run("Calling func while still running should error", func(t *testing.T) {
		source := conetest.NewSource()
		var handler cone.HandlerFunc = func(r cone.Response, e *cone.Event) {}
		c := cone.New(source, handler)

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
}

func TestMiddlewareAroundConsumer(t *testing.T) {
	s := conetest.NewSource()
	h := cone.NewHandlerMux()
	h.HandleFunc("event.subject", func(r cone.Response, e *cone.Event) {
	})

	var beenInsideMiddleware bool
	middleware := func(next cone.Handler) cone.HandlerFunc {
		return func(r cone.Response, e *cone.Event) {
			beenInsideMiddleware = true
			next.Serve(r, e)
		}
	}

	c := cone.New(s, middleware(h))
	r := conetest.NewRecorder()
	c.Serve(r, conetest.NewEvent("event.subject", nil))

	if !beenInsideMiddleware {
		t.Fatal("Expected to hit the middleware, but did not")
	}
}

func TestShutdown(t *testing.T) {
	t.Run("Unstarted consumer should error", func(t *testing.T) {
		s := conetest.NewSource()
		h := cone.NewHandlerMux()
		c := cone.New(s, h)
		err := c.Shutdown(context.Background())
		if err == nil {
			t.Error("Expected error but got nil")
		}
	})

	t.Run("Should wait for all started handles", func(t *testing.T) {
		s := conetest.NewSource()
		s.AddEvent(conetest.NewEvent("event.subject", nil))
		h := cone.NewHandlerMux()
		h.HandleFunc("event.subject", func(r cone.Response, _ *cone.Event) {
			time.Sleep(1 * time.Second)
			_ = r.Ack()
		})

		c := cone.New(s, h)

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

		numAcked := s.NumAckd()
		if numAcked != 1 {
			t.Fatalf("Expected 1 acked message, but got: %d", numAcked)
		}
	})

	t.Run("Should stop eariler ListenAndConsume", func(t *testing.T) {
		s := conetest.NewSource()
		h := cone.NewHandlerMux()
		c := cone.New(s, h)

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
