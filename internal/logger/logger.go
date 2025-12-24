package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global logger based on configuration
func Init(level string, format string) {
	// Set log level
	zerolog.SetGlobalLevel(parseLevel(level))

	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Set output format
	if format == "json" {
		// JSON format (production)
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		// Console format (development)
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}
}

func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Get returns the global logger instance
func Get() zerolog.Logger {
	return log.Logger
}
