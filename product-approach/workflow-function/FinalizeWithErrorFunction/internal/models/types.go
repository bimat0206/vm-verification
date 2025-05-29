package models

import "workflow-function/shared/schema"

// StepFunctionsErrorCause represents the structure of the "Cause" string from Step Functions
type StepFunctionsErrorCause struct {
	ErrorMessage *string `json:"errorMessage,omitempty"`
	ErrorType    *string `json:"errorType,omitempty"`
	StackTrace   *string `json:"stackTrace,omitempty"`
}

// StepFunctionsError is the error object received from Step Functions
type StepFunctionsError struct {
	Error       *string                  `json:"Error,omitempty"`
	Cause       *string                  `json:"Cause,omitempty"`
	ParsedCause *StepFunctionsErrorCause `json:"-"`
}

// FinalizeWithErrorInput is the expected input structure for the Lambda
type FinalizeWithErrorInput struct {
	SchemaVersion       string                        `json:"schemaVersion"`
	VerificationID      string                        `json:"verificationId"`
	Error               StepFunctionsError            `json:"error"`
	ErrorStage          string                        `json:"errorStage"`
	PartialS3References map[string]schema.S3Reference `json:"partialS3References"`
}

// FinalizeWithErrorOutput is the expected output structure of the Lambda
type FinalizeWithErrorOutput struct {
	SchemaVersion       string                        `json:"schemaVersion"`
	VerificationID      string                        `json:"verificationId"`
	S3References        map[string]schema.S3Reference `json:"s3References"`
	Status              string                        `json:"status"`
	Error               schema.ErrorInfo              `json:"error"`
	VerificationContext *schema.VerificationContext   `json:"verificationContext"`
	Summary             map[string]interface{}        `json:"summary"`
}
