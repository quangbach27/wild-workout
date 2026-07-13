package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
	"workout/common"
)

type TrainingRepository interface {
	AddTraining(ctx context.Context, tr *Training) error

	GetTraining(ctx context.Context, trainingUUID TrainingUUID, user User) (*Training, error)

	UpdateTraining(
		ctx context.Context,
		trainingUUID string,
		user User,
		updateFn func(ctx context.Context, tr *Training) (*Training, error),
	) error
}
type TrainingUUID struct {
	common.UUID
}

type Training struct {
	uuid TrainingUUID

	userUUID string
	userName string

	time  time.Time
	notes string

	proposedNewTime time.Time
	moveProposedBy  UserType

	canceled bool
}

func (t *Training) UUID() TrainingUUID {
	return t.uuid
}

func (t *Training) UserUUID() string {
	return t.userUUID
}

func (t *Training) UserName() string {
	return t.userName
}

func (t *Training) Time() time.Time {
	return t.time
}

func (t *Training) Notes() string {
	return t.notes
}

func (t *Training) ProposedNewTime() time.Time {
	return t.proposedNewTime
}

func (t *Training) MoveProposedBy() UserType {
	return t.moveProposedBy
}

func (t *Training) IsCanceled() bool {
	return t.canceled
}

func NewTraining(userUUID string, userName string, trainingTime time.Time) (*Training, error) {
	var errDetails []common.ErrorDetails

	if userUUID == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "Training",
			ErrorSlug:  "empty-user-uuid",
			Message:    "userUUID cannot be empty",
		})
	}
	if userName == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "Training",
			ErrorSlug:  "empty-user-name",
			Message:    "userName cannot be empty",
		})
	}
	if trainingTime.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "Training",
			ErrorSlug:  "zero-training-time",
			Message:    "training time cannot be zero",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError(
			"invalid-training",
			"training is not valid",
		).WithDetails(errDetails)
	}

	return &Training{
		uuid:     TrainingUUID{common.NewUUIDv7()},
		userUUID: userUUID,
		userName: userName,
		time:     trainingTime,
	}, nil
}

func (t Training) CanBeCanceledForFree() bool {
	return time.Until(t.time) >= time.Hour*24
}

func (t *Training) ProposeReschedule(newTime time.Time, proposerType UserType) {
	t.moveProposedBy = proposerType
	t.proposedNewTime = newTime
}

func (t *Training) RescheduleTraining(newTime time.Time) error {
	if !t.CanBeCanceledForFree() {
		return fmt.Errorf(
			"can't reschedule training, not enough time before, training time: %s",
			t.Time(),
		)
	}

	t.time = newTime

	return nil
}

func (t *Training) ApproveReschedule(userType UserType) error {
	if !t.IsRescheduleProposed() {
		return errors.New("no training reschedule was requested yet")
	}

	if t.moveProposedBy == userType {
		return fmt.Errorf(
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
		return errors.New("no training reschedule was requested yet")
	}

	t.proposedNewTime = time.Time{}
	t.moveProposedBy = UserType{}

	return nil
}

func (t *Training) IsRescheduleProposed() bool {
	return !t.moveProposedBy.IsZero() && !t.proposedNewTime.IsZero()
}

func (t *Training) Cancel() error {
	if t.canceled {
		return errors.New("training is already canceled")
	}

	t.canceled = true
	return nil
}
