package app

import "github.com/quangbach27/wild-workout/internal/trainer/app/query"

type Application struct {
	Queries Queries
}

type Queries struct {
	AvailableHoursHandler query.AvailableHoursHandler
}
