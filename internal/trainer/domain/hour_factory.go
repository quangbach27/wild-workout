package domain

import (
	"errors"
	"fmt"
	"time"
	"workout/common"
)

type HourFactoryConfig struct {
	MaxWeeksInTheFutureToSet int
	MinUtcHour               int
	MaxUtcHour               int
}

func (f HourFactoryConfig) Validate() error {
	var errDetails []common.ErrorDetails

	if f.MaxWeeksInTheFutureToSet < 1 {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "HourFactoryConfig",
			ErrorSlug:  "invalid-max-weeks",
			Message: fmt.Sprintf("MaxWeeksInTheFutureToSet should be greater than 1, but is %d",
				f.MaxWeeksInTheFutureToSet),
		})
	}
	if f.MinUtcHour < 0 || f.MinUtcHour > 24 {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "HourFactoryConfig",
			ErrorSlug:  "invalid-min-utc-hour",
			Message: fmt.Sprintf("MinUtcHour should be value between 0 and 24, but is %d",
				f.MinUtcHour),
		})
	}
	if f.MaxUtcHour < 0 || f.MaxUtcHour > 24 {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "HourFactoryConfig",
			ErrorSlug:  "invalid-max-utc-hour",
			Message: fmt.Sprintf("MacUtcHour should be value between 0 and 24, but is %d",
				f.MaxUtcHour),
		})
	}

	if f.MinUtcHour > f.MaxUtcHour {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "HourFactoryConfig",
			ErrorSlug:  "invalid-min-utc-hour",
			Message: fmt.Sprintf(
				"MinUtcHour (%d) can't be after MaxUtcHour (%d)",
				f.MinUtcHour, f.MaxUtcHour,
			),
		})
	}

	if len(errDetails) != 0 {
		return common.NewInvalidInputError(
			"invalid-hour-factory-config",
			"hour factory config is not valid",
		).WithDetails(errDetails)
	}

	return nil
}

type HourFactory struct {
	config HourFactoryConfig
}

func NewFactory(config HourFactoryConfig) (HourFactory, error) {
	if err := config.Validate(); err != nil {
		return HourFactory{}, err
	}

	return HourFactory{config: config}, nil
}

func MustNewFactory(fc HourFactoryConfig) HourFactory {
	f, err := NewFactory(fc)
	if err != nil {
		panic(err)
	}

	return f
}

func (f HourFactory) Config() HourFactoryConfig {
	return f.config
}

func (f HourFactory) IsZero() bool {
	return f == HourFactory{}
}

func (f HourFactory) NewAvailableHour(hour time.Time) (*Hour, error) {
	if err := f.validateTime(hour); err != nil {
		return nil, err
	}

	return &Hour{
		hour:         hour,
		availability: Available,
	}, nil
}

func (f HourFactory) NewNotAvailableHour(hour time.Time) (*Hour, error) {
	if err := f.validateTime(hour); err != nil {
		return nil, err
	}

	return &Hour{
		hour:         hour,
		availability: NotAvailable,
	}, nil
}

var (
	ErrNotFullHour = errors.New("hour should be a full hour")
	ErrPastHour    = errors.New("cannot create hour from past")
)

type TooDistantDateError struct {
	MaxWeeksInTheFutureToSet int
	ProvidedDate             time.Time
}

func (e TooDistantDateError) Error() string {
	return fmt.Sprintf(
		"schedule can be only set for next %d weeks, provided date: %s",
		e.MaxWeeksInTheFutureToSet,
		e.ProvidedDate,
	)
}

type TooEarlyHourError struct {
	MinUtcHour   int
	ProvidedTime time.Time
}

func (e TooEarlyHourError) Error() string {
	return fmt.Sprintf(
		"too early hour, min UTC hour: %d, provided time: %s",
		e.MinUtcHour,
		e.ProvidedTime,
	)
}

type TooLateHourError struct {
	MaxUtcHour   int
	ProvidedTime time.Time
}

func (e TooLateHourError) Error() string {
	return fmt.Sprintf(
		"too late hour, min UTC hour: %d, provided time: %s",
		e.MaxUtcHour,
		e.ProvidedTime,
	)
}
func (f HourFactory) validateTime(hour time.Time) error {
	if !hour.Round(time.Hour).Equal(hour) {
		return ErrNotFullHour
	}

	// AddDate is better than Add for adding days, because not every day have 24h!
	if hour.After(time.Now().AddDate(0, 0, f.config.MaxWeeksInTheFutureToSet*7)) {
		return TooDistantDateError{
			MaxWeeksInTheFutureToSet: f.config.MaxWeeksInTheFutureToSet,
			ProvidedDate:             hour,
		}
	}

	currentHour := time.Now().Truncate(time.Hour)
	if hour.Before(currentHour) || hour.Equal(currentHour) {
		return ErrPastHour
	}
	if hour.UTC().Hour() > f.config.MaxUtcHour {
		return TooLateHourError{
			MaxUtcHour:   f.config.MaxUtcHour,
			ProvidedTime: hour,
		}
	}
	if hour.UTC().Hour() < f.config.MinUtcHour {
		return TooEarlyHourError{
			MinUtcHour:   f.config.MinUtcHour,
			ProvidedTime: hour,
		}
	}

	return nil
}
