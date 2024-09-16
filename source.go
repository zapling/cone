package cone

import "context"

type Source interface {
	Start() error
	Stop(ctx context.Context) error
	Next() (Response, *Event, error)
}
