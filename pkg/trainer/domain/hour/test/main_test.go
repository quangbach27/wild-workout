package hour_test

import (
	"time"

	"github.com/quangbach27/wild-workout/pkg/trainer/domain/hour"
)

var testHourFactory = hour.MustNewFactory(hour.FactoryConfig{
	MaxWeeksInTheFutureToSet: 100,
	MinUtcHour:               0,
	MaxUtcHour:               24,
})

func validTrainingHour() time.Time {
	tomorrow := time.Now().Add(time.Hour * 24)

	return time.Date(
		tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		testHourFactory.Config().MinUtcHour, 0, 0, 0,
		time.UTC,
	)
}

func trainingHourWithMinutes(minute int) time.Time {
	tomorrow := time.Now().Add(time.Hour * 24)

	return time.Date(
		tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		testHourFactory.Config().MaxUtcHour, minute, 0, 0,
		time.UTC,
	)
}
