package log

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	logger *logrus.Logger
}

func NewLogger() *Logger {
	return &Logger{
		logger: logrus.New(),
	}
}

func (l *Logger) Init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.SetOutput(os.Stdout)

	logrus.SetLevel(logrus.InfoLevel)
}

func (l *Logger) Log(msg string) {
	l.logger.Info(msg)
}

func (l *Logger) LogDecision(key string, remaining int,
	allowed bool, reset time.Time) {
	l.logger.Info(fmt.Sprintf("key=%s, allowed=%t, count=%d, reset_at=%v",
		key, allowed, remaining, reset))
}

func (l *Logger) Error(msg string, err error) {
	l.logger.Error(msg, err)
}
