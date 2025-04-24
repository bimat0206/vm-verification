package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger defines the interface for logging
type Logger interface {
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
	Fatal(message string, args ...interface{})
}

// LogLevel represents the log level
type LogLevel int

const (
	// DebugLevel is the lowest log level
	DebugLevel LogLevel = iota
	// InfoLevel is the default log level
	InfoLevel
	// WarnLevel is for warning messages
	WarnLevel
	// ErrorLevel is for error messages
	ErrorLevel
	// FatalLevel is the highest log level
	FatalLevel
)

// StandardLogger implements the Logger interface
type StandardLogger struct {
	logger *log.Logger
	level  LogLevel
}

// NewLogger creates a new logger
func NewLogger() *StandardLogger {
	// Get log level from environment
	level := InfoLevel
	if levelStr, exists := os.LookupEnv("LOG_LEVEL"); exists {
		switch levelStr {
		case "DEBUG":
			level = DebugLevel
		case "INFO":
			level = InfoLevel
		case "WARN":
			level = WarnLevel
		case "ERROR":
			level = ErrorLevel
		case "FATAL":
			level = FatalLevel
		}
	}
	
	// Create logger
	return &StandardLogger{
		logger: log.New(os.Stdout, "", 0),
		level:  level,
	}
}

// Debug logs a debug message
func (l *StandardLogger) Debug(message string, args ...interface{}) {
	if l.level <= DebugLevel {
		l.log("DEBUG", message, args...)
	}
}

// Info logs an info message
func (l *StandardLogger) Info(message string, args ...interface{}) {
	if l.level <= InfoLevel {
		l.log("INFO", message, args...)
	}
}

// Warn logs a warning message
func (l *StandardLogger) Warn(message string, args ...interface{}) {
	if l.level <= WarnLevel {
		l.log("WARN", message, args...)
	}
}

// Error logs an error message
func (l *StandardLogger) Error(message string, args ...interface{}) {
	if l.level <= ErrorLevel {
		l.log("ERROR", message, args...)
	}
}

// Fatal logs a fatal message and exits
func (l *StandardLogger) Fatal(message string, args ...interface{}) {
	if l.level <= FatalLevel {
		l.log("FATAL", message, args...)
		os.Exit(1)
	}
}

// log formats and logs a message
func (l *StandardLogger) log(level, message string, args ...interface{}) {
	// Format message if arguments are provided
	if len(args) > 0 {
		if err, ok := args[0].(error); ok && len(args) == 1 {
			// Special case for a single error argument
			message = fmt.Sprintf("%s: %v", message, err)
		} else {
			// Format message with arguments
			message = fmt.Sprintf(message, args...)
		}
	}
	
	// Format log entry
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	logEntry := fmt.Sprintf("[%s] %s: %s", timestamp, level, message)
	
	// Log the entry
	l.logger.Println(logEntry)
}