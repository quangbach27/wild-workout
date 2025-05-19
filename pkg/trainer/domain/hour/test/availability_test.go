package hour_test

import (
	"testing"

	"github.com/quangbach27/wild-workout/pkg/trainer/domain/hour"
	"github.com/stretchr/testify/require"
)

func TestHour_MakeNotAvailable(t *testing.T) {
	newNotAvailableHour(t)
}

func TestHour_MakeNotAvailable_with_schudule_training(t *testing.T) {
	h := newHourWithScheduledTraining(t)

	require.Equal(t, hour.ErrTrainingScheduled, h.MakeNotAvailable())
}

func newHourWithScheduledTraining(t *testing.T) *hour.Hour {
	h, err := testHourFactory.NewAvailableHour(validTrainingHour())
	require.NoError(t, err)

	require.NoError(t, h.ScheduleTraining())

	return h
}

func newNotAvailableHour(t *testing.T) *hour.Hour {
	h, err := testHourFactory.NewAvailableHour(validTrainingHour())
	require.NoError(t, err)

	require.NoError(t, h.MakeNotAvailable())
	require.False(t, h.IsAvailable())

	return h
}
