package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
)

// LogLevel type for log levels.
type LogLevel string

const (
    InfoLevel  LogLevel = "INFO"
    WarnLevel  LogLevel = "WARN"
    ErrorLevel LogLevel = "ERROR"
)

// LogEntry represents a structured log entry.
type LogEntry struct {
    Timestamp string      `json:"timestamp"`
    Level     LogLevel    `json:"level"`
    Message   string      `json:"message"`
    Details   interface{} `json:"details,omitempty"`
}

// log prints a structured log entry to stdout.
func log(level LogLevel, msg string, details interface{}) {
    entry := LogEntry{
        Timestamp: time.Now().Format(time.RFC3339),
        Level:     level,
        Message:   msg,
        Details:   details,
    }
    data, _ := json.Marshal(entry)
    fmt.Fprintln(os.Stdout, string(data))
}

// Info logs an informational message.
func Info(msg string, details interface{}) {
    log(InfoLevel, msg, details)
}

// Warn logs a warning message.
func Warn(msg string, details interface{}) {
    log(WarnLevel, msg, details)
}

// Error logs an error message.
func Error(msg string, details interface{}) {
    log(ErrorLevel, msg, details)
}
