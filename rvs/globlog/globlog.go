package globlog

import (
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

var Env = "prod"

// Centralized logging function that can be used by any package.

func init() {
	Logger = logrus.New()
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// if the env is dev, we will make the loglevel to debug mode so that we can see the logs in the console
	if Env == "dev" {
		Logger.SetLevel(logrus.DebugLevel)
	} else if Env == "prod" {
		Logger.SetLevel(logrus.InfoLevel)
	}
}

func Log() *logrus.Logger {
	return Logger
}
