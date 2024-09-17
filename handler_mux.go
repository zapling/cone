package cone

import "fmt"

func NewHandlerMux() *HandlerMux {
	return &HandlerMux{handlers: make(map[string]Handler)}
}

type HandlerMux struct {
	handlers map[string]Handler
}

func (h *HandlerMux) Handle(subject string, handler Handler) {
	err := h.register(subject, handler)
	if err != nil {
		panic(err)
	}
}

func (h *HandlerMux) HandleFunc(subject string, handlerFunc HandlerFunc) {
	err := h.register(subject, handlerFunc)
	if err != nil {
		panic(err)
	}
}

func (h *HandlerMux) register(subject string, handler Handler) error {
	if subject == "" {
		return fmt.Errorf("empty subject is not allowed")
	}
	h.handlers[subject] = handler
	return nil
}

func (h *HandlerMux) Serve(r Response, e *Event) {
	if e == nil {
		panic("event is nil")
	}

	if err := h.serveEvent(r, e); err != nil {
		panic(err)
	}
}

func (h *HandlerMux) serveEvent(r Response, e *Event) error {
	handler, ok := h.handlers[e.Subject]
	if !ok {
		return r.Nak()
	}

	handler.Serve(r, e)
	return r.Ack()
}
