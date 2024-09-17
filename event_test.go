package cone_test

import (
	"testing"

	"github.com/zapling/cone"
	"github.com/zapling/cone/conetest"
)

func TestNewEvent(t *testing.T) {
	_, err := cone.NewEvent("event.subject", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
}

func TestHeader(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		event := conetest.NewEvent("event.subject", nil)
		event.Header.Set("some-key", "some-value")
		if len(event.Header) != 1 {
			t.Fatalf("Expected header length to be 1, got: %d", len(event.Header))
		}
	})

	t.Run("Get unset", func(t *testing.T) {
		event := conetest.NewEvent("event.subject", nil)
		if event.Header.Get("some-key") != "" {
			t.Fatalf("Expected empty string but got '%s'", event.Header.Get("some-key"))
		}
	})

	t.Run("Get", func(t *testing.T) {
		event := conetest.NewEvent("event.subject", nil)
		event.Header.Set("some-key", "some-value")
		if event.Header.Get("some-key") != "some-value" {
			t.Fatalf("Expected 'some-value' got '%s'", event.Header.Get("some-key"))
		}
	})

	t.Run("Add", func(t *testing.T) {
		event := conetest.NewEvent("event.subject", nil)
		event.Header.Add("some-key", "value1")
		event.Header.Add("some-key", "value2")
		if len(event.Header) != 1 {
			t.Fatalf("Expected 1 header, got %d", len(event.Header))
		}
		if len(event.Header["some-key"]) != 2 {
			t.Fatalf("Exepcted 2 values, got %d", len(event.Header["some-key"]))
		}
	})

	t.Run("Values on unset key", func(t *testing.T) {
		event := conetest.NewEvent("event.subject", nil)
		values := event.Header.Values("some-key")
		if len(values) != 0 {
			t.Fatalf("Exepcted 0 values but got %d", len(values))
		}
	})

	t.Run("Values", func(t *testing.T) {
		event := conetest.NewEvent("event.subject", nil)
		event.Header.Add("some-key", "value1")
		event.Header.Add("some-key", "value2")
		values := event.Header.Values("some-key")
		if len(values) != 2 {
			t.Fatalf("Expected 2 values but got %d", len(values))
		}

		if values[0] != "value1" {
			t.Fatalf("Expected value 0 to be 'value1' but got '%s'", values[0])
		}

		if values[1] != "value2" {
			t.Fatalf("Expected value 0 to be 'value2' but got '%s'", values[1])
		}
	})
}
