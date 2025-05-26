package training

import (
	"errors"
	"time"
)

func (training Training) CanBeCanceledForFree() bool {
	return training.time.Sub(time.Now()) >= time.Hour*24
}

var ErrTrainingAlreadyCanceled = errors.New("training is already canceled")

func (training *Training) Cancel() error {
	if training.IsCanceled() {
		return ErrTrainingAlreadyCanceled
	}

	training.canceled = true
	return nil
}

func (traing Training) IsCanceled() bool {
	return traing.canceled
}
