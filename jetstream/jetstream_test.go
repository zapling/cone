package jetstream_test

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	conejetstream "github.com/zapling/cone/jetstream"
)

func TestNew(t *testing.T) {
	nc := getNatsConn(t)
	defer nc.Drain()
	consumer := getNatsConsumer(t, nc)
	source := conejetstream.New(consumer)
	if source == nil {
		t.Fatalf("Source should not be nil")
	}
}

func TestStartAndStop(t *testing.T) {
	nc := getNatsConn(t)
	defer nc.Drain()
	consumer := getNatsConsumer(t, nc)
	source := conejetstream.New(consumer)

	t.Run("Start", func(t *testing.T) {
		err := source.Start()
		if err != nil {
			t.Fatalf("Failed to start consumer: %s", err.Error())
		}
	})

	t.Run("Stop", func(t *testing.T) {
		err := source.Stop(context.Background())
		if err != nil {
			t.Fatalf("Failed to stop consumer: %s", err.Error())
		}
	})
}

func TestGetNextEvent(t *testing.T) {
	nc := getNatsConn(t)
	defer nc.Drain()
	js, err := jetstream.New(nc)
	if err != nil {
		t.Fatalf("Failed to get jetstream instance: %s", err.Error())
	}
	consumer := getNatsConsumer(t, nc)
	source := conejetstream.New(consumer)
	err = source.Start()
	if err != nil {
		t.Fatalf("Failed to start consumer: %s", err.Error())
	}
	defer source.Stop(context.Background())

	t.Run("No event", func(t *testing.T) {
		response, event, err := source.Next()
		if err != nil {
			t.Fatalf("Failed to get next event: %s", err.Error())
		}

		if response != nil {
			t.Fatal("Expected nil response but got response")
		}

		if event != nil {
			t.Fatal("Expected nil event but got event")
		}
	})

	t.Run("Event", func(t *testing.T) {
		pubAck, err := js.PublishMsg(context.Background(), &nats.Msg{
			Subject: "test_event",
		})
		if err != nil {
			t.Fatalf("Failed to publish msg: %s", err.Error())
		}

		t.Logf("Published msg to stream: %s", pubAck.Stream)

		response, event, err := source.Next()
		if err != nil {
			t.Fatalf("Failed to get next event: %s", err.Error())
		}

		if response == nil {
			t.Fatal("Expected response but got nil")
		}

		if event == nil {
			t.Fatal("Expected event but got nil")
		}

		if event.Subject != "test_event" {
			t.Fatalf("Got unexpected event: %s", event.Subject)
		}
	})

	t.Run("Event subject", func(t *testing.T) {

	})
}

func getNatsConn(t *testing.T) *nats.Conn {
	t.Helper()
	nc, err := nats.Connect("localhost:4222")
	if err != nil {
		t.Fatalf("Failed to connect to nats: %s", err.Error())
	}
	return nc
}

func getNatsConsumer(t *testing.T, nc *nats.Conn) jetstream.Consumer {
	t.Helper()

	js, err := jetstream.New(nc)
	if err != nil {
		t.Fatalf("Failed get jetstream instance: %s", err.Error())
	}

	stream, err := js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
		Name:     "jetstream-test",
		Subjects: []string{"*"},
	})
	if err != nil {
		t.Fatalf("Failed to create or update stream: %s", err.Error())
	}

	err = stream.Purge(context.Background())
	if err != nil {
		t.Fatalf("Failed to purge stream: %s", err.Error())
	}

	jetstreamConsumer, err := js.CreateOrUpdateConsumer(
		context.Background(),
		"jetstream-test",
		jetstream.ConsumerConfig{
			Name: "jetstream-consumer",
		},
	)
	if err != nil {
		t.Fatalf("Failed to create or update consumer: %s", err.Error())
	}

	return jetstreamConsumer
}
