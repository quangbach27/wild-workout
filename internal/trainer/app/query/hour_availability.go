package query

import (
	"context"
	"time"

	"github.com/quangbach27/wild-workout/internal/trainer/domain/hour"
)

type HourAvailabilityHandler struct {
	hourRepo hour.Repository
}

func NewHourAvailabilityHandler(hourRepo hour.Repository) HourAvailabilityHandler {
	if hourRepo == nil {
		panic("nil hourRepo")
	}

	return HourAvailabilityHandler{hourRepo: hourRepo}
}

func (handler HourAvailabilityHandler) Handle(ctx context.Context, h time.Time) (bool, error) {
	hour, err := handler.hourRepo.GetHour(ctx, h)
	if err != nil {
		return false, err
	}

	return hour.IsAvailable(), nil
}
