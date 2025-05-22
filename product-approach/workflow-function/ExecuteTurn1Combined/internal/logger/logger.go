// internal/utils/logger.go
package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Logger is a lightweight wrapper around slog.Logger that adds a slightly
// friendlier chainable API used throughout the ExecuteTurn1Combined code-base.
type Logger struct {
	base *slog.Logger
}

// singleton slog handler initialised once per Lambda cold-start.
var (
	once   sync.Once
	root   *slog.Logger
)

// New returns a component-scoped structured logger.
// The first call configures the global JSON handler with level picked from
// LOG_LEVEL (default INFO); subsequent calls re-use the same handler.
func New(component string) *Logger {
	once.Do(func() {
		level := parseLevel(os.Getenv("LOG_LEVEL"))
		root = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		}))
	})

	return &Logger{base: root.With("component", component)}
}

// WithContext attaches the context to the logger (kept for symmetry; slog
// already supports ctx, but we return *Logger for fluent chaining).
func (l *Logger) WithContext(_ context.Context) *Logger { return l }

// WithFields appends arbitrary structured key-value pairs.
func (l *Logger) WithFields(kv ...interface{}) *Logger {
	return &Logger{base: l.base.With(kv...)}
}

func (l *Logger) Debug(msg string, kv ...interface{}) { l.base.Debug(msg, kv...) }
func (l *Logger) Info(msg string, kv ...interface{})  { l.base.Info(msg, kv...) }
func (l *Logger) Warn(msg string, kv ...interface{})  { l.base.Warn(msg, kv...) }
func (l *Logger) Error(msg string, kv ...interface{}) { l.base.Error(msg, kv...) }

// parseLevel converts a LOG_LEVEL string into slog.Level.
func parseLevel(lvl string) slog.Level {
	switch strings.ToUpper(lvl) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
