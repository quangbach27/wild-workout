package query

import (
	"context"
	"workout/training/domain"
)

type ListTrainingsForUser struct {
	UserUUID domain.UserUUID
}

func (h *Handler) ListTrainingsForUser(ctx context.Context, query ListTrainingsForUser) ([]Training, error) {
	return h.trainingReadModel.FindTrainingsForUser(ctx, query.UserUUID)
}
