package command

import (
	"context"
	"time"
	"workout/training/domain"
)

type RequestTrainingReschedule struct {
	TrainingUUID domain.TrainingUUID
	NewTime      time.Time

	User domain.User

	NewNotes string
}

func (h *Handler) RequestTrainingReschedule(ctx context.Context, cmd RequestTrainingReschedule) error {
	return h.trainingRepo.UpdateTraining(
		ctx,
		cmd.TrainingUUID,
		cmd.User,
		func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
			if err := tr.UpdateNotes(cmd.NewNotes); err != nil {
				return nil, err
			}

			tr.ProposeReschedule(cmd.NewTime, cmd.User.Type())

			return tr, nil
		},
	)
}
