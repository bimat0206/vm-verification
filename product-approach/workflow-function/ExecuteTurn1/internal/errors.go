package internal

import (
	"fmt"
	"time"
	
	"workflow-function/shared/errors"
)

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	return errors.IsRetryable(err)
}

// ToLambdaErrorResponse converts an error to a Lambda error response
type LambdaErrorResponse struct {
	Error struct {
		ErrorType      string                 `json:"errorType"`
		ErrorMessage   string                 `json:"errorMessage"`
		ErrorCode      string                 `json:"errorCode"`
		Details        map[string]interface{} `json:"details"`
		Severity       string                 `json:"severity"`
		APISource      string                 `json:"apiSource"`
	} `json:"error"`
	VerificationContext struct {
		VerificationID string `json:"verificationId"`
		Status         string `json:"status"`
	} `json:"verificationContext"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"requestId"`
}

// ToLambdaErrorResponse converts an error to a Lambda error response
func ToLambdaErrorResponse(err error, verificationID string) *LambdaErrorResponse {
	errorType := string(errors.GetErrorType(err))
	errorSeverity := string(errors.GetErrorSeverity(err))
	apiSource := string(errors.GetAPISource(err))
	
	if workflowErr, ok := err.(*errors.WorkflowError); ok {
		return &LambdaErrorResponse{
			Error: struct {
				ErrorType      string                 `json:"errorType"`
				ErrorMessage   string                 `json:"errorMessage"`
				ErrorCode      string                 `json:"errorCode"`
				Details        map[string]interface{} `json:"details"`
				Severity       string                 `json:"severity"`
				APISource      string                 `json:"apiSource"`
			}{
				ErrorType:    errorType,
				ErrorMessage: workflowErr.Message,
				ErrorCode:    workflowErr.Code,
				Details:      workflowErr.Details,
				Severity:     errorSeverity,
				APISource:    apiSource,
			},
			VerificationContext: struct {
				VerificationID string `json:"verificationId"`
				Status         string `json:"status"`
			}{
				VerificationID: verificationID,
				Status:         "TURN1_FAILED",
			},
			Timestamp: time.Now(),
			RequestID: workflowErr.RequestID,
		}
	}

	// Fallback for non-WorkflowError types
	return &LambdaErrorResponse{
		Error: struct {
			ErrorType      string                 `json:"errorType"`
			ErrorMessage   string                 `json:"errorMessage"`
			ErrorCode      string                 `json:"errorCode"`
			Details        map[string]interface{} `json:"details"`
			Severity       string                 `json:"severity"`
			APISource      string                 `json:"apiSource"`
		}{
			ErrorType:    errorType,
			ErrorMessage: err.Error(),
			ErrorCode:    "UNKNOWN_ERROR",
			Details:      map[string]interface{}{},
			Severity:     errorSeverity,
			APISource:    apiSource,
		},
		VerificationContext: struct {
			VerificationID string `json:"verificationId"`
			Status         string `json:"status"`
		}{
			VerificationID: verificationID,
			Status:         "TURN1_FAILED",
		},
		Timestamp: time.Now(),
	}
}

// ErrorMetrics tracks error patterns and frequencies
type ErrorMetrics struct {
	TotalErrors          int64                    `json:"totalErrors"`
	ErrorsByType         map[string]int64         `json:"errorsByType"`
	ErrorsByAPI          map[string]int64         `json:"errorsByAPI"`
	ErrorsBySeverity     map[string]int64         `json:"errorsBySeverity"`
	RetryableErrors      int64                    `json:"retryableErrors"`
	NonRetryableErrors   int64                    `json:"nonRetryableErrors"`
	LastErrorTime        time.Time                `json:"lastErrorTime"`
	ErrorRate            float64                  `json:"errorRate"`
}

var globalErrorMetrics = &ErrorMetrics{
	ErrorsByType:     make(map[string]int64),
	ErrorsByAPI:      make(map[string]int64),
	ErrorsBySeverity: make(map[string]int64),
}

// RecordError records error metrics
func RecordError(err error) {
	globalErrorMetrics.TotalErrors++
	
	errorType := string(errors.GetErrorType(err))
	apiSource := string(errors.GetAPISource(err))
	severity := string(errors.GetErrorSeverity(err))
	
	globalErrorMetrics.ErrorsByType[errorType]++
	globalErrorMetrics.ErrorsByAPI[apiSource]++
	globalErrorMetrics.ErrorsBySeverity[severity]++
	
	if errors.IsRetryable(err) {
		globalErrorMetrics.RetryableErrors++
	} else {
		globalErrorMetrics.NonRetryableErrors++
	}
	
	globalErrorMetrics.LastErrorTime = time.Now()
}

// GetErrorMetrics returns current error metrics
func GetErrorMetrics() *ErrorMetrics {
	return globalErrorMetrics
}

// ResetErrorMetrics resets error metrics
func ResetErrorMetrics() {
	globalErrorMetrics = &ErrorMetrics{
		ErrorsByType:     make(map[string]int64),
		ErrorsByAPI:      make(map[string]int64),
		ErrorsBySeverity: make(map[string]int64),
	}
}

// CalculateErrorRate calculates error rate based on total requests
func CalculateErrorRate(totalRequests int64) float64 {
	if totalRequests == 0 {
		return 0.0
	}
	return float64(globalErrorMetrics.TotalErrors) / float64(totalRequests) * 100
}

// GetTopErrors returns the most frequent error types
func GetTopErrors(limit int) map[string]int64 {
	topErrors := make(map[string]int64)
	
	// Simple implementation - in production, you might want to sort
	count := 0
	for errorType, frequency := range globalErrorMetrics.ErrorsByType {
		if count >= limit {
			break
		}
		topErrors[errorType] = frequency
		count++
	}
	
	return topErrors
}

// LogErrorSummary logs a summary of all errors
func LogErrorSummary() {
	fmt.Printf("Error Summary:\n")
	fmt.Printf("  Total Errors: %d\n", globalErrorMetrics.TotalErrors)
	fmt.Printf("  Retryable: %d\n", globalErrorMetrics.RetryableErrors)
	fmt.Printf("  Non-retryable: %d\n", globalErrorMetrics.NonRetryableErrors)
	
	fmt.Printf("  By Type:\n")
	for errorType, count := range globalErrorMetrics.ErrorsByType {
		fmt.Printf("    %s: %d\n", errorType, count)
	}
	
	fmt.Printf("  By API:\n")
	for apiSource, count := range globalErrorMetrics.ErrorsByAPI {
		fmt.Printf("    %s: %d\n", apiSource, count)
	}
	
	fmt.Printf("  By Severity:\n")
	for severity, count := range globalErrorMetrics.ErrorsBySeverity {
		fmt.Printf("    %s: %d\n", severity, count)
	}
}
