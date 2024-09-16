package cone

func NewEvent(subject string, body []byte) (*Event, error) {
	return &Event{Subject: subject, Body: body}, nil
}

type Event struct {
	Subject string
	Body    []byte
}
