package dynamodbhelper

import (
	"testing"

	"workflow-function/FinalizeAndStoreResults/internal/models"
	"workflow-function/shared/schema"
)

func TestValidateVerificationResultItem(t *testing.T) {
	tests := []struct {
		name        string
		item        models.VerificationResultItem
		expectError bool
		errorField  string
	}{
		{
			name: "valid item",
			item: models.VerificationResultItem{
				VerificationID:     "verif-20250605045209-9e81",
				VerificationAt:     "2025-06-05T04:53:12Z",
				VerificationType:   schema.VerificationTypeLayoutVsChecking,
				CurrentStatus:      schema.StatusCompleted,
				VerificationStatus: schema.VerificationStatusCorrect,
			},
			expectError: false,
		},
		{
			name: "missing verificationId",
			item: models.VerificationResultItem{
				VerificationAt:     "2025-06-05T04:53:12Z",
				VerificationType:   schema.VerificationTypeLayoutVsChecking,
				CurrentStatus:      schema.StatusCompleted,
				VerificationStatus: schema.VerificationStatusCorrect,
			},
			expectError: true,
			errorField:  "verificationId",
		},
		{
			name: "missing verificationAt",
			item: models.VerificationResultItem{
				VerificationID:     "verif-20250605045209-9e81",
				VerificationType:   schema.VerificationTypeLayoutVsChecking,
				CurrentStatus:      schema.StatusCompleted,
				VerificationStatus: schema.VerificationStatusCorrect,
			},
			expectError: true,
			errorField:  "verificationAt",
		},
		{
			name: "missing verificationType",
			item: models.VerificationResultItem{
				VerificationID:     "verif-20250605045209-9e81",
				VerificationAt:     "2025-06-05T04:53:12Z",
				CurrentStatus:      schema.StatusCompleted,
				VerificationStatus: schema.VerificationStatusCorrect,
			},
			expectError: true,
			errorField:  "verificationType",
		},
		{
			name: "missing currentStatus",
			item: models.VerificationResultItem{
				VerificationID:     "verif-20250605045209-9e81",
				VerificationAt:     "2025-06-05T04:53:12Z",
				VerificationType:   schema.VerificationTypeLayoutVsChecking,
				VerificationStatus: schema.VerificationStatusCorrect,
			},
			expectError: true,
			errorField:  "currentStatus",
		},
		{
			name: "empty verificationStatus",
			item: models.VerificationResultItem{
				VerificationID:     "verif-20250605045209-9e81",
				VerificationAt:     "2025-06-05T04:53:12Z",
				VerificationType:   schema.VerificationTypeLayoutVsChecking,
				CurrentStatus:      schema.StatusCompleted,
				VerificationStatus: "", // This should cause validation error
			},
			expectError: true,
			errorField:  "verificationStatus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVerificationResultItem(tt.item)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for field %s, but got none", tt.errorField)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error, but got: %v", err)
				}
			}
		})
	}
}
