package decorator

import (
	"context"

	"github.com/sirupsen/logrus"
)

// QueryHandler handles a query.
type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

func ApplyDecorateQueryHandler[Q any, R any](baseHandler QueryHandler[Q, R], logger *logrus.Entry, metricsClient MetricsClient) QueryHandler[Q, R] {
	loggingHandler := loggingQueryHandler[Q, R]{
		next:   baseHandler,
		logger: logger,
	}

	metricsHandler := metricsQueryHandler[Q, R]{
		next: loggingHandler,
	}
	return metricsHandler
}
