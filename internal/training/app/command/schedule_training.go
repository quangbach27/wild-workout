package command

import (
	"context"
	"fmt"
	"time"
	"workout/training/domain"
)

type ScheduleTraining struct {
	TrainingUUID domain.TrainingUUID

	UserUUID domain.UserUUID
	UserName string

	TrainingTime time.Time
	Notes        string
}

func (h *Handler) ScheduleTraining(ctx context.Context, cmd ScheduleTraining) error {
	training, err := domain.NewTraining(cmd.UserUUID, cmd.UserName, cmd.TrainingTime)
	if err != nil {
		return err
	}

	err = h.trainingRepo.AddTraining(ctx, training)
	if err != nil {
		return fmt.Errorf("failed to save training to db: %w", err)
	}

	err = h.userSerivce.UpdateTrainingBalance(ctx, training.UserUUID(), -1)
	if err != nil {
		return fmt.Errorf("failed to update training balance: %w", err)
	}

	err = h.trainerService.ScheduleTraining(ctx, training.Time())
	if err != nil {
		return fmt.Errorf("failed to schedule training: %w", err)
	}

	return nil

}
