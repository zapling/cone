package cone

type (
	Handler interface {
		Serve(Response, Event)
	}

	HandlerFunc func(Response, Event)

	Event interface {
		Subject() string
		Body() []byte
	}

	Response interface {
		Ack() error
		Nak() error
	}
)

func (h HandlerFunc) Serve(r Response, e Event) {
	h(r, e)
}
