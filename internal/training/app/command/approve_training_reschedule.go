package command

import (
	"context"
	"workout/training/domain"
)

type ApproveTrainingReschedule struct {
	TrainingUUID domain.TrainingUUID
	User         domain.User
}

func (h *Handler) ApproveTrainingReschedule(ctx context.Context, cmd ApproveTrainingReschedule) error {
	return h.trainingRepo.UpdateTraining(
		ctx,
		cmd.TrainingUUID,
		cmd.User,
		func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
			originalTrainingTime := tr.Time()

			if err := tr.ApproveReschedule(cmd.User.Type()); err != nil {
				return nil, err
			}

			err := h.trainerService.MoveTraining(ctx, tr.Time(), originalTrainingTime)
			if err != nil {
				return nil, err
			}

			return tr, nil
		},
	)
}
