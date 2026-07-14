package command

import (
	"context"
	"workout/training/domain"

	"github.com/pkg/errors"
)

type CancelTraining struct {
	TrainingUUID domain.TrainingUUID
	User         domain.User
}

func (h *Handler) CancelTraining(ctx context.Context, cmd CancelTraining) error {
	return h.trainingRepo.UpdateTraining(
		ctx,
		cmd.TrainingUUID,
		cmd.User,
		func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
			if err := tr.Cancel(); err != nil {
				return nil, err
			}

			if balanceDelta := domain.CancelBalanceDelta(*tr, cmd.User.Type()); balanceDelta != 0 {
				err := h.userSerivce.UpdateTrainingBalance(ctx, tr.UserID(), balanceDelta)
				if err != nil {
					return nil, errors.Wrap(err, "unable to change trainings balance")
				}
			}

			if err := h.trainerService.CancelTraining(ctx, tr.Time()); err != nil {
				return nil, errors.Wrap(err, "unable to cancel training")
			}

			return tr, nil
		},
	)
}
