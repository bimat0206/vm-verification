package utils

import (
	"encoding/json"

	"workflow-function/FinalizeWithErrorFunction/internal/models"
)

// ParseStepFunctionsErrorCause parses the JSON string from Step Functions error cause
func ParseStepFunctionsErrorCause(causeJSONString string) (*models.StepFunctionsErrorCause, error) {
	var cause models.StepFunctionsErrorCause
	if causeJSONString == "" {
		return &cause, nil
	}
	if err := json.Unmarshal([]byte(causeJSONString), &cause); err != nil {
		return nil, err
	}
	return &cause, nil
}
