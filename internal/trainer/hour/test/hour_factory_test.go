package hour_test

import (
	"testing"

	"github.com/quangbach27/wild-workout/internal/trainer/domain/hour"
	"github.com/stretchr/testify/assert"
)

func TestFactoryConfig_Validate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name        string
		Config      hour.FactoryConfig
		ExpectedErr string
	}{
		{
			Name: "valid",
			Config: hour.FactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               10,
				MaxUtcHour:               12,
			},
			ExpectedErr: "",
		},
		{
			Name: "equal_min_and_max_hour",
			Config: hour.FactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               12,
				MaxUtcHour:               12,
			},
			ExpectedErr: "",
		},
		{
			Name: "min_hour_after_max_hour",
			Config: hour.FactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               13,
				MaxUtcHour:               12,
			},
			ExpectedErr: "MinUtcHour (13) can't be after MaxUtcHour (12)",
		},
		{
			Name: "zero_max_weeks",
			Config: hour.FactoryConfig{
				MaxWeeksInTheFutureToSet: 0,
				MinUtcHour:               10,
				MaxUtcHour:               12,
			},
			ExpectedErr: "MaxWeeksInTheFutureToSet should be greater than 1, but is 0",
		},
		{
			Name: "sub_zero_min_hour",
			Config: hour.FactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               -1,
				MaxUtcHour:               12,
			},
			ExpectedErr: "MinUtcHour should be value between 0 and 24, but is -1",
		},
		{
			Name: "sub_zero_max_hour",
			Config: hour.FactoryConfig{
				MaxWeeksInTheFutureToSet: 10,
				MinUtcHour:               10,
				MaxUtcHour:               -1,
			},
			ExpectedErr: "MinUtcHour should be value between 0 and 24, but is -1; MinUtcHour (10) can't be after MaxUtcHour (-1)",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			err := testCase.Config.Validate()

			expectedErrStr := testCase.ExpectedErr
			if expectedErrStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, expectedErrStr)
			}
		})
	}
}

func TestNewFactory_invalid_config(t *testing.T) {
	factory, err := hour.NewFactory(hour.FactoryConfig{})	
	assert.Error(t, err)
	assert.Zero(t, factory)
}
