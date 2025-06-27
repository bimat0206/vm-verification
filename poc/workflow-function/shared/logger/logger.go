// Package logger provides a standardized logging interface for Lambda functions
package logger

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
	Level         LogLevel               `json:"level"`
	Timestamp     string                 `json:"timestamp"`
	Service       string                 `json:"service"`
	Function      string                 `json:"function"`
	Message       string                 `json:"message"`
	CorrelationId string                 `json:"correlationId,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// Logger interface defines the methods a logger must implement
type Logger interface {
	Debug(message string, details map[string]interface{})
	Info(message string, details map[string]interface{})
	Warn(message string, details map[string]interface{})
	Error(message string, details map[string]interface{})
	LogReceivedEvent(event interface{}) 
	LogOutputEvent(event interface{})
	WithCorrelationId(correlationId string) Logger
	WithFields(fields map[string]interface{}) Logger
}

// StructuredLogger implements the Logger interface with JSON formatting
type StructuredLogger struct {
	service       string
	function      string
	correlationId string
	fields        map[string]interface{}
}

// New creates a new structured logger with the specified service and function names
func New(service, function string) Logger {
	return &StructuredLogger{
		service:  service,
		function: function,
		fields:   make(map[string]interface{}),
	}
}

// WithCorrelationId returns a new logger with the specified correlation ID
func (l *StructuredLogger) WithCorrelationId(correlationId string) Logger {
	newLogger := &StructuredLogger{
		service:       l.service,
		function:      l.function,
		correlationId: correlationId,
		fields:        copyMap(l.fields),
	}
	return newLogger
}

// WithFields returns a new logger with additional fields
func (l *StructuredLogger) WithFields(fields map[string]interface{}) Logger {
	newFields := copyMap(l.fields)
	for k, v := range fields {
		newFields[k] = v
	}
	
	return &StructuredLogger{
		service:       l.service,
		function:      l.function,
		correlationId: l.correlationId,
		fields:        newFields,
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

// LogReceivedEvent logs an incoming event at INFO level
func (l *StructuredLogger) LogReceivedEvent(event interface{}) {
	var details map[string]interface{}
	
	// If we can marshal the event to JSON, include it in the details
	// Otherwise, use a simpler string representation
	eventJSON, err := json.Marshal(event)
	if err == nil {
		details = map[string]interface{}{
			"event": json.RawMessage(eventJSON),
		}
	} else {
		details = map[string]interface{}{
			"event": fmt.Sprintf("%+v", event),
		}
	}
	
	l.log(LogLevelInfo, "Received event", details)
}

// LogOutputEvent logs an outgoing event at INFO level
func (l *StructuredLogger) LogOutputEvent(event interface{}) {
	var details map[string]interface{}
	
	// If we can marshal the event to JSON, include it in the details
	// Otherwise, use a simpler string representation
	eventJSON, err := json.Marshal(event)
	if err == nil {
		details = map[string]interface{}{
			"response": json.RawMessage(eventJSON),
		}
	} else {
		details = map[string]interface{}{
			"response": fmt.Sprintf("%+v", event),
		}
	}
	
	l.log(LogLevelInfo, "Output event", details)
}

// log creates and outputs a log entry
func (l *StructuredLogger) log(level LogLevel, message string, details map[string]interface{}) {
	// Create a copy of details to avoid modifying the original
	detailsCopy := copyMap(l.fields)
	if details != nil {
		for k, v := range details {
			detailsCopy[k] = v
		}
	}

	// If verificationId is in details, use it for correlationId if not already set
	correlationId := l.correlationId
	if correlationId == "" {
		if verificationId, ok := detailsCopy["verificationId"].(string); ok && verificationId != "" {
			correlationId = verificationId
			// Remove from details to avoid duplication
			delete(detailsCopy, "verificationId")
		} else if requestId, ok := detailsCopy["requestId"].(string); ok && requestId != "" {
			correlationId = requestId
			// Do not remove requestId as it's still valuable context
		}
	}

	// Create the log entry
	entry := LogEntry{
		Level:     level,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   l.service,
		Function:  l.function,
		Message:   message,
		Details:   detailsCopy,
	}

	// Add correlation ID if available
	if correlationId != "" {
		entry.CorrelationId = correlationId
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

// copyMap creates a deep copy of a map
func copyMap(original map[string]interface{}) map[string]interface{} {
	if original == nil {
		return make(map[string]interface{})
	}
	
	copy := make(map[string]interface{}, len(original))
	for k, v := range original {
		copy[k] = v
	}
	return copy
}