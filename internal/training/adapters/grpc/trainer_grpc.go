package grpc

import (
	"context"
	"fmt"
	"time"
	"workout/common/grpc/protobuf/trainer"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type TrainerGrpc struct {
	client trainer.TrainerServiceClient
}

func NewTrainerGrpc(client trainer.TrainerServiceClient) *TrainerGrpc {
	if client == nil {
		panic("trainerServiceClient can't be nil")
	}

	return &TrainerGrpc{
		client: client,
	}
}

func (g *TrainerGrpc) ScheduleTraining(ctx context.Context, trainingTime time.Time) error {
	_, err := g.client.ScheduleTraining(ctx, &trainer.UpdateHourRequest{
		Time: timestamppb.New(trainingTime),
	})
	return err
}

func (g *TrainerGrpc) CancelTraining(ctx context.Context, trainingTime time.Time) error {
	_, err := g.client.CancelTraining(ctx, &trainer.UpdateHourRequest{
		Time: timestamppb.New(trainingTime),
	})
	return err
}

func (g *TrainerGrpc) MoveTraining(
	ctx context.Context,
	newTime time.Time,
	originalTrainingTime time.Time,
) error {
	err := g.ScheduleTraining(ctx, newTime)
	if err != nil {
		return fmt.Errorf("unable to schedule training: %w", err)
	}

	err = g.CancelTraining(ctx, originalTrainingTime)
	if err != nil {
		return fmt.Errorf("unable to cancel training: %w", err)
	}

	return nil
}
