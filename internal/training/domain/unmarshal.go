package domain

import (
	"time"
)

// UnmarshalTrainingFromDB rebuilds a Training from data already persisted in the database.
// It skips the validation performed by NewTraining because data coming from our own
// storage is trusted.
func UnmarshalTrainingFromDB(
	trainingUUID TrainingUUID,
	userUUID UserUUID,
	userName string,
	trainingTime time.Time,
	notes string,
	proposedNewTime time.Time,
	moveProposedBy UserType,
	canceled bool,
) *Training {
	return &Training{
		uuid:            trainingUUID,
		userUUID:        userUUID,
		userName:        userName,
		time:            trainingTime,
		notes:           notes,
		proposedNewTime: proposedNewTime,
		moveProposedBy:  moveProposedBy,
		canceled:        canceled,
	}
}
