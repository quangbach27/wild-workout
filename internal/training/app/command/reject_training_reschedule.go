package command

import (
	"context"
	"workout/training/domain"
)

type RejectTrainingReschedule struct {
	TrainingUUID domain.TrainingUUID
	User         domain.User
}

func (h *Handler) RejectTrainingReschedule(ctx context.Context, cmd RejectTrainingReschedule) error {
	return h.trainingRepo.UpdateTraining(
		ctx,
		cmd.TrainingUUID,
		cmd.User,
		func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
			if err := tr.RejectReschedule(); err != nil {
				return nil, err
			}

			return tr, nil
		},
	)
}
