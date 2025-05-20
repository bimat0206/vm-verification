package processor

import (
	"log/slog"
	"workflow-function/shared/logger"
)

// SlogLoggerAdapter adapts a *slog.Logger to the shared logger.Logger interface
type SlogLoggerAdapter struct {
	slogger *slog.Logger
}

// NewSlogLoggerAdapter creates a new adapter for *slog.Logger
func NewSlogLoggerAdapter(logger *slog.Logger) logger.Logger {
	return &SlogLoggerAdapter{
		slogger: logger,
	}
}

// Info logs information
func (a *SlogLoggerAdapter) Info(msg string, fields map[string]interface{}) {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	a.slogger.Info(msg, attrs...)
}

// Warn logs a warning
func (a *SlogLoggerAdapter) Warn(msg string, fields map[string]interface{}) {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	a.slogger.Warn(msg, attrs...)
}

// Error logs an error
func (a *SlogLoggerAdapter) Error(msg string, fields map[string]interface{}) {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	a.slogger.Error(msg, attrs...)
}

// Debug logs debug information
func (a *SlogLoggerAdapter) Debug(msg string, fields map[string]interface{}) {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	a.slogger.Debug(msg, attrs...)
}

// GetSlogLogger returns the underlying *slog.Logger
func (a *SlogLoggerAdapter) GetSlogLogger() *slog.Logger {
	return a.slogger
}

// LogOutputEvent logs an output event (implements logger.Logger interface)
func (a *SlogLoggerAdapter) LogOutputEvent(event interface{}) {
	a.slogger.Info("Output event", "event", event)
}

// LogReceivedEvent logs a received event (implements logger.Logger interface)
func (a *SlogLoggerAdapter) LogReceivedEvent(event interface{}) {
	a.slogger.Info("Received event", "event", event)
}

// WithCorrelationId returns a new logger with correlation ID (implements logger.Logger interface)
func (a *SlogLoggerAdapter) WithCorrelationId(correlationId string) logger.Logger {
	return &SlogLoggerAdapter{
		slogger: a.slogger.With("correlationId", correlationId),
	}
}

// WithFields returns a new logger with additional fields (implements logger.Logger interface)
func (a *SlogLoggerAdapter) WithFields(fields map[string]interface{}) logger.Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return &SlogLoggerAdapter{
		slogger: a.slogger.With(attrs...),
	}
}