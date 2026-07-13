package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"workout/common"
	"workout/trainer/domain"
)

func validFactoryConfig() domain.HourFactoryConfig {
	return domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 10,
		MinUtcHour:               0,
		MaxUtcHour:               24,
	}
}

func TestNewFactory(t *testing.T) {
	config := validFactoryConfig()

	factory, err := domain.NewFactory(config)
	require.NoError(t, err)

	assert.False(t, factory.IsZero())
	assert.Equal(t, config, factory.Config())
}

func TestHourFactory_IsZero(t *testing.T) {
	var factory domain.HourFactory
	assert.True(t, factory.IsZero())

	factory = domain.MustNewFactory(validFactoryConfig())
	assert.False(t, factory.IsZero())
}

func TestHourFactoryConfig_Validate(t *testing.T) {
	testCases := []struct {
		Name   string
		Config domain.HourFactoryConfig
	}{
		{
			Name:   "valid",
			Config: validFactoryConfig(),
		},
		{
			Name: "max_weeks_zero",
			Config: domain.HourFactoryConfig{
				MaxWeeksInTheFutureToSet: 0,
				MinUtcHour:               0,
				MaxUtcHour:               24,
			},
		},
		{
			Name: "min_utc_hour_negative",
			Config: domain.HourFactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               -1,
				MaxUtcHour:               24,
			},
		},
		{
			Name: "min_utc_hour_too_big",
			Config: domain.HourFactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               25,
				MaxUtcHour:               25,
			},
		},
		{
			Name: "max_utc_hour_negative",
			Config: domain.HourFactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               0,
				MaxUtcHour:               -1,
			},
		},
		{
			Name: "max_utc_hour_too_big",
			Config: domain.HourFactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               0,
				MaxUtcHour:               25,
			},
		},
		{
			Name: "min_after_max",
			Config: domain.HourFactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               20,
				MaxUtcHour:               10,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Config.Validate()

			if tc.Name == "valid" {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)

			var commonErr common.Error
			require.ErrorAs(t, err, &commonErr)
			assert.NotEmpty(t, commonErr.Details)
		})
	}
}

func TestNewFactory_invalid_config(t *testing.T) {
	config := domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 0,
		MinUtcHour:               0,
		MaxUtcHour:               24,
	}

	factory, err := domain.NewFactory(config)

	require.Error(t, err)
	assert.True(t, factory.IsZero())
}

func TestMustNewFactory_panics_on_invalid_config(t *testing.T) {
	config := domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 0,
		MinUtcHour:               0,
		MaxUtcHour:               24,
	}

	assert.Panics(t, func() {
		domain.MustNewFactory(config)
	})
}

func TestHourFactory_NewAvailableHour(t *testing.T) {
	factory := domain.MustNewFactory(validFactoryConfig())
	hourTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)

	hour, err := factory.NewAvailableHour(hourTime)
	require.NoError(t, err)

	assert.True(t, hour.IsAvailable())
	assert.True(t, hourTime.Equal(hour.Time()))
}

func TestHourFactory_NewNotAvailableHour(t *testing.T) {
	factory := domain.MustNewFactory(validFactoryConfig())
	hourTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)

	hour, err := factory.NewNotAvailableHour(hourTime)
	require.NoError(t, err)

	assert.False(t, hour.IsAvailable())
	assert.Equal(t, domain.NotAvailable, hour.Availability())
}

func TestHourFactory_validateTime_not_full_hour(t *testing.T) {
	factory := domain.MustNewFactory(validFactoryConfig())
	hourTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour).Add(30 * time.Minute)

	_, err := factory.NewAvailableHour(hourTime)

	assert.ErrorIs(t, err, domain.ErrNotFullHour)
}

func TestHourFactory_validateTime_past_hour(t *testing.T) {
	factory := domain.MustNewFactory(validFactoryConfig())
	hourTime := time.Now().Add(-24 * time.Hour).Truncate(time.Hour)

	_, err := factory.NewAvailableHour(hourTime)

	assert.ErrorIs(t, err, domain.ErrPastHour)
}

func TestHourFactory_validateTime_current_hour(t *testing.T) {
	factory := domain.MustNewFactory(validFactoryConfig())
	hourTime := time.Now().Truncate(time.Hour)

	_, err := factory.NewAvailableHour(hourTime)

	assert.ErrorIs(t, err, domain.ErrPastHour)
}

func TestHourFactory_validateTime_too_distant_date(t *testing.T) {
	factory := domain.MustNewFactory(domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 1,
		MinUtcHour:               0,
		MaxUtcHour:               24,
	})
	hourTime := time.Now().Add(3 * 7 * 24 * time.Hour).Truncate(time.Hour)

	_, err := factory.NewAvailableHour(hourTime)

	var tooDistantErr domain.TooDistantDateError
	require.True(t, errors.As(err, &tooDistantErr))
	assert.Equal(t, 1, tooDistantErr.MaxWeeksInTheFutureToSet)
}

func TestHourFactory_validateTime_too_early_hour(t *testing.T) {
	factory := domain.MustNewFactory(domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 10,
		MinUtcHour:               8,
		MaxUtcHour:               20,
	})

	hourTime := nextDayAtUtcHour(t, 5)

	_, err := factory.NewAvailableHour(hourTime)

	var tooEarlyErr domain.TooEarlyHourError
	require.True(t, errors.As(err, &tooEarlyErr))
	assert.Equal(t, 8, tooEarlyErr.MinUtcHour)
}

func TestHourFactory_validateTime_too_late_hour(t *testing.T) {
	factory := domain.MustNewFactory(domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 10,
		MinUtcHour:               8,
		MaxUtcHour:               20,
	})

	hourTime := nextDayAtUtcHour(t, 22)

	_, err := factory.NewAvailableHour(hourTime)

	var tooLateErr domain.TooLateHourError
	require.True(t, errors.As(err, &tooLateErr))
	assert.Equal(t, 20, tooLateErr.MaxUtcHour)
}

// nextDayAtUtcHour returns a full, future UTC hour so tests are not flaky
// depending on the time of day they run.
func nextDayAtUtcHour(t *testing.T, utcHour int) time.Time {
	t.Helper()

	now := time.Now().UTC().AddDate(0, 0, 1)
	return time.Date(now.Year(), now.Month(), now.Day(), utcHour, 0, 0, 0, time.UTC)
}
