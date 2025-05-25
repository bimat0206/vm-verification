// Package main provides the entry point for the initialization Lambda function
// This file contains model aliases for backward compatibility
package main

import (
	"time"
	"workflow-function/shared/schema"
)

// Verification types - using constants from schema package
const (
	VerificationTypeLayoutVsChecking  = schema.VerificationTypeLayoutVsChecking
	VerificationTypePreviousVsCurrent = schema.VerificationTypePreviousVsCurrent
)

// Verification status constants - using constants from schema package
const (
	StatusVerificationRequested  = schema.StatusVerificationRequested
	StatusVerificationInitialized = schema.StatusVerificationInitialized
	StatusFetchingImages         = schema.StatusFetchingImages
	StatusImagesFetched          = schema.StatusImagesFetched
	StatusPromptPrepared         = schema.StatusPromptPrepared
	StatusTurn1PromptReady       = schema.StatusTurn1PromptReady
	StatusTurn1Completed         = schema.StatusTurn1Completed
	StatusTurn1Processed         = schema.StatusTurn1Processed
	StatusTurn2PromptReady       = schema.StatusTurn2PromptReady
	StatusTurn2Completed         = schema.StatusTurn2Completed
	StatusTurn2Processed         = schema.StatusTurn2Processed
	StatusResultsFinalized       = schema.StatusResultsFinalized
	StatusResultsStored          = schema.StatusResultsStored
	StatusNotificationSent       = schema.StatusNotificationSent
	StatusCompleted              = schema.StatusCompleted
	
	// Error states
	StatusInitializationFailed    = schema.StatusInitializationFailed
	StatusHistoricalFetchFailed   = schema.StatusHistoricalFetchFailed
	StatusImageFetchFailed        = schema.StatusImageFetchFailed
	StatusBedrockProcessingFailed = schema.StatusBedrockProcessingFailed
	StatusVerificationFailed      = schema.StatusVerificationFailed
)

// Using SchemaVersion from schema package
const SchemaVersion = schema.SchemaVersion

// InitResponse matches the standardized output schema
type InitResponse struct {
	SchemaVersion       string                    `json:"schemaVersion"`
	VerificationContext *schema.VerificationContext `json:"verificationContext"`
	Message             string                    `json:"message,omitempty"`
}

// Legacy model type aliases for backward compatibility
type TurnConfig = schema.TurnConfig
type TurnTimestamps = schema.TurnTimestamps
type RequestMetadata = schema.RequestMetadata
type ResourceValidation = schema.ResourceValidation

// FormatISO8601 returns now in RFC3339 - using schema implementation
func FormatISO8601() string {
	return schema.FormatISO8601()
}

// GetCurrentTimestamp returns time formatted for IDs - legacy API compatibility
func GetCurrentTimestamp() string {
	return time.Now().UTC().Format("20060102150405")
}