package query

import (
	"context"
	"time"
)

type IsHourAvailable struct {
	Hour time.Time
}

func (h *Handler) IsHourAvailable(ctx context.Context, query IsHourAvailable) (bool, error) {
	hour, err := h.hourRepo.GetHour(ctx, query.Hour)
	if err != nil {
		return false, err
	}

	return hour.IsAvailable(), nil
}
