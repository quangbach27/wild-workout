package command

import (
	"context"
	"time"
	"workout/common"
	"workout/trainer/domain"
)

type MakeHourUnavailable struct {
	Hours []time.Time
}

func (h *Handler) MakeHourUnavailable(ctx context.Context, cmd MakeHourUnavailable) error {
	for _, hourTime := range cmd.Hours {
		err := h.hourRepo.UpdateHour(
			ctx,
			hourTime,
			func(h *domain.Hour) (*domain.Hour, error) {
				err := h.MakeNotAvailable()
				if err != nil {
					return nil, common.NewInvalidInputError(
						"unable-to-update-unavailable",
						"%s", err.Error(),
					).WithInternalError(err)
				}

				return h, nil
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
