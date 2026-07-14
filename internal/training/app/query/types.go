package query

import (
	"time"
	"workout/training/domain"
)

type Training struct {
	UUID   domain.TrainingUUID
	UserID domain.UserID
	User   string

	Time  time.Time
	Notes string

	ProposedTime   *time.Time
	MoveProposedBy *string

	CanBeCancelled bool
}
