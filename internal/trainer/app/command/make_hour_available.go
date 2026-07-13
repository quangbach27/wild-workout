package command

import (
	"context"
	"time"
	"workout/common"
	"workout/trainer/domain"
)

type MakeHourAvailable struct {
	Hours []time.Time
}

func (h *Handler) MakeHourAvailable(ctx context.Context, cmd MakeHourAvailable) error {
	for _, hourTime := range cmd.Hours {
		err := h.hourRepo.UpdateHour(
			ctx,
			hourTime,
			func(h *domain.Hour) (*domain.Hour, error) {
				err := h.MakeAvailable()
				if err != nil {
					return nil, common.NewInvalidInputError(
						"unable-to-update-available",
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
