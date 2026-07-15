package db

import (
	"context"
	"fmt"
	"workout/user/adapters/db/dbmodels"

	"github.com/jackc/pgx/v5/pgxpool"
)

// UserBalanceReadModel is a read-only, minimal query for a user's training
// balance, used by the gRPC GetTrainingBalance endpoint. It's kept separate
// from UserRepo so that read doesn't have to pull (and unmarshal) the full
// user aggregate just to answer "what's the balance".
type UserBalanceReadModel struct {
	db *pgxpool.Pool
}

func NewUserBalanceReadModel(db *pgxpool.Pool) *UserBalanceReadModel {
	if db == nil {
		panic("db can't be nil")
	}

	return &UserBalanceReadModel{db: db}
}

func (r *UserBalanceReadModel) GetUserBalance(ctx context.Context, firebaseUserID string) (int, error) {
	queries := dbmodels.New(r.db)

	balance, err := queries.GetUserBalance(ctx, firebaseUserID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user balance: %w", err)
	}

	return int(balance), nil
}
