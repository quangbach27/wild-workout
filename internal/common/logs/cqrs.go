package logs

import "github.com/sirupsen/logrus"

func LogCommandExecution(commandName string, cmd any, err error) {
	log := logrus.WithField("cmd", cmd)

	if err != nil {
		log.WithError(err).Error(commandName + " command failed")
	} else {
		log.Info(commandName + " command succeeded")
	}
}
