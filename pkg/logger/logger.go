package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// Logger is a wrapper around zerolog.Logger
type Logger interface {
	Debug() *zerolog.Event // Changed return type to pointer
	Info() *zerolog.Event  // Changed return type to pointer
	Warn() *zerolog.Event  // Changed return type to pointer
	Error() *zerolog.Event // Changed return type to pointer
	Fatal() *zerolog.Event // Changed return type to pointer
}

// zerologLogger is an implementation of Logger using zerolog
type zerologLogger struct {
	logger zerolog.Logger
}

// NewLogger creates a new logger instance
func NewLogger(level string) Logger {
	// Fixed the constant usage
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix // Changed to an existing constant

	// Set log level
	var logLevel zerolog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	return &zerologLogger{
		logger: zerolog.New(os.Stdout).
			Level(logLevel).
			With().
			Timestamp().
			Logger(),
	}
}

// Debug returns a debug event logger
func (l *zerologLogger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info returns an info event logger
func (l *zerologLogger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Warn returns a warn event logger
func (l *zerologLogger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error returns an error event logger
func (l *zerologLogger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Fatal returns a fatal event logger
func (l *zerologLogger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}
