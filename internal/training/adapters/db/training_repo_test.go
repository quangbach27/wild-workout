package db_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workout/common"
	"workout/common/testutils"
	"workout/training/adapters/db"
	"workout/training/domain"
)

func newTestRepo(t *testing.T) *db.TrainingRepo {
	t.Helper()

	pool, cleanup := testutils.NewDB(context.Background())
	t.Cleanup(cleanup)

	return db.NewTraningRepository(pool)
}

func newTestUser(t *testing.T, userType domain.UserType) domain.User {
	t.Helper()

	user, err := domain.NewUser(domain.UserUUID{UUID: common.NewUUIDv7()}, userType)
	require.NoError(t, err)

	return user
}

func newTestTraining(t *testing.T, owner domain.User) *domain.Training {
	t.Helper()

	tr, err := domain.NewTraining(owner.UUID(), "user-name", time.Now().Add(48*time.Hour))
	require.NoError(t, err)

	return tr
}

func TestTrainingRepo_AddTraining_GetTraining(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)

	err := repo.AddTraining(ctx, tr)
	require.NoError(t, err)

	fetched, err := repo.GetTraining(ctx, tr.UUID(), owner)
	require.NoError(t, err)

	assert.Equal(t, tr.UUID(), fetched.UUID())
	assert.Equal(t, tr.UserUUID(), fetched.UserUUID())
	assert.Equal(t, tr.UserName(), fetched.UserName())
	assert.True(t, tr.Time().Equal(fetched.Time()))
	assert.Equal(t, tr.Notes(), fetched.Notes())
	assert.Equal(t, tr.IsCanceled(), fetched.IsCanceled())
	assert.False(t, fetched.IsRescheduleProposed())
}

func TestTrainingRepo_GetTraining_VisibleToTrainer(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	trainer := newTestUser(t, domain.Trainer)

	fetched, err := repo.GetTraining(ctx, tr.UUID(), trainer)
	require.NoError(t, err)
	assert.Equal(t, tr.UUID(), fetched.UUID())
}

func TestTrainingRepo_GetTraining_ForbiddenForOtherAttendee(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	other := newTestUser(t, domain.Attendee)

	_, err := repo.GetTraining(ctx, tr.UUID(), other)
	require.Error(t, err)

	var commonErr common.Error
	require.ErrorAs(t, err, &commonErr)
	assert.Equal(t, http.StatusForbidden, commonErr.HttpErrorCode)
}

func TestTrainingRepo_GetTraining_NotFound(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)

	_, err := repo.GetTraining(ctx, domain.TrainingUUID{UUID: common.NewUUIDv7()}, owner)
	require.Error(t, err)
}

func TestTrainingRepo_UpdateTraining_Notes(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	err := repo.UpdateTraining(ctx, tr.UUID(), owner, func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
		require.NoError(t, tr.UpdateNotes("updated notes"))
		return tr, nil
	})
	require.NoError(t, err)

	fetched, err := repo.GetTraining(ctx, tr.UUID(), owner)
	require.NoError(t, err)
	assert.Equal(t, "updated notes", fetched.Notes())
}

func TestTrainingRepo_UpdateTraining_Reschedule(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	newTime := time.Now().Add(72 * time.Hour)

	err := repo.UpdateTraining(ctx, tr.UUID(), owner, func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
		require.NoError(t, tr.RescheduleTraining(newTime))
		return tr, nil
	})
	require.NoError(t, err)

	fetched, err := repo.GetTraining(ctx, tr.UUID(), owner)
	require.NoError(t, err)
	assert.True(t, newTime.Equal(fetched.Time()))
}

func TestTrainingRepo_UpdateTraining_ProposeReschedule(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	proposedTime := time.Now().Add(96 * time.Hour)

	err := repo.UpdateTraining(ctx, tr.UUID(), owner, func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
		tr.ProposeReschedule(proposedTime, domain.Attendee)
		return tr, nil
	})
	require.NoError(t, err)

	fetched, err := repo.GetTraining(ctx, tr.UUID(), owner)
	require.NoError(t, err)
	require.True(t, fetched.IsRescheduleProposed())
	assert.True(t, proposedTime.Equal(fetched.ProposedNewTime()))
	assert.Equal(t, domain.Attendee, fetched.MoveProposedBy())
}

func TestTrainingRepo_UpdateTraining_Cancel(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	err := repo.UpdateTraining(ctx, tr.UUID(), owner, func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
		require.NoError(t, tr.Cancel())
		return tr, nil
	})
	require.NoError(t, err)

	fetched, err := repo.GetTraining(ctx, tr.UUID(), owner)
	require.NoError(t, err)
	assert.True(t, fetched.IsCanceled())
}

func TestTrainingRepo_UpdateTraining_ForbiddenForOtherAttendee(t *testing.T) {
	t.Parallel()

	repo := newTestRepo(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	other := newTestUser(t, domain.Attendee)

	err := repo.UpdateTraining(ctx, tr.UUID(), other, func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
		require.NoError(t, tr.UpdateNotes("should not be applied"))
		return tr, nil
	})
	require.Error(t, err)

	var commonErr common.Error
	require.ErrorAs(t, err, &commonErr)
	assert.Equal(t, http.StatusForbidden, commonErr.HttpErrorCode)

	fetched, err := repo.GetTraining(ctx, tr.UUID(), owner)
	require.NoError(t, err)
	assert.Equal(t, tr.Notes(), fetched.Notes())
}
