package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Logger zerolog.Logger
)

// InitLogging initializes the logging system based on configuration
func InitLogging(cfg *Config) error {
	// Determine output writer
	var output io.Writer
	if cfg.LogFile != "" {
		// Open log file
		file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	} else {
		output = os.Stderr
	}

	// Configure output format
	if cfg.LogFormat == "text" {
		// Console writer with color support
		if cfg.ColorOutput && isTerminal(output) {
			consoleWriter := zerolog.ConsoleWriter{
				Out:        output,
				TimeFormat: time.RFC3339,
				NoColor:    false,
			}
			if cfg.ShowTimestamp {
				consoleWriter.TimeFormat = "15:04:05"
			} else {
				consoleWriter.FormatTimestamp = func(i interface{}) string {
					return ""
				}
			}
			output = consoleWriter
		}
	}

	// Create logger
	Logger = zerolog.New(output).With().Timestamp().Logger()

	// Set log level
	level, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	zerolog.SetGlobalLevel(level)

	// Set global logger
	log.Logger = Logger

	return nil
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(level string) (zerolog.Level, error) {
	switch level {
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

// isTerminal checks if the writer is a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// Structured logging helpers

// LogInfo logs an informational message
func LogInfo(msg string) {
	Logger.Info().Msg(msg)
}

// LogDebug logs a debug message
func LogDebug(msg string) {
	Logger.Debug().Msg(msg)
}

// LogWarn logs a warning message
func LogWarn(msg string) {
	Logger.Warn().Msg(msg)
}

// LogError logs an error message
func LogError(msg string, err error) {
	Logger.Error().Err(err).Msg(msg)
}

// LogFatal logs a fatal error and exits
func LogFatal(msg string, err error) {
	Logger.Fatal().Err(err).Msg(msg)
}

// LogInfof logs a formatted informational message
func LogInfof(format string, args ...interface{}) {
	Logger.Info().Msgf(format, args...)
}

// LogDebugf logs a formatted debug message
func LogDebugf(format string, args ...interface{}) {
	Logger.Debug().Msgf(format, args...)
}

// LogWarnf logs a formatted warning message
func LogWarnf(format string, args ...interface{}) {
	Logger.Warn().Msgf(format, args...)
}

// LogErrorf logs a formatted error message
func LogErrorf(format string, args ...interface{}) {
	Logger.Error().Msgf(format, args...)
}

// Security event logging

// LogSecurityEvent logs a security-related event
func LogSecurityEvent(event string, details map[string]interface{}) {
	e := Logger.Warn().Str("event_type", "security")
	for k, v := range details {
		e = e.Interface(k, v)
	}
	e.Msg(event)
}

// LogFileAccess logs file access events (for audit trail)
func LogFileAccess(operation string, path string, success bool) {
	Logger.Info().
		Str("event_type", "file_access").
		Str("operation", operation).
		Str("path", path).
		Bool("success", success).
		Msg("File access")
}

// LogSchemaValidation logs schema validation events
func LogSchemaValidation(schemaPath string, valid bool, errors []error) {
	e := Logger.Info().
		Str("event_type", "schema_validation").
		Str("schema_path", schemaPath).
		Bool("valid", valid)

	if !valid && len(errors) > 0 {
		errorMsgs := make([]string, len(errors))
		for i, err := range errors {
			errorMsgs[i] = err.Error()
		}
		e = e.Strs("errors", errorMsgs)
	}

	e.Msg("Schema validation")
}

// LogDataGeneration logs data generation events
func LogDataGeneration(tableName string, rowCount int, duration time.Duration) {
	Logger.Info().
		Str("event_type", "data_generation").
		Str("table", tableName).
		Int("rows", rowCount).
		Dur("duration", duration).
		Msg("Data generation complete")
}

// LogPipelineStart logs the start of a generation pipeline
func LogPipelineStart(schemaPath string, outputFormat string) {
	Logger.Info().
		Str("event_type", "pipeline_start").
		Str("schema", schemaPath).
		Str("format", outputFormat).
		Msg("Starting data generation pipeline")
}

// LogPipelineComplete logs the completion of a generation pipeline
func LogPipelineComplete(totalRows int, totalTables int, duration time.Duration) {
	Logger.Info().
		Str("event_type", "pipeline_complete").
		Int("total_rows", totalRows).
		Int("total_tables", totalTables).
		Dur("duration", duration).
		Msg("Data generation pipeline complete")
}

// LogPipelineError logs a pipeline error
func LogPipelineError(stage string, err error) {
	Logger.Error().
		Str("event_type", "pipeline_error").
		Str("stage", stage).
		Err(err).
		Msg("Pipeline error")
}

// Performance logging

// LogPerformance logs performance metrics
func LogPerformance(operation string, duration time.Duration, metrics map[string]interface{}) {
	e := Logger.Debug().
		Str("event_type", "performance").
		Str("operation", operation).
		Dur("duration", duration)

	for k, v := range metrics {
		e = e.Interface(k, v)
	}

	e.Msg("Performance metrics")
}

// LogMemoryUsage logs memory usage statistics
func LogMemoryUsage(stage string, allocMB float64, sysMB float64) {
	Logger.Debug().
		Str("event_type", "memory_usage").
		Str("stage", stage).
		Float64("alloc_mb", allocMB).
		Float64("sys_mb", sysMB).
		Msg("Memory usage")
}

// Configuration logging

// LogConfigLoad logs configuration load events
func LogConfigLoad(source string, success bool) {
	Logger.Info().
		Str("event_type", "config_load").
		Str("source", source).
		Bool("success", success).
		Msg("Configuration loaded")
}

// LogConfigValue logs configuration value changes (for debugging)
func LogConfigValue(key string, value interface{}) {
	Logger.Debug().
		Str("event_type", "config_value").
		Str("key", key).
		Interface("value", value).
		Msg("Configuration value")
}

// Error context helpers

// NewErrorContext creates a new error context for structured error logging
func NewErrorContext() *zerolog.Event {
	return Logger.Error()
}

// NewWarnContext creates a new warning context for structured warning logging
func NewWarnContext() *zerolog.Event {
	return Logger.Warn()
}

// NewInfoContext creates a new info context for structured info logging
func NewInfoContext() *zerolog.Event {
	return Logger.Info()
}

// NewDebugContext creates a new debug context for structured debug logging
func NewDebugContext() *zerolog.Event {
	return Logger.Debug()
}
