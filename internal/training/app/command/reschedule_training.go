package command

import (
	"context"
	"time"
	"workout/training/domain"
)

type RescheduleTraining struct {
	TrainingUUID domain.TrainingUUID
	NewTime      time.Time

	User domain.User

	NewNotes string
}

func (h *Handler) RescheduleTraining(ctx context.Context, cmd RescheduleTraining) error {
	var originalTrainingTime time.Time

	err := h.trainingRepo.UpdateTraining(
		ctx,
		cmd.TrainingUUID,
		cmd.User,
		func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
			originalTrainingTime = tr.Time()

			if err := tr.UpdateNotes(cmd.NewNotes); err != nil {
				return nil, err
			}

			if err := tr.RescheduleTraining(cmd.NewTime); err != nil {
				return nil, err
			}

			return tr, nil
		},
	)
	if err != nil {
		return err
	}

	err = h.trainerService.MoveTraining(ctx, cmd.NewTime, originalTrainingTime)
	if err != nil {
		return err
	}

	return nil
}
