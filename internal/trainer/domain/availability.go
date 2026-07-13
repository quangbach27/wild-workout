package domain

import "workout/common"

type Availability struct {
	common.Enum[AvailabilityTypes]
}

type AvailabilityTypes string

func (at AvailabilityTypes) Values() []string {
	return []string{"available", "not_available", "training_scheduled"}
}

var (
	Available         = common.MustEnum[Availability]("available")
	NotAvailable      = common.MustEnum[Availability]("not_available")
	TrainingScheduled = common.MustEnum[Availability]("training_scheduled")
)
