package command

import (
	"context"
	"time"

	"github.com/quangbach27/wild-workout/internal/common/errors"
	"github.com/quangbach27/wild-workout/internal/trainer/domain/hour"
)

type CancelTrainingHandler struct {
	hourRepo hour.Repository
}

func NewCancelTrainingHandler(hourRepo hour.Repository) CancelTrainingHandler {
	if hourRepo == nil {
		panic("nil hourRepo")
	}

	return CancelTrainingHandler{hourRepo: hourRepo}
}

func (handler CancelTrainingHandler) Handle(ctx context.Context, hourToCancel time.Time) error {
	err := handler.hourRepo.UpdateHour(ctx, hourToCancel, func(h *hour.Hour) (*hour.Hour, error) {
		if err := h.CancelTraining(); err != nil {
			return nil, err
		}

		return h, nil
	})

	if err != nil {
		return errors.NewSlugError(err.Error(), "unable-to-update-availability")
	}

	return nil
}
