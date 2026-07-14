package query

import (
	"context"
	"time"
	"workout/trainer/domain"
)

type AvailableHoursReadModel interface {
	ListAvailableHours(ctx context.Context, from time.Time, to time.Time) ([]Date, error)
}

type Handler struct {
	hourReadModel AvailableHoursReadModel
	hourRepo      domain.HourRepository
}

func NewHandler(
	hourReadModel AvailableHoursReadModel,
	hourRepo domain.HourRepository,
) *Handler {
	if hourReadModel == nil {
		panic("hourReadModel can't be nil")
	}
	if hourRepo == nil {
		panic("hourRepo can't be nil")
	}

	return &Handler{
		hourReadModel: hourReadModel,
		hourRepo:      hourRepo,
	}
}
