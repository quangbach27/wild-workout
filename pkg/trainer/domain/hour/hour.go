package hour

import (
	"errors"
	"time"
)

type Hour struct {
	hour         time.Time
	availability Availability
}

func (h Hour) Time() time.Time {
	return h.hour
}

func (h Hour) Availability() Availability {
	return h.availability
}

func (f Factory) UnmarshalHourFromDatabase(hour time.Time, availability Availability) (*Hour, error) {
	if err := f.validateTime(hour); err != nil {
		return nil, err
	}

	if availability.IsZero() {
		return nil, errors.New("empty availability")
	}

	return &Hour{
		hour:         hour,
		availability: availability,
	}, nil
}
