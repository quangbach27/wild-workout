package domain

import (
	"time"
)

// UnmarshalTrainingFromDB rebuilds a Training from data already persisted in the database.
// It skips the validation performed by NewTraining because data coming from our own
// storage is trusted.
func UnmarshalTrainingFromDB(
	trainingUUID TrainingUUID,
	userID UserID,
	userName string,
	trainingTime time.Time,
	notes string,
	proposedNewTime time.Time,
	moveProposedBy UserType,
	canceled bool,
) *Training {
	return &Training{
		uuid:            trainingUUID,
		userID:          userID,
		userName:        userName,
		time:            trainingTime,
		notes:           notes,
		proposedNewTime: proposedNewTime,
		moveProposedBy:  moveProposedBy,
		canceled:        canceled,
	}
}
