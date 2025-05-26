package models

// Turn2Request represents the input payload for ExecuteTurn2Combined
// It references artifacts produced by Turn1

type Turn2Request struct {
    VerificationID      string              `json:"verificationId"`
    VerificationContext VerificationContext `json:"verificationContext"`
    S3Refs              Turn2RequestS3Refs  `json:"s3Refs"`
    InputInitializationFileRef S3Reference  `json:"-"`
}

type Turn2RequestS3Refs struct {
    Prompts    PromptRefs           `json:"prompts"`
    Images     Turn2ImageRefs       `json:"images"`
    Processing Turn2ProcessingRefs  `json:"processing"`
}

type Turn2ImageRefs struct {
    CheckingBase64 S3Reference `json:"checkingBase64"`
}

type Turn2ProcessingRefs struct {
    Turn1Markdown S3Reference `json:"turn1Markdown"`
}

// StepFunctionResponse is a simplified output for Step Functions
// Provided here for main wrapper

type StepFunctionResponse struct {
    VerificationID string           `json:"verificationId"`
    Status         string           `json:"status"`
    S3References   map[string]S3Reference `json:"s3References"`
}
