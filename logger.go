package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// DefaultLogger is the main logger instance used by this package
	DefaultLogger zerolog.Logger

	initOnce sync.Once

	defaultConfig = Config{
		Level:      "info",
		Pretty:     false,
		WithCaller: false,
		TimeFormat: time.RFC3339,
		Output:     os.Stderr,
	}
)

// Config defines configuration options for the logger
type Config struct {
	Level      string    // Log level: debug, info, warn, error, fatal, panic
	Pretty     bool      // Enable pretty (human-readable) logging
	WithCaller bool      // Include caller information in logs
	TimeFormat string    // Timestamp format
	Output     io.Writer // Output writer (defaults to stderr)
}

// Standard log levels mapped to zerolog levels
var Levels = map[string]zerolog.Level{
	"debug":    zerolog.DebugLevel,
	"info":     zerolog.InfoLevel,
	"warn":     zerolog.WarnLevel,
	"error":    zerolog.ErrorLevel,
	"fatal":    zerolog.FatalLevel,
	"panic":    zerolog.PanicLevel,
	"disabled": zerolog.Disabled,
}

// InitLogger initializes the global logger with the given configuration
// This configuration applies to ALL packages that use this logger,
// including libraries that import this package.
// This method is safe to call multiple times, but only the first call
// will take effect to prevent configuration conflicts.
func InitLogger(cfg Config) {
	initOnce.Do(func() {
		// Fill in defaults for any missing config values
		if cfg.Output == nil {
			cfg.Output = defaultConfig.Output
		}
		if cfg.TimeFormat == "" {
			cfg.TimeFormat = defaultConfig.TimeFormat
		}

		// Set global time format for all loggers
		zerolog.TimeFieldFormat = cfg.TimeFormat

		// Set global log level - this affects ALL zerolog instances
		level := zerolog.InfoLevel
		if lvl, ok := Levels[strings.ToLower(cfg.Level)]; ok {
			level = lvl
		}
		zerolog.SetGlobalLevel(level)

		// Create and configure the logger
		var logger zerolog.Logger
		if cfg.Pretty {
			logger = zerolog.New(zerolog.ConsoleWriter{
				Out:        cfg.Output,
				TimeFormat: cfg.TimeFormat,
			})
		} else {
			logger = zerolog.New(cfg.Output)
		}

		// Add timestamp to all logs
		logger = logger.With().Timestamp().Logger()

		// Optionally add caller info
		if cfg.WithCaller {
			logger = logger.With().Caller().Logger()
		}

		// Set both our package-level DefaultLogger and zerolog's global logger
		// This ensures ALL code using either one will get the same configuration
		DefaultLogger = logger
		log.Logger = logger
	})
}

// GetLogger returns a contextualized logger with the component field set
// This is useful for identifying which module generated a log entry
func GetLogger(component string) zerolog.Logger {

	return DefaultLogger.With().Str("component", component).Logger()
}

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	if len(args) > 0 && strings.Contains(msg, "%") {
		DefaultLogger.Debug().Msgf(msg, args...)
	} else if len(args) > 0 {
		evt := DefaultLogger.Debug()
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				evt = evt.Interface(fmt.Sprint(args[i]), args[i+1])
			}
		}
		evt.Msg(msg)
	} else {
		DefaultLogger.Debug().Msg(msg)
	}
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	if len(args) > 0 && strings.Contains(msg, "%") {
		DefaultLogger.Info().Msgf(msg, args...)
	} else if len(args) > 0 {
		evt := DefaultLogger.Info()
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				evt = evt.Interface(fmt.Sprint(args[i]), args[i+1])
			}
		}
		evt.Msg(msg)
	} else {
		DefaultLogger.Info().Msg(msg)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	if len(args) > 0 && strings.Contains(msg, "%") {
		DefaultLogger.Warn().Msgf(msg, args...)
	} else if len(args) > 0 {
		evt := DefaultLogger.Warn()
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				evt = evt.Interface(fmt.Sprint(args[i]), args[i+1])
			}
		}
		evt.Msg(msg)
	} else {
		DefaultLogger.Warn().Msg(msg)
	}
}

// Error logs an error message
func Error(err error, msg string, args ...interface{}) {
	event := DefaultLogger.Error().Err(err)
	if len(args) > 0 && strings.Contains(msg, "%") {
		event.Msgf(msg, args...)
	} else if len(args) > 0 {
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				event = event.Interface(fmt.Sprint(args[i]), args[i+1])
			}
		}
		event.Msg(msg)
	} else {
		event.Msg(msg)
	}
}

// Fatal logs a fatal message and exits
func Fatal(err error, msg string, args ...interface{}) {
	event := DefaultLogger.Fatal().Err(err)
	if len(args) > 0 && strings.Contains(msg, "%") {
		event.Msgf(msg, args...)
	} else if len(args) > 0 {
		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				event = event.Interface(fmt.Sprint(args[i]), args[i+1])
			}
		}
		event.Msg(msg)
	} else {
		event.Msg(msg)
	}
}

// WithField adds a field to the logger context
func WithField(key string, value interface{}) zerolog.Logger {
	return DefaultLogger.With().Interface(key, value).Logger()
}

// FormatError creates a formatted error string
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%v", err)
}

func init() {
	// Initialize with default configuration
	// This ensures logger works before explicit initialization
	// The first call to InitLogger will override these settings
	DefaultLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	log.Logger = DefaultLogger
}
