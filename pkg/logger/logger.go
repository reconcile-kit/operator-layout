package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	LevelDebug Level = 5
	LevelInfo  Level = 4
	LevelWarn  Level = 3
	LevelError Level = 2
)

type Logger struct {
	*logrus.Logger
}

type Level int32

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("%s [%s] %s\n", timestamp, entry.Level.String(), entry.Message)
	return []byte(logMessage), nil
}

func New(level Level) *Logger {
	logger := &Logger{Logger: logrus.New()}
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&CustomFormatter{})
	logger.SetLevel(logrus.Level(level))
	return logger
}
