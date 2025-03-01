package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger defines methods for logging
type Logger interface {
	Debug() zerolog.Event
	Info() zerolog.Event
	Warn() zerolog.Event
	Error() zerolog.Event
	Fatal() zerolog.Event
	With() zerolog.Context
	Sync() error
}

// ZapLogger implements Logger with Zerolog
type ZeroLogger struct {
	logger zerolog.Logger
}

// NewZapLogger creates a new logger
func NewZapLogger(level string) *ZeroLogger {
	// Set global time format
	zerolog.TimeFieldFormat = zerolog.TimeFormatISO8601

	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Setup console writer
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	// Create logger
	logger := zerolog.New(consoleWriter).
		With().
		Timestamp().
		Logger().
		Level(logLevel)

	return &ZeroLogger{
		logger: logger,
	}
}

// Debug returns a debug level event
func (l *ZeroLogger) Debug() zerolog.Event {
	return l.logger.Debug()
}

// Info returns an info level event
func (l *ZeroLogger) Info() zerolog.Event {
	return l.logger.Info()
}

// Warn returns a warn level event
func (l *ZeroLogger) Warn() zerolog.Event {
	return l.logger.Warn()
}

// Error returns an error level event
func (l *ZeroLogger) Error() zerolog.Event {
	return l.logger.Error()
}

// Fatal returns a fatal level event
func (l *ZeroLogger) Fatal() zerolog.Event {
	return l.logger.Fatal()
}

// With returns a context with fields
func (l *ZeroLogger) With() zerolog.Context {
	return l.logger.With()
}

// Sync flushes any buffered logs
func (l *ZeroLogger) Sync() error {
	// Zerolog doesn't require explicit synchronization
	return nil
}
