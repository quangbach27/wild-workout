package http

import (
	"context"
	"workout/common"
	"workout/trainer/app/command"
	"workout/trainer/app/query"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Handler struct {
	commandHandler *command.Handler
	queryHandler   *query.Handler
}

func NewHandler(
	commandHandler *command.Handler,
	queryHandler *query.Handler,
) Handler {
	if commandHandler == nil {
		panic("commandHandler can't be nil")
	}
	if queryHandler == nil {
		panic("queryHandler can't be nil")
	}
	return Handler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

// (GET /trainer/calendar)
func (h Handler) GetTrainerAvailableHours(ctx context.Context, request GetTrainerAvailableHoursRequestObject) (GetTrainerAvailableHoursResponseObject, error) {
	dates, err := h.queryHandler.ListAvailableHours(ctx, query.ListAvailableHours{
		From: request.Params.DateFrom,
		To:   request.Params.DateTo,
	})
	if err != nil {
		return nil, err
	}

	response := make(GetTrainerAvailableHours200JSONResponse, 0, len(dates))
	for _, d := range dates {
		response = append(response, toAPIDate(d))
	}

	return response, nil
}

// (PUT /trainer/calendar/make-hour-available)
func (h Handler) MakeHourAvailable(ctx context.Context, request MakeHourAvailableRequestObject) (MakeHourAvailableResponseObject, error) {
	err := h.commandHandler.MakeHourAvailable(ctx, command.MakeHourAvailable{
		Hours: request.Body.Hours,
	})
	if err != nil {
		return nil, err
	}

	return MakeHourAvailable204JSONResponse{}, nil
}

// (PUT /trainer/calendar/make-hour-unavailable)
func (h Handler) MakeHourUnavailable(ctx context.Context, request MakeHourUnavailableRequestObject) (MakeHourUnavailableResponseObject, error) {
	err := h.commandHandler.MakeHourUnavailable(ctx, command.MakeHourUnavailable{
		Hours: request.Body.Hours,
	})
	if err != nil {
		return nil, err
	}

	return MakeHourUnavailable204JSONResponse{}, nil
}

func toAPIDate(d query.Date) Date {
	hours := make([]Hour, 0, len(d.Hours))
	for _, h := range d.Hours {
		hours = append(hours, Hour{
			Hour:                 h.Hour,
			Available:            h.Available,
			HasTrainingScheduled: h.HasTrainingScheduled,
		})
	}

	return Date{
		Date:         openapi_types.Date{Time: d.Date},
		HasFreeHours: d.HasFreeHours,
		Hours:        hours,
	}
}

func Register(e common.EchoRouter, handler Handler) error {
	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}
