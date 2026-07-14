package query

import (
	"context"
	"workout/training/domain"
)

type TrainingReadModel interface {
	FindTrainingsForUser(ctx context.Context, userUUID domain.UserUUID) ([]Training, error)
	ListAllTrainings(ctx context.Context) ([]Training, error)
}

type Handler struct {
	trainingReadModel TrainingReadModel
}

func NewHandler(trainingReadModel TrainingReadModel) *Handler {
	if trainingReadModel == nil {
		panic("trainingReadModel can't be nil")
	}

	return &Handler{
		trainingReadModel: trainingReadModel,
	}
}
