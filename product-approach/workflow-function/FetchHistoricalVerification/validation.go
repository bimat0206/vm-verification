package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMissingVerificationID    = errors.New("missing verificationId")
	ErrMissingVerificationType  = errors.New("missing verificationType")
	ErrInvalidVerificationType  = errors.New("invalid verificationType, expected 'previous_vs_current'")
	ErrMissingReferenceImageURL = errors.New("missing referenceImageUrl")
	ErrMissingCheckingImageURL  = errors.New("missing checkingImageUrl")
	ErrMissingVendingMachineID  = errors.New("missing vendingMachineId")
)

// validateInput validates the input parameters
func validateInput(ctx VerificationContext) error {
	// Check required fields
	if ctx.VerificationID == "" {
		return ErrMissingVerificationID
	}

	if ctx.VerificationType == "" {
		return ErrMissingVerificationType
	}

	// Ensure verificationType is 'previous_vs_current'
	if !strings.EqualFold(ctx.VerificationType, "previous_vs_current") {
		return ErrInvalidVerificationType
	}

	if ctx.ReferenceImageURL == "" {
		return ErrMissingReferenceImageURL
	}

	if ctx.CheckingImageURL == "" {
		return ErrMissingCheckingImageURL
	}

	if ctx.VendingMachineID == "" {
		return ErrMissingVendingMachineID
	}

	// Verify S3 URL format for reference image
	if !strings.HasPrefix(ctx.ReferenceImageURL, "s3://") {
		return fmt.Errorf("invalid reference image URL format, expected s3:// prefix: %s", ctx.ReferenceImageURL)
	}

	// For previous_vs_current, reference image should be in the checking bucket
	if !isCheckingBucketURL(ctx.ReferenceImageURL) {
		return fmt.Errorf("for previous_vs_current verification, referenceImageUrl must point to the checking bucket: %s", ctx.ReferenceImageURL)
	}

	return nil
}

// isCheckingBucketURL checks if the URL is from the checking bucket
func isCheckingBucketURL(url string) bool {
	checkingBucket := getCheckingBucketName()
	return strings.Contains(url, checkingBucket)
}