package domain

import "time"

func UnmarshalHour(
	hour time.Time,
	availability Availability,
) *Hour {
	return &Hour{
		hour:         hour,
		availability: availability,
	}
}
