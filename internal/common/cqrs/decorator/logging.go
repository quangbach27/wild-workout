package decorator

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func generateActionName(handler any) string {
	return strings.Split(fmt.Sprintf("%T", handler), ".")[1]
}

// loggingCommandHandler decorates a CommandHandler with logging.
type loggingCommandHandler[C any] struct {
	logger *logrus.Entry
	next   CommandHandler[C]
}

func (l loggingCommandHandler[C]) Handle(ctx context.Context, cmd C) (err error) {
	handlerType := generateActionName(cmd)
	logger := l.logger.WithFields(logrus.Fields{
		"command":      handlerType,
		"command_body": fmt.Sprintf("%#v", cmd),
	})
	logger.Infof("Handling command")
	defer func() {

		if err != nil {
			logger.Errorf("Command failed: %v", err)
		} else {
			logger.Infof("Command succeeded")
		}
	}()

	return l.next.Handle(ctx, cmd)
}

// loggingQueryHandler decorates a QueryHandler with logging.
type loggingQueryHandler[Q any, R any] struct {
	logger *logrus.Entry
	next   QueryHandler[Q, R]
}

func (l loggingQueryHandler[Q, R]) Handle(ctx context.Context, query Q) (result R, err error) {
	handlerType := generateActionName(query)
	logger := l.logger.WithFields(logrus.Fields{
		"query":      handlerType,
		"query_body": fmt.Sprintf("%#v", query),
	})
	logger.Infof("Handling query")
	defer func() {
		if err != nil {
			logger.Errorf("Query failed: %v", err)
		} else {
			logger.Infof("Query succeeded")
		}
	}()
	return l.next.Handle(ctx, query)
}
