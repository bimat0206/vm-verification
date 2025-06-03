package models

import "workflow-function/shared/schema"

// StepFunctionsErrorCause represents the structured error cause from Step Functions.
type StepFunctionsErrorCause struct {
	ErrorMessage string   `json:"errorMessage,omitempty"`
	ErrorType    string   `json:"errorType,omitempty"`
	StackTrace   []string `json:"stackTrace,omitempty"`
}

// LambdaInput is the input event for the FinalizeWithError Lambda function.
type LambdaInput struct {
	VerificationID      string                   `json:"verificationId"`
	Error               map[string]interface{}   `json:"error"`
	ErrorStage          string                   `json:"errorStage"`
	PartialS3References schema.InputS3References `json:"partialS3References"`
}

// LambdaOutput is the result returned from the FinalizeWithError Lambda.
type LambdaOutput struct {
	VerificationID string `json:"verificationId"`
	Status         string `json:"status"`
	ErrorStage     string `json:"errorStage"`
	ErrorMessage   string `json:"errorMessage"`
	Message        string `json:"message"`
}
