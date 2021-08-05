package logging

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger provides closest to stdlib log.Logger interface
type Logger interface {
	logrus.FieldLogger
}

// Fields type, used to pass to `WithFields`.
type Fields map[string]interface{}

// StructuredLogger provides the logger interface using Sirupsen/logrus
type StructuredLogger struct {
	*logrus.Logger
}

// NewLogger creates StructuredLogger and configure it
func NewLogger(format, level string, output io.Writer) *StructuredLogger {
	logger := logrus.New()
	logger.Out = output
	formatter := strings.ToLower(format)

	switch formatter {
	case "json":
		logger.Formatter = &logrus.JSONFormatter{}
	case "text":
		logger.Formatter = &logrus.TextFormatter{ForceColors: false}
	default:
		logger.Warnf("log: invalid formatter: %v, continue with default", formatter)
	}

	level = strings.ToLower(level)
	levelParsed, err := logrus.ParseLevel(level)

	if err != nil {
		logger.Warnf("log: invalid level: %v, continue with info", level)
		levelParsed = logrus.InfoLevel
	}

	logger.Level = levelParsed

	return &StructuredLogger{logger}
}

func NewLoggerDefault() *StructuredLogger {
	return NewLogger("json", "error", os.Stdout)
}
