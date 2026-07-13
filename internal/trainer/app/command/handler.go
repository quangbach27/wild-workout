package command

import "workout/trainer/domain"

type Handler struct {
	hourRepo domain.HourRepository
}

func NewHandler(hourRepo domain.HourRepository) *Handler {
	if hourRepo == nil {
		panic("hourRepo can't be nil")
	}

	return &Handler{
		hourRepo: hourRepo,
	}
}
