package decorator

import (
	"context"

	"github.com/sirupsen/logrus"
)

// CommandHandler handles a command.
type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

// DecorateCommandHandler applies metrics and logging decorators to a command handler.
func ApplyDecorateCommandHandler[C any](handler CommandHandler[C], logger *logrus.Entry, metrics MetricsClient) CommandHandler[C] {
	loggingHandler := loggingCommandHandler[C]{
		logger: logger,
		next:   handler,
	}

	metricsHandler := metricsCommandHandler[C]{
		next: loggingHandler,
	}
	return metricsHandler
}
