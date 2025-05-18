package hour

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type FactoryConfig struct {
	MaxWeeksInTheFutureToSet int
	MinUtcHour               int
	MaxUtcHour               int
}

func (factoryConfig FactoryConfig) Validate() error {
	var err error

	if factoryConfig.MaxWeeksInTheFutureToSet < 1 {
		err = multierr.Append(
			err,
			errors.Errorf(
				"MaxWeeksInTheFutureToSet should be greater than 1, but is %d",
				factoryConfig.MaxWeeksInTheFutureToSet,
			),
		)
	}
	if factoryConfig.MinUtcHour < 0 || factoryConfig.MinUtcHour > 24 {
		err = multierr.Append(
			err,
			errors.Errorf(
				"MinUtcHour should be value between 0 and 24, but is %d",
				factoryConfig.MinUtcHour,
			),
		)
	}
	if factoryConfig.MaxUtcHour < 0 || factoryConfig.MaxUtcHour > 24 {
		err = multierr.Append(
			err,
			errors.Errorf(
				"MinUtcHour should be value between 0 and 24, but is %d",
				factoryConfig.MaxUtcHour,
			),
		)
	}

	if factoryConfig.MinUtcHour > factoryConfig.MaxUtcHour {
		err = multierr.Append(
			err,
			errors.Errorf(
				"MinUtcHour (%d) can't be after MaxUtcHour (%d)",
				factoryConfig.MinUtcHour, factoryConfig.MaxUtcHour,
			),
		)
	}

	return err
}

type Factory struct {
	factoryConfig FactoryConfig
}

func NewFactory(config FactoryConfig) (Factory, error) {
	if err := config.Validate(); err != nil {
		return Factory{}, errors.Wrap(err, "invalid config passed to factory")
	}

	return Factory{factoryConfig: config}, nil
}

func MustNewFactory(config FactoryConfig) Factory {
	factory, err := NewFactory(config)
	if err != nil {
		panic(err)
	}

	return factory
}

func (factory Factory) Config() FactoryConfig {
	return factory.factoryConfig
}

func (factory Factory) IsZero() bool {
	return factory == Factory{}
}

func (factory Factory) NewAvailableHour(hour time.Time) (*Hour, error) {
	if err := factory.validateTime(hour); err != nil {
		return nil, err
	}

	return &Hour{
		hour:         hour,
		availability: Available,
	}, nil
}

func (factory Factory) NewNotAvailableHour(hour time.Time) (*Hour, error) {
	if err := factory.validateTime(hour); err != nil {
		return nil, err
	}

	return &Hour{
		hour:         hour,
		availability: NotAvailable,
	}, nil
}

// Validate Time
var (
	ErrNotFullHour = errors.New("hour should be a full hour")
	ErrPastHour    = errors.New("cannot create hour from past")
)

// If you have the error with a more complex context,
// it's a good idea to define it as a separate type.
// There is nothing worst, than error "invalid date" without knowing what date was passed and what is the valid value!
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

func (factory Factory) validateTime(hour time.Time) error {
	if !hour.Round(time.Hour).Equal(hour) {
		return ErrNotFullHour
	}

	if hour.After(time.Now().AddDate(0, 0, factory.factoryConfig.MaxWeeksInTheFutureToSet*7)) {
		return TooDistantDateError{
			MaxWeeksInTheFutureToSet: factory.factoryConfig.MaxWeeksInTheFutureToSet,
			ProvidedDate:             hour,
		}
	}

	currentHour := time.Now().Truncate(time.Hour)
	if hour.Before(currentHour) || hour.Equal(currentHour) {
		return ErrPastHour
	}

	if hour.UTC().Hour() > factory.factoryConfig.MaxUtcHour {
		return TooLateHourError{
			MaxUtcHour:   factory.factoryConfig.MaxUtcHour,
			ProvidedTime: hour,
		}
	}

	if hour.UTC().Hour() < factory.factoryConfig.MinUtcHour {
		return TooEarlyHourError{
			MinUtcHour:   factory.factoryConfig.MinUtcHour,
			ProvidedTime: hour,
		}
	}

	return nil
}
