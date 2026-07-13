package db_test

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workout/common/testutils"
	adaptersdb "workout/trainer/adapters/db"
	"workout/trainer/adapters/db/dbmodels"
	"workout/trainer/domain"
)

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool, cleanUp := testutils.NewDB(context.Background())
	t.Cleanup(cleanUp)

	return pool
}

func testHourFactory() domain.HourFactory {
	return domain.MustNewFactory(domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 10,
		MinUtcHour:               0,
		MaxUtcHour:               24,
	})
}

func newTestRepo(t *testing.T) (*adaptersdb.HourRepo, *pgxpool.Pool) {
	t.Helper()

	pool := newTestPool(t)
	repo := adaptersdb.NewHourRepository(pool, testHourFactory())

	return repo, pool
}

var hourOffset atomic.Int64

func seedHour(t *testing.T, pool *pgxpool.Pool, hour time.Time, availability domain.Availability) {
	t.Helper()

	queries := dbmodels.New(pool)
	err := queries.UpsertHour(context.Background(), dbmodels.UpsertHourParams{
		Hour:         hour,
		Availability: availability,
	})
	require.NoError(t, err)
}

func TestNewHourRepository(t *testing.T) {
	t.Parallel()

	t.Run("panics_on_nil_db", func(t *testing.T) {
		t.Parallel()

		assert.Panics(t, func() {
			adaptersdb.NewHourRepository(nil, testHourFactory())
		})
	})

	t.Run("panics_on_zero_hour_factory", func(t *testing.T) {
		t.Parallel()

		pool := newTestPool(t)

		assert.Panics(t, func() {
			adaptersdb.NewHourRepository(pool, domain.HourFactory{})
		})
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		pool := newTestPool(t)

		repo := adaptersdb.NewHourRepository(pool, testHourFactory())

		assert.NotNil(t, repo)
	})
}

func TestHourRepo_GetHour(t *testing.T) {
	t.Parallel()

	repo, pool := newTestRepo(t)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		hourTime := newValidHourTime()
		seedHour(t, pool, hourTime, domain.Available)

		hour, err := repo.GetHour(context.Background(), hourTime)
		require.NoError(t, err)

		assert.True(t, hourTime.Equal(hour.Time()))
		assert.True(t, hour.IsAvailable())
	})

	t.Run("returns_not_available_hour_when_hour_not_found_in_db", func(t *testing.T) {
		t.Parallel()

		hourTime := newValidHourTime()

		hour, err := repo.GetHour(context.Background(), hourTime)
		require.NoError(t, err)

		assert.True(t, hourTime.Equal(hour.Time()))
		assert.Equal(t, domain.NotAvailable, hour.Availability())
	})
}

func TestHourRepo_UpdateHour(t *testing.T) {
	t.Parallel()

	repo, pool := newTestRepo(t)

	t.Run("updates_existing_hour", func(t *testing.T) {
		t.Parallel()

		hourTime := newValidHourTime()
		seedHour(t, pool, hourTime, domain.Available)

		err := repo.UpdateHour(context.Background(), hourTime, func(h *domain.Hour) (*domain.Hour, error) {
			require.NoError(t, h.MakeNotAvailable())
			return h, nil
		})
		require.NoError(t, err)

		hour, err := repo.GetHour(context.Background(), hourTime)
		require.NoError(t, err)
		assert.False(t, hour.IsAvailable())
	})

	t.Run("creates_hour_from_default_when_missing", func(t *testing.T) {
		t.Parallel()

		hourTime := newValidHourTime()
		err := repo.UpdateHour(context.Background(), hourTime, func(h *domain.Hour) (*domain.Hour, error) {
			assert.Equal(t, domain.NotAvailable, h.Availability())
			require.NoError(t, h.MakeAvailable())
			return h, nil
		})
		require.NoError(t, err)

		hour, err := repo.GetHour(context.Background(), hourTime)
		require.NoError(t, err)
		assert.True(t, hour.IsAvailable())
	})

	t.Run("update_fn_error_rolls_back", func(t *testing.T) {
		t.Parallel()

		hourTime := newValidHourTime()
		seedHour(t, pool, hourTime, domain.Available)

		updateErr := errors.New("update failed")
		err := repo.UpdateHour(context.Background(), hourTime, func(h *domain.Hour) (*domain.Hour, error) {
			return nil, updateErr
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, updateErr)

		hour, err := repo.GetHour(context.Background(), hourTime)
		require.NoError(t, err)
		assert.True(t, hour.IsAvailable())
	})
}

// usedHours is storing hours used during the test,
// to ensure that within one test run we are not using the same hour
// (it should be not a problem between test runs)
var usedHours = sync.Map{}

func newValidHourTime() time.Time {
	for {
		minTime := time.Now().AddDate(0, 0, 1)

		minTimestamp := minTime.Unix()
		maxTimestamp := minTime.AddDate(0, 0, testHourFactory().Config().MaxWeeksInTheFutureToSet*7).Unix()

		t := time.Unix(rand.Int63n(maxTimestamp-minTimestamp)+minTimestamp, 0).Truncate(time.Hour).Local()

		_, alreadyUsed := usedHours.LoadOrStore(t.Unix(), true)
		if !alreadyUsed {
			return t
		}
	}
}
