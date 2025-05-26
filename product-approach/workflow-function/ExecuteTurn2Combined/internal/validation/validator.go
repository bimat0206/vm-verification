package validation

import (
	"fmt"

	"workflow-function/ExecuteTurn2Combined/internal/models"
)

// ValidateRequest performs minimal validation of the Turn2Request.
func ValidateRequest(req *models.Turn2Request) error {
	if req.VerificationID == "" {
		return fmt.Errorf("verificationId is required")
	}
	if req.S3Refs.Prompts.System.Key == "" || req.S3Refs.Prompts.System.Bucket == "" {
		return fmt.Errorf("system prompt reference is required")
	}
	if req.S3Refs.Images.CheckingBase64.Key == "" {
		return fmt.Errorf("checking image reference is required")
	}
	if req.S3Refs.Processing.Turn1Markdown.Key == "" {
		return fmt.Errorf("turn1 markdown reference is required")
	}
	return nil
}
