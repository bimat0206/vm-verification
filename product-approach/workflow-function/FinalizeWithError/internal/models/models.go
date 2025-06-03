package models

import "workflow-function/shared/schema"

// StepFunctionsErrorCause represents the structured error cause from Step Functions.
type StepFunctionsErrorCause struct {
	ErrorMessage string   `json:"errorMessage,omitempty"`
	ErrorType    string   `json:"errorType,omitempty"`
	StackTrace   []string `json:"stackTrace,omitempty"`
}

// InitializationData represents the initialization data structure
type InitializationData struct {
	SchemaVersion       string                     `json:"schemaVersion"`
	VerificationContext schema.VerificationContext `json:"verificationContext"`
	SystemPrompt        interface{}                `json:"systemPrompt"`
	LayoutMetadata      *schema.LayoutMetadata     `json:"layoutMetadata,omitempty"`
	// Error tracking fields
	Status              string                     `json:"status,omitempty"`
	ErrorStage          string                     `json:"errorStage,omitempty"`
	ErrorMessage        string                     `json:"errorMessage,omitempty"`
	FailedAt            string                     `json:"failedAt,omitempty"`
}

// S3Reference represents a pointer to an object in S3
type S3Reference struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	ETag   string `json:"etag,omitempty"`
	Size   int64  `json:"size,omitempty"`
}

// S3ErrorReference represents error-specific S3 storage details
type S3ErrorReference struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Size   int64  `json:"size"`
}

// InputS3References represents S3 references for the function input
type InputS3References struct {
	InitializationS3Ref S3Reference `json:"processing_initialization"`
	Turn1ProcessedS3Ref S3Reference `json:"turn1Processed"`
	Turn2ProcessedS3Ref S3Reference `json:"turn2Processed"`
}

// LambdaInput is the input event for the FinalizeWithError Lambda function.
type LambdaInput struct {
	VerificationID string             `json:"verificationId"`
	Error          map[string]interface{} `json:"error"`
	ErrorStage     string             `json:"errorStage"`
	S3References   InputS3References  `json:"s3References"`
}

// LambdaOutput is the result returned from the FinalizeWithError Lambda.
type LambdaOutput struct {
	VerificationID string             `json:"verificationId"`
	S3References   struct {
		ProcessingInitialization S3Reference      `json:"processing_initialization"`
		Error                    S3ErrorReference `json:"error"`
	} `json:"s3References"`
	Status       string `json:"status"`
	ErrorStage   string `json:"errorStage"`
	ErrorMessage string `json:"errorMessage"`
	Message      string `json:"message"`
}
