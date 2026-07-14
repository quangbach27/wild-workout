package grpc

import (
	"context"
	"time"
	"workout/common/grpc/protobuf/trainer"
	"workout/trainer/app/command"
	"workout/trainer/app/query"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	trainer.UnimplementedTrainerServiceServer

	commandHandler *command.Handler
	queryHandler   *query.Handler
}

func NewHandler(
	commandHandler *command.Handler,
	queryHandler *query.Handler,
) *Handler {
	if commandHandler == nil {
		panic("commandHandler can't be nil")
	}

	if queryHandler == nil {
		panic("queryHandler can't be nil")
	}

	return &Handler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

func (h *Handler) IsHourAvailable(ctx context.Context, req *trainer.IsHourAvailableRequest) (*trainer.IsHourAvailableResponse, error) {
	available, err := h.queryHandler.IsHourAvailable(ctx, query.IsHourAvailable{Hour: req.GetTime().AsTime()})
	if err != nil {
		return nil, err
	}

	return &trainer.IsHourAvailableResponse{IsAvailable: available}, nil
}

func (h *Handler) ScheduleTraining(ctx context.Context, req *trainer.UpdateHourRequest) (*emptypb.Empty, error) {
	if err := h.commandHandler.ScheduleTraining(ctx, command.ScheduleTraining{
		Hour: req.GetTime().AsTime()}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) CancelTraining(ctx context.Context, req *trainer.UpdateHourRequest) (*emptypb.Empty, error) {
	if err := h.commandHandler.CancelTraining(ctx, command.CancelTraining{
		Hour: req.GetTime().AsTime(),
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) MakeHourAvailable(ctx context.Context, req *trainer.UpdateHourRequest) (*emptypb.Empty, error) {
	if err := h.commandHandler.MakeHourAvailable(ctx, command.MakeHourAvailable{
		Hours: []time.Time{req.GetTime().AsTime()},
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
