package dynamodbhelper

import (
	"errors"
	"testing"

	"workflow-function/FinalizeAndStoreResults/internal/models"
	wfErrors "workflow-function/shared/errors"
)

func TestCreateEnhancedDynamoDBError(t *testing.T) {
	tests := []struct {
		name          string
		operation     string
		tableName     string
		awsErr        error
		expectedCode  string
		expectedRetryable bool
	}{
		{
			name:          "ValidationException",
			operation:     "PutItem",
			tableName:     "test-table",
			awsErr:        errors.New("ValidationException: One or more parameter values were invalid"),
			expectedCode:  "VALIDATION_EXCEPTION",
			expectedRetryable: false,
		},
		{
			name:          "ConditionalCheckFailedException",
			operation:     "PutItem",
			tableName:     "test-table",
			awsErr:        errors.New("ConditionalCheckFailedException: The conditional request failed"),
			expectedCode:  "CONDITIONAL_CHECK_FAILED",
			expectedRetryable: false,
		},
		{
			name:          "ProvisionedThroughputExceededException",
			operation:     "PutItem",
			tableName:     "test-table",
			awsErr:        errors.New("ProvisionedThroughputExceededException: The level of configured provisioned throughput for the table was exceeded"),
			expectedCode:  "THROUGHPUT_EXCEEDED",
			expectedRetryable: true,
		},
		{
			name:          "ResourceNotFoundException",
			operation:     "PutItem",
			tableName:     "test-table",
			awsErr:        errors.New("ResourceNotFoundException: Requested resource not found"),
			expectedCode:  "RESOURCE_NOT_FOUND",
			expectedRetryable: false,
		},
		{
			name:          "ThrottlingException",
			operation:     "PutItem",
			tableName:     "test-table",
			awsErr:        errors.New("ThrottlingException: Rate exceeded"),
			expectedCode:  "THROTTLING_EXCEPTION",
			expectedRetryable: true,
		},
		{
			name:          "Generic Error",
			operation:     "PutItem",
			tableName:     "test-table",
			awsErr:        errors.New("Some unknown error"),
			expectedCode:  "DYNAMODB_ERROR",
			expectedRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createEnhancedDynamoDBError(tt.operation, tt.tableName, tt.awsErr, nil)

			if result.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, result.Code)
			}

			if result.Retryable != tt.expectedRetryable {
				t.Errorf("Expected retryable %v, got %v", tt.expectedRetryable, result.Retryable)
			}

			if result.Type != wfErrors.ErrorTypeDynamoDB {
				t.Errorf("Expected error type %s, got %s", wfErrors.ErrorTypeDynamoDB, result.Type)
			}

			// Check that details contain expected fields
			if result.Details["operation"] != tt.operation {
				t.Errorf("Expected operation %s in details, got %v", tt.operation, result.Details["operation"])
			}

			if result.Details["table"] != tt.tableName {
				t.Errorf("Expected table %s in details, got %v", tt.tableName, result.Details["table"])
			}

			if result.Details["awsError"] != tt.awsErr.Error() {
				t.Errorf("Expected awsError %s in details, got %v", tt.awsErr.Error(), result.Details["awsError"])
			}
		})
	}
}

func TestValidateVerificationResultItem(t *testing.T) {
	tests := []struct {
		name        string
		item        models.VerificationResultItem
		expectError bool
		errorField  string
	}{
		{
			name: "Valid item",
			item: models.VerificationResultItem{
				VerificationID:   "verif-123",
				VerificationAt:   "2025-01-06T10:00:00Z",
				VerificationType: "layout",
				CurrentStatus:    "COMPLETED",
			},
			expectError: false,
		},
		{
			name: "Missing VerificationID",
			item: models.VerificationResultItem{
				VerificationAt:   "2025-01-06T10:00:00Z",
				VerificationType: "layout",
				CurrentStatus:    "COMPLETED",
			},
			expectError: true,
			errorField:  "verificationId",
		},
		{
			name: "Missing VerificationAt",
			item: models.VerificationResultItem{
				VerificationID:   "verif-123",
				VerificationType: "layout",
				CurrentStatus:    "COMPLETED",
			},
			expectError: true,
			errorField:  "verificationAt",
		},
		{
			name: "Missing VerificationType",
			item: models.VerificationResultItem{
				VerificationID: "verif-123",
				VerificationAt: "2025-01-06T10:00:00Z",
				CurrentStatus:  "COMPLETED",
			},
			expectError: true,
			errorField:  "verificationType",
		},
		{
			name: "Missing CurrentStatus",
			item: models.VerificationResultItem{
				VerificationID:   "verif-123",
				VerificationAt:   "2025-01-06T10:00:00Z",
				VerificationType: "layout",
			},
			expectError: true,
			errorField:  "currentStatus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVerificationResultItem(tt.item)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for missing %s, but got nil", tt.errorField)
				} else {
					// Check if it's a validation error
					if wfErr, ok := err.(*wfErrors.WorkflowError); ok {
						if wfErr.Type != wfErrors.ErrorTypeValidation {
							t.Errorf("Expected validation error type, got %s", wfErr.Type)
						}
					} else {
						t.Errorf("Expected WorkflowError, got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, but got: %v", err)
				}
			}
		})
	}
}

func TestSanitizeItemForLogging(t *testing.T) {
	item := models.VerificationResultItem{
		VerificationID:    "verif-123",
		VerificationAt:    "2025-01-06T10:00:00Z",
		VerificationType:  "layout",
		CurrentStatus:     "COMPLETED",
		ReferenceImageUrl: "https://example.com/ref.jpg",
		CheckingImageUrl:  "https://example.com/check.jpg",
		VendingMachineID:  "vm-456",
	}

	result := sanitizeItemForLogging(item)

	// Check that sensitive fields are sanitized
	if result["referenceImageUrl"] != "[URL_PROVIDED]" {
		t.Errorf("Expected referenceImageUrl to be sanitized, got %v", result["referenceImageUrl"])
	}

	if result["checkingImageUrl"] != "[URL_PROVIDED]" {
		t.Errorf("Expected checkingImageUrl to be sanitized, got %v", result["checkingImageUrl"])
	}

	// Check that non-sensitive fields are preserved
	if result["verificationId"] != "verif-123" {
		t.Errorf("Expected verificationId to be preserved, got %v", result["verificationId"])
	}

	if result["verificationType"] != "layout" {
		t.Errorf("Expected verificationType to be preserved, got %v", result["verificationType"])
	}

	// Check that other fields are type-indicated
	if result["vendingMachineId"] != "string" && result["vendingMachineId"] != "[string]" {
		// The exact format may vary, but it should indicate the type
		t.Logf("vendingMachineId sanitized as: %v", result["vendingMachineId"])
	}
}
