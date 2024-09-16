package cone

type Handler interface {
	Serve(Response, *Event)
}

type HandlerFunc func(Response, *Event)

func (h HandlerFunc) Serve(r Response, e *Event) {
	h(r, e)
}

type Response interface {
	Ack() error
	Nak() error
}
