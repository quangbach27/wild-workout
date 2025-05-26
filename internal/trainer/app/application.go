package app

import (
	"github.com/quangbach27/wild-workout/internal/trainer/app/command"
	"github.com/quangbach27/wild-workout/internal/trainer/app/query"
)

type Application struct {
	Queries Queries
	Commands Commands
}

type Queries struct {
	TrainerAvailableHours query.AvailableHoursHandler
	HourAvailability      query.HourAvailabilityHandler
}

type Commands struct {
	CancelTraining   command.CancelTrainingHandler
	ScheduleTraining command.ScheduleTrainingHandler

	MakeHoursAvailable   command.MakeHoursAvailableHandler
	MakeHoursUnavailable command.MakeHoursUnavailableHandler
}
