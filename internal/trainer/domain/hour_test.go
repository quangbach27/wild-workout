package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workout/trainer/domain"
)

func newAvailableHour(t *testing.T) *domain.Hour {
	t.Helper()

	factory := domain.MustNewFactory(domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 10,
		MinUtcHour:               0,
		MaxUtcHour:               24,
	})

	hourTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	hour, err := factory.NewAvailableHour(hourTime)
	require.NoError(t, err)

	return hour
}

func TestHour_Time(t *testing.T) {
	factory := domain.MustNewFactory(domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 10,
		MinUtcHour:               0,
		MaxUtcHour:               24,
	})

	hourTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	hour, err := factory.NewAvailableHour(hourTime)
	require.NoError(t, err)

	assert.True(t, hourTime.Equal(hour.Time()))
}

func TestHour_IsAvailable(t *testing.T) {
	hour := newAvailableHour(t)

	assert.True(t, hour.IsAvailable())
	assert.False(t, hour.HasTrainingScheduled())
	assert.Equal(t, domain.Available, hour.Availability())
}

func TestHour_HasTrainingScheduled(t *testing.T) {
	hour := newAvailableHour(t)

	require.NoError(t, hour.ScheduleTraining())

	assert.True(t, hour.HasTrainingScheduled())
	assert.False(t, hour.IsAvailable())
}

func TestHour_MakeNotAvailable(t *testing.T) {
	hour := newAvailableHour(t)

	err := hour.MakeNotAvailable()
	require.NoError(t, err)

	assert.Equal(t, domain.NotAvailable, hour.Availability())
	assert.False(t, hour.IsAvailable())
}

func TestHour_MakeNotAvailable_error_when_training_scheduled(t *testing.T) {
	hour := newAvailableHour(t)
	require.NoError(t, hour.ScheduleTraining())

	err := hour.MakeNotAvailable()

	assert.ErrorIs(t, err, domain.ErrTrainingScheduled)
	assert.True(t, hour.HasTrainingScheduled())
}

func TestHour_MakeAvailable(t *testing.T) {
	hour := newAvailableHour(t)
	require.NoError(t, hour.MakeNotAvailable())

	err := hour.MakeAvailable()
	require.NoError(t, err)

	assert.True(t, hour.IsAvailable())
}

func TestHour_MakeAvailable_error_when_training_scheduled(t *testing.T) {
	hour := newAvailableHour(t)
	require.NoError(t, hour.ScheduleTraining())

	err := hour.MakeAvailable()

	assert.ErrorIs(t, err, domain.ErrTrainingScheduled)
	assert.True(t, hour.HasTrainingScheduled())
}

func TestHour_ScheduleTraining(t *testing.T) {
	hour := newAvailableHour(t)

	err := hour.ScheduleTraining()
	require.NoError(t, err)

	assert.True(t, hour.HasTrainingScheduled())
}

func TestHour_ScheduleTraining_error_when_not_available(t *testing.T) {
	hour := newAvailableHour(t)
	require.NoError(t, hour.MakeNotAvailable())

	err := hour.ScheduleTraining()

	assert.ErrorIs(t, err, domain.ErrHourNotAvailable)
	assert.False(t, hour.HasTrainingScheduled())
}

func TestHour_ScheduleTraining_error_when_already_scheduled(t *testing.T) {
	hour := newAvailableHour(t)
	require.NoError(t, hour.ScheduleTraining())

	err := hour.ScheduleTraining()

	assert.ErrorIs(t, err, domain.ErrHourNotAvailable)
}

func TestHour_CancelTraining(t *testing.T) {
	hour := newAvailableHour(t)
	require.NoError(t, hour.ScheduleTraining())

	err := hour.CancelTraining()
	require.NoError(t, err)

	assert.True(t, hour.IsAvailable())
}

func TestHour_CancelTraining_error_when_not_scheduled(t *testing.T) {
	hour := newAvailableHour(t)

	err := hour.CancelTraining()

	assert.ErrorIs(t, err, domain.ErrNoTrainingScheduled)
}
