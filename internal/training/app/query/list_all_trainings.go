package query

import "context"

type ListAllTrainings struct {
}

func (h *Handler) ListAllTrainings(ctx context.Context, query ListAllTrainings) ([]Training, error) {
	return h.trainingReadModel.ListAllTrainings(ctx)
}
