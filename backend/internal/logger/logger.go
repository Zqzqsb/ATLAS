// Package logger provides structured logging via Go's standard library log/slog.
// Initialize once at startup with logger.Init(), then use logger.L() everywhere.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

var globalLogger *slog.Logger

// Init initializes the global structured logger.
// level: "debug", "info", "warn", "error" (default "info")
// format: "json" or "text" (default "text")
func Init(level, format string) {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: lvl}

	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	globalLogger = slog.New(handler)
	slog.SetDefault(globalLogger)
}

// L returns the global logger. Safe to call before Init() — returns slog.Default().
func L() *slog.Logger {
	if globalLogger == nil {
		return slog.Default()
	}
	return globalLogger
}

// With returns a child logger with additional default attributes.
// Usage: log := logger.With("component", "react_engine")
func With(args ...any) *slog.Logger {
	return L().With(args...)
}

// Writer returns an io.Writer that writes to the logger at Info level.
// Useful for bridging legacy code that expects an io.Writer.
func Writer() io.Writer {
	return &logWriter{logger: L()}
}

type logWriter struct {
	logger *slog.Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logger.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}
