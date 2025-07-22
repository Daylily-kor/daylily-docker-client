package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var globalLogger zerolog.Logger

// Init initializes the global logger with pretty console output
func Init() {
	// Configure pretty console output
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Kitchen, // "15:04PM"
	}

	// Set up global logger
	globalLogger = zerolog.New(output).With().Timestamp().Logger()

	// Set as default zerolog logger
	log.Logger = globalLogger
}

// Get returns the global logger
func Get() *zerolog.Logger {
	return &globalLogger
}

// Info logs an info message with key-value pairs
func Info(msg string, fields ...any) {
	logWithFields(globalLogger.Info(), msg, fields...)
}

// Debug logs a debug message with key-value pairs
func Debug(msg string, fields ...any) {
	logWithFields(globalLogger.Debug(), msg, fields...)
}

// Warn logs a warning message with key-value pairs
func Warn(msg string, fields ...any) {
	logWithFields(globalLogger.Warn(), msg, fields...)
}

// Error logs an error message with key-value pairs
func Error(msg string, fields ...any) {
	logWithFields(globalLogger.Error(), msg, fields...)
}

// Fatal logs an error message and exits with key-value pairs
func Fatal(msg string, fields ...any) {
	logWithFields(globalLogger.Fatal(), msg, fields...)
}

// logWithFields is a helper function that adds fields to an event and logs the message
func logWithFields(event *zerolog.Event, msg string, fields ...any) {
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if ok {
				event = event.Interface(key, fields[i+1])
			}
		}
	}
	event.Msg(msg)
}

// Helper functions for structured logging

// WithError returns the error as a string for logging
func WithError(err error) string {
	return err.Error()
}

// WithDuration returns the duration as a string for logging
func WithDuration(d time.Duration) string {
	return d.String()
}
