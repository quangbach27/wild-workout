package http

import (
	"context"
	"fmt"

	"workout/common"
	"workout/training/app/command"
	"workout/training/app/query"
	"workout/training/domain"
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
		panic("commandHandler can't be empty")
	}
	if queryHandler == nil {
		panic("queryHandler can't be empty")
	}
	return Handler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}
func Register(e common.EchoRouter, handler Handler) error {
	RegisterHandlers(e, NewStrictHandler(handler, nil))
	return nil
}

// (GET /trainings)
func (h Handler) GetTrainings(ctx context.Context, request GetTrainingsRequestObject) (GetTrainingsResponseObject, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var trainings []query.Training
	if user.Type() == domain.Trainer {
		trainings, err = h.queryHandler.ListAllTrainings(ctx, query.ListAllTrainings{})
	} else {
		trainings, err = h.queryHandler.ListTrainingsForUser(ctx, query.ListTrainingsForUser{
			UserUUID: user.UUID(),
		})
	}
	if err != nil {
		return nil, err
	}

	apiTrainings := make([]Training, 0, len(trainings))
	for _, tr := range trainings {
		apiTraining, err := toAPITraining(tr)
		if err != nil {
			return nil, err
		}
		apiTrainings = append(apiTrainings, apiTraining)
	}

	return GetTrainings200JSONResponse{Trainings: apiTrainings}, nil
}

// (POST /trainings)
func (h Handler) CreateTraining(ctx context.Context, request CreateTrainingRequestObject) (CreateTrainingResponseObject, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = h.commandHandler.ScheduleTraining(ctx, command.ScheduleTraining{
		UserUUID:     user.UUID(),
		UserName:     user.Name,
		TrainingTime: request.Body.Time,
		Notes:        request.Body.Notes,
	})
	if err != nil {
		return nil, err
	}

	return CreateTraining204Response{}, nil
}

// (DELETE /trainings/{trainingUUID})
func (h Handler) CancelTraining(ctx context.Context, request CancelTrainingRequestObject) (CancelTrainingResponseObject, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = h.commandHandler.CancelTraining(ctx, command.CancelTraining{
		TrainingUUID: request.TrainingUUID,
		User:         user.User,
	})
	if err != nil {
		return nil, err
	}

	return CancelTraining204Response{}, nil
}

// (PUT /trainings/{trainingUUID}/approve-reschedule)
func (h Handler) ApproveRescheduleTraining(ctx context.Context, request ApproveRescheduleTrainingRequestObject) (ApproveRescheduleTrainingResponseObject, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = h.commandHandler.ApproveTrainingReschedule(ctx, command.ApproveTrainingReschedule{
		TrainingUUID: request.TrainingUUID,
		User:         user.User,
	})
	if err != nil {
		return nil, err
	}

	return ApproveRescheduleTraining204Response{}, nil
}

// (PUT /trainings/{trainingUUID}/reject-reschedule)
func (h Handler) RejectRescheduleTraining(ctx context.Context, request RejectRescheduleTrainingRequestObject) (RejectRescheduleTrainingResponseObject, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = h.commandHandler.RejectTrainingReschedule(ctx, command.RejectTrainingReschedule{
		TrainingUUID: request.TrainingUUID,
		User:         user.User,
	})
	if err != nil {
		return nil, err
	}

	return RejectRescheduleTraining204Response{}, nil
}

// (PUT /trainings/{trainingUUID}/request-reschedule)
func (h Handler) RequestRescheduleTraining(ctx context.Context, request RequestRescheduleTrainingRequestObject) (RequestRescheduleTrainingResponseObject, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = h.commandHandler.RequestTrainingReschedule(ctx, command.RequestTrainingReschedule{
		TrainingUUID: request.TrainingUUID,
		NewTime:      request.Body.Time,
		User:         user.User,
		NewNotes:     request.Body.Notes,
	})
	if err != nil {
		return nil, err
	}

	return RequestRescheduleTraining204Response{}, nil
}

// (PUT /trainings/{trainingUUID}/reschedule)
func (h Handler) RescheduleTraining(ctx context.Context, request RescheduleTrainingRequestObject) (RescheduleTrainingResponseObject, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = h.commandHandler.RescheduleTraining(ctx, command.RescheduleTraining{
		TrainingUUID: request.TrainingUUID,
		NewTime:      request.Body.Time,
		User:         user.User,
		NewNotes:     request.Body.Notes,
	})
	if err != nil {
		return nil, err
	}

	return RescheduleTraining204Response{}, nil
}

// toAPITraining converts a query.Training into its openapi representation.
// MoveRequiresAccept is true whenever a reschedule is pending, regardless of
// who proposed it — the frontend can compare MoveProposedBy against the
// current user to decide whether to show "accept" or "waiting" UI.
func toAPITraining(tr query.Training) (Training, error) {
	apiTraining := Training{
		Uuid:           tr.UUID,
		UserUuid:       tr.UserUUID,
		User:           tr.User,
		Notes:          tr.Notes,
		Time:           tr.Time,
		CanBeCancelled: tr.CanBeCancelled,
		ProposedTime:   tr.ProposedTime,
	}

	if tr.MoveProposedBy != nil {
		var moveProposedBy domain.UserType
		if err := moveProposedBy.UnmarshalText([]byte(*tr.MoveProposedBy)); err != nil {
			return Training{}, fmt.Errorf("invalid moveProposedBy value %q: %w", *tr.MoveProposedBy, err)
		}
		apiTraining.MoveProposedBy = &moveProposedBy
		apiTraining.MoveRequiresAccept = true
	}

	return apiTraining, nil
}
