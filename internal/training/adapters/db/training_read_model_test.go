//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workout/common/testutils"
	"workout/training/adapters/db"
	"workout/training/app/query"
	"workout/training/domain"
)

func newTestRepoAndReadModel(t *testing.T) (*db.TrainingRepo, *db.TrainingReadModel) {
	t.Helper()

	pool, cleanup := testutils.NewDB(context.Background())
	t.Cleanup(cleanup)

	return db.NewTraningRepository(pool), db.NewTrainingReadModel(pool)
}

func findByUUID(trainings []query.Training, uuid domain.TrainingUUID) (query.Training, bool) {
	for _, tr := range trainings {
		if tr.UUID == uuid {
			return tr, true
		}
	}

	return query.Training{}, false
}

func TestTrainingReadModel_FindTrainingsForUser(t *testing.T) {
	t.Parallel()

	repo, rm := newTestRepoAndReadModel(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	other := newTestUser(t, domain.Attendee)
	otherTr := newTestTraining(t, other)
	require.NoError(t, repo.AddTraining(ctx, otherTr))

	trainings, err := rm.FindTrainingsForUser(ctx, owner.ID())
	require.NoError(t, err)
	require.Len(t, trainings, 1)

	got := trainings[0]
	assert.Equal(t, tr.UUID(), got.UUID)
	assert.Equal(t, tr.UserID(), got.UserID)
	assert.Equal(t, tr.UserName(), got.User)
	assert.True(t, tr.Time().Equal(got.Time))
	assert.Equal(t, tr.Notes(), got.Notes)
	assert.Nil(t, got.ProposedTime)
	assert.Nil(t, got.MoveProposedBy)
	assert.True(t, got.CanBeCancelled)
}

func TestTrainingReadModel_FindTrainingsForUser_ExcludesCanceled(t *testing.T) {
	t.Parallel()

	repo, rm := newTestRepoAndReadModel(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	err := repo.UpdateTraining(ctx, tr.UUID(), owner, func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
		require.NoError(t, tr.Cancel())
		return tr, nil
	})
	require.NoError(t, err)

	trainings, err := rm.FindTrainingsForUser(ctx, owner.ID())
	require.NoError(t, err)
	assert.Empty(t, trainings)
}

func TestTrainingReadModel_FindTrainingsForUser_ProposedReschedule(t *testing.T) {
	t.Parallel()

	repo, rm := newTestRepoAndReadModel(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	tr := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, tr))

	proposedTime := time.Now().Add(96 * time.Hour).Truncate(time.Microsecond)

	err := repo.UpdateTraining(ctx, tr.UUID(), owner, func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
		tr.ProposeReschedule(proposedTime, domain.Attendee)
		return tr, nil
	})
	require.NoError(t, err)

	trainings, err := rm.FindTrainingsForUser(ctx, owner.ID())
	require.NoError(t, err)
	require.Len(t, trainings, 1)

	got := trainings[0]
	require.NotNil(t, got.ProposedTime)
	assert.True(t, proposedTime.Equal(*got.ProposedTime))
	require.NotNil(t, got.MoveProposedBy)
	assert.Equal(t, domain.Attendee.String(), *got.MoveProposedBy)
}

func TestTrainingReadModel_ListAllTrainings(t *testing.T) {
	t.Parallel()

	repo, rm := newTestRepoAndReadModel(t)
	ctx := context.Background()

	owner := newTestUser(t, domain.Attendee)
	visible := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, visible))

	canceled := newTestTraining(t, owner)
	require.NoError(t, repo.AddTraining(ctx, canceled))
	err := repo.UpdateTraining(ctx, canceled.UUID(), owner, func(ctx context.Context, tr *domain.Training) (*domain.Training, error) {
		require.NoError(t, tr.Cancel())
		return tr, nil
	})
	require.NoError(t, err)

	trainings, err := rm.ListAllTrainings(ctx)
	require.NoError(t, err)

	got, ok := findByUUID(trainings, visible.UUID())
	require.True(t, ok, "expected training to be present in ListAllTrainings")
	assert.Equal(t, visible.UserID(), got.UserID)

	_, ok = findByUUID(trainings, canceled.UUID())
	assert.False(t, ok, "canceled training should not appear in ListAllTrainings")
}
