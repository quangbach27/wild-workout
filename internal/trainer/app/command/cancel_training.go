package command

import (
	"context"
	"time"
	"workout/common"
	"workout/trainer/domain"
)

type CancelTraining struct {
	Hour time.Time
}

func (h *Handler) CancelTraining(ctx context.Context, cmd CancelTraining) error {
	return h.hourRepo.UpdateHour(ctx, cmd.Hour, func(h *domain.Hour) (*domain.Hour, error) {
		if err := h.CancelTraining(); err != nil {
			return nil, common.NewInvalidInputError("cancel-training-failed", "%s", err.Error())
		}
		return h, nil
	})
}
