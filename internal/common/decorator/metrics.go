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

type commandMetricsDecorator[C any] struct {
	base   CommandHandler[C]
	client MetricsClient
}

func (commandMetricsDecorator commandMetricsDecorator[C]) Handle(ctx context.Context, cmd C) (err error) {
	start := time.Now()

	actionName := generateActionName(cmd)

	defer func() {
		end := time.Since(start)

		commandMetricsDecorator.client.Inc(
			fmt.Sprintf("commands.%s.duration", actionName),
			int(end.Seconds()),
		)

		if err == nil {
			commandMetricsDecorator.client.Inc(
				fmt.Sprintf("commands.%s.success", actionName),
				1,
			)
		} else {
			commandMetricsDecorator.client.Inc(
				fmt.Sprintf("commands.%s.failure", actionName),
				1,
			)
		}
	}()

	return commandMetricsDecorator.base.Handle(ctx, cmd)
}

type queryMetricsDecorator[Q any, R any] struct {
	base   QueryHandler[Q, R]
	client MetricsClient
}

func (queryMetricsDecorator queryMetricsDecorator[Q, R]) Handle(ctx context.Context, query Q) (result R, err error) {
	start := time.Now()

	actionName := strings.ToLower(generateActionName(query))

	defer func() {
		end := time.Since(start)

		queryMetricsDecorator.client.Inc(
			fmt.Sprintf("querys.%s.duration", actionName),
			int(end.Seconds()),
		)

		if err == nil {
			queryMetricsDecorator.client.Inc(
				fmt.Sprintf("querys.%s.success", actionName),
				1,
			)
		} else {
			queryMetricsDecorator.client.Inc(
				fmt.Sprintf("querys.%s.failure", actionName),
				1,
			)
		}
	}()

	return queryMetricsDecorator.base.Handle(ctx, query)
}
