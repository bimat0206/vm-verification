// internal/utils/errors.go
package errors

import "fmt"

// ExecutionStage represents the high-level stage where an error occurred.
type ExecutionStage string

const (
	StageValidation       ExecutionStage = "validation"
	StageContextLoading   ExecutionStage = "context_loading"
	StagePromptGeneration ExecutionStage = "prompt_generation"
	StageBedrockCall      ExecutionStage = "bedrock_invocation"
	StageProcessing       ExecutionStage = "response_processing"
	StageStorage          ExecutionStage = "state_storage"
	StageDynamoDB         ExecutionStage = "dynamodb_update"
)

// RetryableError marks an error that should trigger Step-Functions retry logic.
type RetryableError struct {
	Stage ExecutionStage
	Err   error
	Msg   string
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("%s [retryable] %s: %v", e.Stage, e.Msg, e.Err)
}

// NonRetryableError marks an error that should be sent to the catch path immediately.
type NonRetryableError struct {
	Stage ExecutionStage
	Err   error
	Msg   string
}

func (e *NonRetryableError) Error() string {
	return fmt.Sprintf("%s [non-retryable] %s: %v", e.Stage, e.Msg, e.Err)
}

// WrapRetryable annotates err as retryable at the given stage.
func WrapRetryable(err error, stage ExecutionStage, msg string) error {
	return &RetryableError{Stage: stage, Err: err, Msg: msg}
}

// WrapNonRetryable annotates err as non-retryable at the given stage.
func WrapNonRetryable(err error, stage ExecutionStage, msg string) error {
	return &NonRetryableError{Stage: stage, Err: err, Msg: msg}
}

// IsRetryable returns true if err was wrapped as a RetryableError.
func IsRetryable(err error) bool {
	_, ok := err.(*RetryableError)
	return ok
}

// ToStepFnError prepares an error for Step Functions consumption.
// We return the error itself so its concrete type becomes the Lambda errorType.
func ToStepFnError(err error) error {
	return err
}
