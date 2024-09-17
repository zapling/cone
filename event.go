package cone

func NewEvent(subject string, body []byte) (*Event, error) {
	return &Event{Subject: subject, Body: body, Header: make(Header)}, nil
}

type Event struct {
	Subject string
	Body    []byte
	Header  Header
}

type Header map[string][]string

func (h Header) Set(key, value string) {
	h[key] = []string{value}
}

func (h Header) Add(key, value string) {
	h[key] = append(h[key], value)
}

func (h Header) Get(key string) string {
	values, ok := h[key]
	if !ok {
		return ""
	}
	return values[0]
}

func (h Header) Values(key string) []string {
	values, ok := h[key]
	if !ok {
		return nil
	}
	return values
}
