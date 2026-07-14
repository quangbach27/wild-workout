package command

import (
	"context"
	"time"
	"workout/training/domain"
)

type UserService interface {
	UpdateTrainingBalance(ctx context.Context, userUUID domain.UserUUID, amountChange int) error
}

type TrainerService interface {
	ScheduleTraining(ctx context.Context, trainingTime time.Time) error
	CancelTraining(ctx context.Context, trainingTime time.Time) error
	MoveTraining(
		ctx context.Context,
		newTime time.Time,
		originalTrainingTime time.Time,
	) error
}

type Handler struct {
	trainingRepo   domain.TrainingRepository
	userSerivce    UserService
	trainerService TrainerService
}

func NewHandler(
	trainingRepo domain.TrainingRepository,
	userService UserService,
	trainerService TrainerService,
) *Handler {
	if trainingRepo == nil {
		panic("trainingRepo can't be nil")
	}

	if userService == nil {
		panic("userService can't be nil")
	}

	if trainerService == nil {
		panic("trainerService can't be nil")
	}

	return &Handler{
		trainingRepo:   trainingRepo,
		userSerivce:    userService,
		trainerService: trainerService,
	}
}
