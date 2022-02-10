package logger

import (
	"log"
	"os"
	"strings"
)

type LoggerInterface interface {
	Printf(format string, v ...interface{})

	Fatalf(format string, v ...interface{})
}

type silentLogger struct {
}

func (silentLogger *silentLogger) Printf(format string, v ...interface{}) {

}

func (silentLogger *silentLogger) Fatalf(format string, v ...interface{}) {
	os.Exit(1)
}

var logger LoggerInterface

func init() {
	logLevelValue, ok := os.LookupEnv("LOG_LEVEL")
	if ok && strings.ToUpper(logLevelValue) == "DEBUG" {
		logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	} else {
		logger = &silentLogger{}
	}
}

func GetLogger() LoggerInterface {
	return logger
}
