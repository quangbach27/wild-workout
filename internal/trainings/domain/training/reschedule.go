package training

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

func (training Training) MovedPropesedBy() UserType {
	return training.moveProposedBy
}

func (training Training) ProposedNewTime() time.Time {
	return training.proposedNewTime
}

type CannotRescheduleBeforeTimeError struct {
	TrainingTime time.Time
}

func (error CannotRescheduleBeforeTimeError) Error() string {
	return fmt.Sprintf(
		"cannot reschedule training, not enough time before, training time: %s",
		error.TrainingTime,
	)
}

func (t *Training) RescheduleTraining(newTime time.Time) error {
	if !t.CanBeCanceledForFree() {
		err := CannotRescheduleBeforeTimeError{
			TrainingTime: t.Time(),
		}
		return errors.WithStack(err)
	}
	t.time = newTime

	return nil
}

func (t *Training) ProposeReschedule(newTime time.Time, proposerType UserType) {
	t.moveProposedBy = proposerType
	t.proposedNewTime = newTime
}

func (t *Training) IsRescheduleProposed() bool {
	return !t.moveProposedBy.IsZero() && !t.proposedNewTime.IsZero()
}

var ErrNoRescheduleRequested = errors.New("no training reschedule was requested yet")

func (t *Training) ApproveReschedule(userType UserType) error {
	if !t.IsRescheduleProposed() {
		return errors.WithStack(ErrNoRescheduleRequested)
	}

	if t.moveProposedBy == userType {
		return errors.Errorf(
			"trying to approve reschedule by the same user type which proposed reschedule (%s)",
			userType.String(),
		)
	}

	t.time = t.proposedNewTime

	t.proposedNewTime = time.Time{}
	t.moveProposedBy = UserType{}

	return nil
}

func (t *Training) RejectReschedule() error {
	if !t.IsRescheduleProposed() {
		return errors.WithStack(ErrNoRescheduleRequested)
	}

	t.proposedNewTime = time.Time{}
	t.moveProposedBy = UserType{}

	return nil
}
