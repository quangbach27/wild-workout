package hour

import "errors"

type Availability struct {
	value string
}

func (availability Availability) String() string {
	return availability.value
}

var (
	Available        = Availability{value: "available"}
	NotAvailable     = Availability{value: "not_available"}
	ScheduleTraining = Availability{value: "scheduled"}
)

var (
	ErrTrainingScheduled   = errors.New("unable to modify hour, because scheduled training")
	ErrNoTrainingScheduled = errors.New("training is not scheduled")
	ErrHourNotAvailable    = errors.New("hour is not available")
)

func (availability Availability) IsZero() bool {
	return availability == Availability{}
}

func (hour Hour) IsAvailable() bool {
	return hour.availability == Available
}

func (hour Hour) hasTrainingSchedule() bool {
	return hour.availability == ScheduleTraining
}

func (h *Hour) MakeAvailable() error {
	if h.hasTrainingSchedule() {
		return ErrTrainingScheduled
	}

	h.availability = Available
	return nil
}

func (h *Hour) MakeNotAvailable() error {
	if h.hasTrainingSchedule() {
		return ErrTrainingScheduled
	}

	h.availability = NotAvailable
	return nil
}

func (h *Hour) ScheduleTraining() error {
	if !h.IsAvailable() {
		return ErrHourNotAvailable
	}

	h.availability = ScheduleTraining
	return nil
}

func (h *Hour) CancelTraining() error {
	if !h.hasTrainingSchedule() {
		return ErrNoTrainingScheduled
	}

	h.availability = Available
	return nil
}
