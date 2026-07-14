package command

import (
	"context"
	"time"
	"workout/trainer/domain"
)

type ScheduleTraining struct {
	Hour time.Time
}

func (h *Handler) ScheduleTraining(ctx context.Context, cmd ScheduleTraining) error {
	return h.hourRepo.UpdateHour(ctx, cmd.Hour, func(h *domain.Hour) (*domain.Hour, error) {
		if err := h.ScheduleTraining(); err != nil {
			return nil, err
		}
		return h, nil
	})
}
