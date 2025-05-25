// Package main provides the entry point for the initialization Lambda function
// This file contains only type aliases to maintain backward compatibility
package main

import (
	"workflow-function/shared/logger"
)

// Logger interface type alias for backward compatibility
type Logger = logger.Logger

// LogLevel type alias for backward compatibility
type LogLevel = logger.LogLevel

// Constants
const (
	LogLevelDebug = logger.LogLevelDebug
	LogLevelInfo  = logger.LogLevelInfo
	LogLevelWarn  = logger.LogLevelWarn 
	LogLevelError = logger.LogLevelError
)

// NewStructuredLogger is a wrapper around the shared logger for backward compatibility
func NewStructuredLogger() Logger {
	return logger.New("kootoro-verification", "InitializeFunction")
}