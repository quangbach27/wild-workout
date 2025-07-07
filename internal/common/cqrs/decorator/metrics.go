package decorator

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type MetricsClient interface {
	Inc(key string, value int)
}

// MetricsCommandHandler decorates a CommandHandler with metrics and logging, then runs the command.
type MetricsCommandHandler[C any] struct {
	next   CommandHandler[C]
	client MetricsClient
}

func (m *MetricsCommandHandler[C]) Handle(ctx context.Context, cmd C) (err error) {
	start := time.Now()

	actionName := strings.ToLower(generateActionName(cmd))

	defer func() {
		end := time.Since(start)

		m.client.Inc(fmt.Sprintf("commands.%s.duration", actionName), int(end.Seconds()))

		if err == nil {
			m.client.Inc(fmt.Sprintf("commands.%s.success", actionName), 1)
		} else {
			m.client.Inc(fmt.Sprintf("commands.%s.failure", actionName), 1)
		}
	}()

	return m.next.Handle(ctx, cmd)
}

// MetricsQueryHandler decorates a QueryHandler with metrics and logging, then runs the query.
type MetricsQueryHandler[Q any, R any] struct {
	next   QueryHandler[Q, R]
	client MetricsClient
}

func (m *MetricsQueryHandler[Q, R]) Handle(ctx context.Context, query Q) (result R, err error) {
	start := time.Now()
	actionName := strings.ToLower(generateActionName(query))

	defer func() {
		end := time.Since(start)

		m.client.Inc(fmt.Sprintf("query.%s.duration", actionName), int(end.Seconds()))

		if err == nil {
			m.client.Inc(fmt.Sprintf("query.%s.success", actionName), 1)
		} else {
			m.client.Inc(fmt.Sprintf("query.%s.failure", actionName), 1)
		}
	}()

	return m.next.Handle(ctx, query)
}
