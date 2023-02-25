// Package log centralizes logging for the app.
package log

import (
	"os"

	golog "git.kirsle.net/go/log"
)

var log golog.Logger

func init() {
	log = *golog.GetLogger("main")
	log.Configure(&golog.Config{
		Colors: golog.ExtendedColor,
		Theme:  golog.DarkTheme,
	})

	log.Config.Level = golog.InfoLevel
}

// SetDebug toggles debug level logging.
func SetDebug(v bool) {
	if v {
		log.Config.Level = golog.DebugLevel
	} else {
		log.Config.Level = golog.InfoLevel
	}
}

// Info log.
func Info(message string, v ...interface{}) {
	log.Info(message, v...)
}

// Debug log.
func Debug(message string, v ...interface{}) {
	log.Debug(message, v...)
}

// Warn log.
func Warn(message string, v ...interface{}) {
	log.Warn(message, v...)
}

// Error log.
func Error(message string, v ...interface{}) {
	log.Error(message, v...)
}

// Fatal logs an error and exits.
func Fatal(message string, v ...interface{}) {
	log.Error(message, v...)
	os.Exit(1)
}
