package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger(level string) *logrus.Logger {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)
	Log.SetFormatter(&logrus.JSONFormatter{})

	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		parsedLevel = logrus.InfoLevel
	}
	Log.SetLevel(parsedLevel)

	return Log
}

func GetLogger() *logrus.Logger {
	if Log == nil {
		return InitLogger("info")
	}
	return Log
}
