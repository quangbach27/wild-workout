package query

import (
	"context"
	"time"
)

type AvailableHoursReadModel interface {
	ListAvailableHours(ctx context.Context, from time.Time, to time.Time) ([]Date, error)
}

type Handler struct {
	hourReadModel AvailableHoursReadModel
}

func NewHandler(hourReadModel AvailableHoursReadModel) *Handler {
	if hourReadModel == nil {
		panic("hourReadModel can't be nil")
	}

	return &Handler{
		hourReadModel: hourReadModel,
	}
}
