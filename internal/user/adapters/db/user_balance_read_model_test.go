//go:build integration

package db_test

import (
	"context"
	"testing"
	"workout/common/testutils"
	"workout/user/adapters/db"
	"workout/user/app/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestBalanceReadModel(t *testing.T) *db.UserBalanceReadModel {
	t.Helper()

	pool, cleanup := testutils.NewDB(context.Background())
	t.Cleanup(cleanup)

	return db.NewUserBalanceReadModel(pool)
}

func TestUserBalanceReadModel_GetUserBalance(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	rm := newTestBalanceReadModel(t)
	ctx := context.Background()

	firebaseUID := newFirebaseUID(t)
	user, err := models.NewUser(firebaseUID, "Test User", models.RoleAttendee)
	require.NoError(t, err)
	require.NoError(t, repo.CreateUser(ctx, user))

	balance, err := rm.GetUserBalance(ctx, firebaseUID)
	require.NoError(t, err)
	assert.Equal(t, user.Balance(), balance)

	require.NoError(t, repo.UpdateBalance(ctx, firebaseUID, -1))

	balance, err = rm.GetUserBalance(ctx, firebaseUID)
	require.NoError(t, err)
	assert.Equal(t, user.Balance()-1, balance)
}

func TestUserBalanceReadModel_GetUserBalance_UnknownUser(t *testing.T) {
	t.Parallel()

	rm := newTestBalanceReadModel(t)

	_, err := rm.GetUserBalance(context.Background(), "does-not-exist")
	require.Error(t, err)
}
