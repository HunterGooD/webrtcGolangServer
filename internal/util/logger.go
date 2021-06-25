package util

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)

	logLevel, err := log.ParseLevel(os.Getenv("LOG_LVL"))
	if err != nil {
		logLevel = log.InfoLevel
	}

	log.SetLevel(logLevel)
}

func Infof(str string, v ...interface{}) {
	log.Infof(str, v...)
}

func Warnf(str string, v ...interface{}) {
	log.Warnf(str, v...)
}

func Errorf(str string, v ...interface{}) {
	log.Errorf(str, v...)
}

func Debugf(str string, v ...interface{}) {
	log.Debugf(str, v...)
}

func Panicf(str string, v ...interface{}) {
	log.Panicf(str, v...)
}
