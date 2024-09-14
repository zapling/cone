package cone

import "context"

type (
	Handler interface {
		Serve(Response, Event)
	}

	HandlerFunc func(Response, Event)

	Response interface {
		Ack() error
		Nak() error
	}

	Event interface {
		Subject() string
		Body() []byte
	}

	Source interface {
		Start() error
		Stop(ctx context.Context) error
		GetNextEvent() (ResponseAndEvent, error)
	}

	ResponseAndEvent interface {
		Response
		Event
	}
)

func (h HandlerFunc) Serve(r Response, e Event) {
	h(r, e)
}
