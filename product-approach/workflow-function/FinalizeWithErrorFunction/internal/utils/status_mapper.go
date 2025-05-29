package utils

import "workflow-function/shared/schema"

// DetermineSpecificErrorStatus maps an error stage to a schema status constant
func DetermineSpecificErrorStatus(stage string) string {
	switch stage {
	case "INITIALIZATION":
		return schema.StatusInitializationFailed
	case "IMAGE_FETCH":
		return schema.StatusImageFetchFailed
	case "BEDROCK_PROCESSING":
		return schema.StatusBedrockProcessingFailed
	case "HISTORICAL_FETCH":
		return schema.StatusHistoricalFetchFailed
	default:
		return schema.StatusVerificationFailed
	}
}
