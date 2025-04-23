package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LogLevel defines the severity of the log message
type LogLevel string

const (
	// LogLevelDebug is for detailed troubleshooting
	LogLevelDebug LogLevel = "DEBUG"
	// LogLevelInfo is for normal operation
	LogLevelInfo LogLevel = "INFO"
	// LogLevelWarn is for potential issues
	LogLevelWarn LogLevel = "WARN"
	// LogLevelError is for critical issues
	LogLevelError LogLevel = "ERROR"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Level       LogLevel               `json:"level"`
	Timestamp   string                 `json:"timestamp"`
	Service     string                 `json:"service"`
	Function    string                 `json:"function"`
	Message     string                 `json:"message"`
	CorrelationId string               `json:"correlationId,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// Logger interface defines the methods a logger must implement
type Logger interface {
	Debug(message string, details map[string]interface{})
	Info(message string, details map[string]interface{})
	Warn(message string, details map[string]interface{})
	Error(message string, details map[string]interface{})
	WithCorrelationId(correlationId string) Logger
}

// StructuredLogger implements the Logger interface with JSON formatting
type StructuredLogger struct {
	service       string
	function      string
	correlationId string
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger() Logger {
	return &StructuredLogger{
		service:  "kootoro-verification",
		function: "InitializeFunction",
	}
}

// WithCorrelationId returns a new logger with the specified correlation ID
func (l *StructuredLogger) WithCorrelationId(correlationId string) Logger {
	return &StructuredLogger{
		service:       l.service,
		function:      l.function,
		correlationId: correlationId,
	}
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(message string, details map[string]interface{}) {
	l.log(LogLevelDebug, message, details)
}

// Info logs an info message
func (l *StructuredLogger) Info(message string, details map[string]interface{}) {
	l.log(LogLevelInfo, message, details)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(message string, details map[string]interface{}) {
	l.log(LogLevelWarn, message, details)
}

// Error logs an error message
func (l *StructuredLogger) Error(message string, details map[string]interface{}) {
	l.log(LogLevelError, message, details)
}

// log creates and outputs a log entry
func (l *StructuredLogger) log(level LogLevel, message string, details map[string]interface{}) {
	// Create a copy of details to avoid modifying the original
	detailsCopy := make(map[string]interface{})
	if details != nil {
		for k, v := range details {
			detailsCopy[k] = v
		}
	}

	// If verificationId is in details, use it for correlationId if not already set
	if l.correlationId == "" {
		if verificationId, ok := detailsCopy["verificationId"].(string); ok {
			l.correlationId = verificationId
			// Remove from details to avoid duplication
			delete(detailsCopy, "verificationId")
		}
	}

	// Create the log entry
	entry := LogEntry{
		Level:       level,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Service:     l.service,
		Function:    l.function,
		Message:     message,
		Details:     detailsCopy,
	}

	// Add correlation ID if available
	if l.correlationId != "" {
		entry.CorrelationId = l.correlationId
	}

	// Marshal the log entry to JSON
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// If marshaling fails, fall back to a simple format
		fmt.Fprintf(os.Stderr, "ERROR: Failed to marshal log entry: %v\n", err)
		fmt.Fprintf(os.Stderr, "%s [%s] %s: %s\n", entry.Timestamp, entry.Level, entry.Function, entry.Message)
		return
	}

	// Output the JSON log entry
	if level == LogLevelError {
		fmt.Fprintln(os.Stderr, string(jsonBytes))
	} else {
		fmt.Fprintln(os.Stdout, string(jsonBytes))
	}
}