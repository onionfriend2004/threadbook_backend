package logger

import (
	"strings"

	"github.com/onionfriend2004/threadbook_backend/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// parseLogLevel converts a string level to zapcore.Level.
// Defaults to InfoLevel if unknown.
func parseLogLevel(levelStr string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// New creates a new zap.Logger based on the provided config.
// Uses human-readable format in debug mode, JSON in production.
func New(cfg *config.Config) *zap.Logger {
	level := parseLogLevel(cfg.Log.Level)

	var logger *zap.Logger
	var err error

	if level == zapcore.DebugLevel {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	zap.ReplaceGlobals(logger)
	return logger
}
