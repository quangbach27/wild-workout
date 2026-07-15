//go:build integration

package db_test

import (
	"context"
	"testing"
	"workout/common"
	"workout/common/testutils"
	"workout/user/adapters/db"
	"workout/user/app/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepo(t *testing.T) *db.UserRepo {
	t.Helper()

	pool, cleanup := testutils.NewDB(context.Background())
	t.Cleanup(cleanup)

	return db.NewUserRepository(pool)
}

func newFirebaseUID(t *testing.T) string {
	t.Helper()

	return common.NewUUIDv7().String()
}

func TestUserRepo_CreateUser_GetUser(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	user, err := models.NewUser(newFirebaseUID(t), "Mariusz Pudzianowski", models.RoleAttendee)
	require.NoError(t, err)

	require.NoError(t, repo.CreateUser(ctx, user))

	fetched, err := repo.GetUser(ctx, user.UUID())
	require.NoError(t, err)

	assert.Equal(t, user.UUID(), fetched.UUID())
	assert.Equal(t, user.FirebaseUUID(), fetched.FirebaseUUID())
	assert.Equal(t, user.Username(), fetched.Username())
	assert.Equal(t, user.Role(), fetched.Role())
	assert.Equal(t, user.Balance(), fetched.Balance())
}

func TestUserRepo_GetUser_NotFound(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.GetUser(ctx, models.UserUUID{UUID: common.NewUUIDv7()})
	require.Error(t, err)
}

func TestUserRepo_UpdateBalance(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	firebaseUID := newFirebaseUID(t)
	user, err := models.NewUser(firebaseUID, "Test User", models.RoleAttendee)
	require.NoError(t, err)
	require.NoError(t, repo.CreateUser(ctx, user))

	startingBalance := user.Balance()

	require.NoError(t, repo.UpdateBalance(ctx, firebaseUID, -1))

	fetched, err := repo.GetUser(ctx, user.UUID())
	require.NoError(t, err)
	assert.Equal(t, startingBalance-1, fetched.Balance())

	require.NoError(t, repo.UpdateBalance(ctx, firebaseUID, 2))

	fetched, err = repo.GetUser(ctx, user.UUID())
	require.NoError(t, err)
	assert.Equal(t, startingBalance+1, fetched.Balance())
}

func TestUserRepo_UpdateBalance_RejectsNegativeResult(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	firebaseUID := newFirebaseUID(t)
	user, err := models.NewUser(firebaseUID, "Test User", models.RoleAttendee)
	require.NoError(t, err)
	require.NoError(t, repo.CreateUser(ctx, user))

	err = repo.UpdateBalance(ctx, firebaseUID, -(user.Balance() + 1))
	require.Error(t, err)

	fetched, err := repo.GetUser(ctx, user.UUID())
	require.NoError(t, err)
	assert.Equal(t, user.Balance(), fetched.Balance(), "balance must be unchanged after a rejected update")
}

func TestUserRepo_UpdateBalance_UnknownUser(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.UpdateBalance(ctx, newFirebaseUID(t), 1)
	require.Error(t, err)
}
