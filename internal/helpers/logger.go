package helpers

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger sets up a global Logrus logger
var Logger = logrus.New()

// LoggerInit sets default logging options through predefined environment variables (currently only one):
// BRUBOT_LOGLEVEL: specifies the default loglevel to run under, defaults to Info if not set.
func LoggerInit() {

	// Logrus supported logging levels, see https://github.com/sirupsen/logrus#level-logging
	supportedLogLevels := map[string]int{
		"TRACE": 6,
		"DEBUG": 5,
		"INFO":  4,
		"WARN":  3,
		"ERROR": 2,
	}

	logLevel, logLevelSet := os.LookupEnv("BRUBOT_LOGLEVEL")

	// Confirm logLevel environment variable is a supported loglevel,
	// if valid, set loglevel from environment variable
	if logLevelSet && supportedLogLevels[logLevel] != 0 {

		Logger.SetLevel(logrus.Level(supportedLogLevels[logLevel]))

	} else {

		// This happens by default but being a bit pedantic is not
		// the end of the world
		Logger.SetLevel(logrus.InfoLevel)

	}

	// Set sane defaults that might need to move into a config file at some point
	Logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	Logger.SetReportCaller(false)
	// 12Factor, baby!
	Logger.SetOutput(os.Stdout)

}
