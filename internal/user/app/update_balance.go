package app

import (
	"context"
)

type UpdateBalance struct {
	FirebaseUserUUID string
	AmountChange     int
}

func (h *Handler) UpdateBalance(ctx context.Context, cmd UpdateBalance) error {
	return nil
}
