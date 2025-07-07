package decorator

import "context"

// QueryHandler handles a query.
type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}
