package query

import (
	"context"
	"time"
	"workout/common"
)

type ListAvailableHours struct {
	From time.Time
	To   time.Time
}

type Date struct {
	Date         time.Time
	HasFreeHours bool
	Hours        []Hour
}

type Hour struct {
	Available            bool
	HasTrainingScheduled bool
	Hour                 time.Time
}

func (h *Handler) ListAvailableHours(ctx context.Context, query ListAvailableHours) ([]Date, error) {
	if query.From.After(query.To) {
		return nil, common.NewInvalidInputError("date-from-after-date-to", "Date from after date to")
	}

	return h.hourReadModel.ListAvailableHours(ctx, query.From, query.To)
}
