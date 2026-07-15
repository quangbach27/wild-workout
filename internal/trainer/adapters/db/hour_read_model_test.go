//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adaptersdb "workout/trainer/adapters/db"
	"workout/trainer/domain"
)

func testReadModelFactoryConfig() domain.HourFactoryConfig {
	return domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 10,
		MinUtcHour:               8,
		MaxUtcHour:               10,
	}
}

func newTestReadModel(t *testing.T) (*adaptersdb.HourReadModel, *pgxpool.Pool) {
	t.Helper()

	pool := newTestPool(t)
	rm := adaptersdb.NewHourReadModel(pool, testReadModelFactoryConfig())

	return rm, pool
}

func endOfDay(day time.Time) time.Time {
	return day.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
}

func TestNewHourReadModel(t *testing.T) {
	t.Parallel()

	t.Run("panics_on_nil_db", func(t *testing.T) {
		t.Parallel()

		assert.Panics(t, func() {
			adaptersdb.NewHourReadModel(nil, testReadModelFactoryConfig())
		})
	})

	t.Run("panics_on_invalid_factory_config", func(t *testing.T) {
		t.Parallel()

		pool := newTestPool(t)

		assert.Panics(t, func() {
			adaptersdb.NewHourReadModel(pool, domain.HourFactoryConfig{})
		})
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		pool := newTestPool(t)

		rm := adaptersdb.NewHourReadModel(pool, testReadModelFactoryConfig())

		assert.NotNil(t, rm)
	})
}

// TestHourReadModel_ListAvailableHours uses a single day (and its
// following day, for the multi-day case) shared across all its subtests,
// each touching a different hour slot. newValidHourTime only guarantees
// distinct hours, not distinct days, so subtests can't safely draw their
// own independent day each — two could land on the same calendar day and
// stomp on each other's expectations. Subtests therefore run sequentially
// (no t.Parallel inside them) since they share this mutable day's rows.
func TestHourReadModel_ListAvailableHours(t *testing.T) {
	t.Parallel()

	rm, pool := newTestReadModel(t)

	dayTime := newValidHourTime().UTC()
	day := time.Date(dayTime.Year(), dayTime.Month(), dayTime.Day(), 0, 0, 0, 0, time.UTC)

	t.Run("returns_seeded_hours_with_availability", func(t *testing.T) {
		seedHour(t, pool, day.Add(8*time.Hour), domain.Available)
		seedHour(t, pool, day.Add(9*time.Hour), domain.TrainingScheduled)
		seedHour(t, pool, day.Add(10*time.Hour), domain.NotAvailable)

		dates, err := rm.ListAvailableHours(context.Background(), day, endOfDay(day))
		require.NoError(t, err)
		require.Len(t, dates, 1)

		got := dates[0]
		assert.True(t, day.Equal(got.Date))
		assert.True(t, got.HasFreeHours)
		require.Len(t, got.Hours, 3)

		// available
		assert.True(t, day.Add(8*time.Hour).Equal(got.Hours[0].Hour))
		assert.True(t, got.Hours[0].Available)
		assert.False(t, got.Hours[0].HasTrainingScheduled)

		// training scheduled
		assert.True(t, day.Add(9*time.Hour).Equal(got.Hours[1].Hour))
		assert.False(t, got.Hours[1].Available)
		assert.True(t, got.Hours[1].HasTrainingScheduled)

		// not available
		assert.True(t, day.Add(10*time.Hour).Equal(got.Hours[2].Hour))
		assert.False(t, got.Hours[2].Available)
		assert.False(t, got.Hours[2].HasTrainingScheduled)
	})

	t.Run("spans_multiple_days", func(t *testing.T) {
		day2 := day.AddDate(0, 0, 1)
		seedHour(t, pool, day2.Add(9*time.Hour), domain.Available)

		dates, err := rm.ListAvailableHours(context.Background(), day, endOfDay(day2))
		require.NoError(t, err)
		require.Len(t, dates, 2)

		// day's HasFreeHours is already true from the previous subtest's
		// hour 8 seed — this subtest only asserts the range/day-grouping.
		assert.True(t, day.Equal(dates[0].Date))
		assert.True(t, day2.Equal(dates[1].Date))
		assert.True(t, dates[1].HasFreeHours)
	})

	t.Run("returns_error_when_context_canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := rm.ListAvailableHours(ctx, day, endOfDay(day))
		assert.Error(t, err)
	})
}
