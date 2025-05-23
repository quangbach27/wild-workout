package query

import (
	"context"
	"time"

	"github.com/quangbach27/wild-workout/internal/common/errors"
	"github.com/sirupsen/logrus"
)

type AvailableHoursReadModel interface {
	AvailableHours(ctx context.Context, from time.Time, to time.Time) ([]Date, error)
}

type AvailableHoursHandler struct {
	readModel AvailableHoursReadModel
}

func NewAvailableHourHandler(readModel AvailableHoursReadModel) AvailableHoursHandler {
	return AvailableHoursHandler{readModel: readModel}
}

type AvailableHours struct {
	From time.Time
	To   time.Time
}

func (handler AvailableHoursHandler) Handle(ctx context.Context, cmd AvailableHours) (d []Date, err error) {
	start := time.Now()
	defer func() {
		logrus.
			WithError(err).
			WithField("duration", time.Since(start)).
			Debug("AvailableHoursHandler executed")
	}()

	if cmd.From.After(cmd.To) {
		return nil, errors.NewIncorrectInputError("date-from-after-date-to", "Date from after date to")
	}

	return handler.readModel.AvailableHours(ctx, cmd.From, cmd.To)
}
