// examples/error_handling.go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/kootoro/s3state"
)

func main() {
	// Example 1: Create and check different error types
	s3Err := s3state.NewS3Error("GetObject", "failed to get object", errors.New("network timeout"))
	validationErr := s3state.NewValidationError("Store", "bucket name is required")
	jsonErr := s3state.NewJSONError("Marshal", "invalid data structure", errors.New("json: unsupported type"))

	// Check error types
	if s3state.IsS3Error(s3Err) {
		fmt.Println("This is an S3 error")
	}

	if s3state.IsValidationError(validationErr) {
		fmt.Println("This is a validation error")
	}

	if s3state.IsJSONError(jsonErr) {
		fmt.Println("This is a JSON error")
	}

	// Example 2: Check if errors are retryable
	timeoutErr := errors.New("connection timeout")
	if s3state.IsRetryable(timeoutErr) {
		fmt.Println("Error is retryable")
	}

	// Example 3: Validation helpers
	ref := &s3state.Reference{
		Bucket: "",  // Invalid - empty bucket
		Key:    "some-key",
	}

	if err := s3state.ValidateReference(ref, "Store"); err != nil {
		fmt.Printf("Reference validation failed: %v\n", err)
	}

	// Example 4: Validate category
	if err := s3state.ValidateCategory("invalid-category", "Store"); err != nil {
		fmt.Printf("Category validation failed: %v\n", err)
	}

	// Example 5: Validate envelope
	envelope := &s3state.Envelope{
		VerificationID: "", // Invalid - empty ID
		References:     make(map[string]*s3state.Reference),
	}

	if err := s3state.ValidateEnvelope(envelope, "Process"); err != nil {
		fmt.Printf("Envelope validation failed: %v\n", err)
	}

	// Example 6: Error list for collecting multiple errors
	errorList := s3state.NewErrorList()
	errorList.Add(s3state.NewValidationError("Step1", "first error"))
	errorList.Add(s3state.NewValidationError("Step2", "second error"))
	errorList.Add(nil) // This will be ignored

	if errorList.HasErrors() {
		fmt.Printf("Collected errors: %v\n", errorList.ToError())
		fmt.Printf("First error: %v\n", errorList.First())
	}

	// Example 7: Error wrapping
	originalErr := errors.New("original problem")
	wrappedErr := s3state.WrapError("ProcessData", originalErr)
	fmt.Printf("Wrapped error: %v\n", wrappedErr)
}