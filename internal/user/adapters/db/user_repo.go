package db

import (
	"context"
	"fmt"
	"workout/common"
	"workout/user/adapters/db/dbmodels"
	"workout/user/app/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepo {
	if db == nil {
		panic("db can't be nil")
	}

	return &UserRepo{db: db}
}

func (r *UserRepo) GetUser(ctx context.Context, uuid models.UserUUID) (*models.User, error) {
	queries := dbmodels.New(r.db)

	dbUser, err := queries.GetUser(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return unmarshalUserDbToDomain(dbUser), nil
}

func (r *UserRepo) CreateUser(ctx context.Context, user *models.User) error {
	queries := dbmodels.New(r.db)

	_, err := queries.CreateUser(ctx, dbmodels.CreateUserParams{
		UserUuid:    user.UUID(),
		FirebaseUid: user.FirebaseUUID(),
		Username:    user.Username(),
		Role:        user.Role(),
		Balance:     int32(user.Balance()),
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateBalance fetches the user by their Firebase uid, applies
// amountChange via User.UpdateBalance (which rejects the change if it would
// take the balance below zero), and persists the resulting balance. The
// fetch and write happen in a single serializable transaction so concurrent
// balance changes for the same user can't race each other.
func (r *UserRepo) UpdateBalance(ctx context.Context, userID string, amountChange int) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		dbUser, err := queries.GetUserByFirebaseUID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		user := unmarshalUserDbToDomain(dbUser)

		if err := user.UpdateBalance(amountChange); err != nil {
			return err
		}

		err = queries.UpdateBalance(ctx, dbmodels.UpdateBalanceParams{
			FirebaseUid: userID,
			Balance:     int32(user.Balance()),
		})
		if err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}

		return nil
	})
}

func unmarshalUserDbToDomain(dbUser dbmodels.UserUser) *models.User {
	return models.UnmarshalUserFromDB(
		dbUser.UserUuid,
		dbUser.FirebaseUid,
		dbUser.Username,
		dbUser.Role,
		int(dbUser.Balance),
	)
}
