package main

import (
	"time"
)

// LayoutMetadata represents the layout metadata stored in DynamoDB
type LayoutMetadata struct {
	LayoutId          int                    `json:"layoutId"`
	LayoutPrefix      string                 `json:"layoutPrefix"`
	VendingMachineId  string                 `json:"vendingMachineId"`
	Location          string                 `json:"location"`
	CreatedAt         string                 `json:"createdAt"`
	UpdatedAt         string                 `json:"updatedAt"`
	ReferenceImageUrl string                 `json:"referenceImageUrl"`
	SourceJsonUrl     string                 `json:"sourceJsonUrl"`
	MachineStructure  map[string]interface{} `json:"machineStructure"`
	ProductPositionMap map[string]interface{} `json:"productPositionMap"`
}

// TurnConfig represents the turn configuration for a verification
type TurnConfig struct {
	MaxTurns           int `json:"maxTurns"`
	ReferenceImageTurn int `json:"referenceImageTurn"`
	CheckingImageTurn  int `json:"checkingImageTurn"`
}

// RequestMetadata represents metadata about the verification request
type RequestMetadata struct {
	RequestId         string `json:"requestId"`
	RequestTimestamp  string `json:"requestTimestamp"`
	ProcessingStarted string `json:"processingStarted"`
}

// ResourceValidation represents validation information for resources
type ResourceValidation struct {
	LayoutExists         bool   `json:"layoutExists"`
	ReferenceImageExists bool   `json:"referenceImageExists"`
	CheckingImageExists  bool   `json:"checkingImageExists"`
	ValidationTimestamp  string `json:"validationTimestamp"`
}

// VerificationRecord represents a verification record stored in DynamoDB
type VerificationRecord struct {
	VerificationId      string                 `json:"verificationId"`
	VerificationType    string                 `json:"verificationType"`
	VendingMachineId    string                 `json:"vendingMachineId"`
	VerificationAt      string                 `json:"verificationAt"`
	Status              string                 `json:"status"`
	TurnConfig          *TurnConfig            `json:"turnConfig"`
	RequestMetadata     *RequestMetadata       `json:"requestMetadata"`
	ResourceValidation  *ResourceValidation    `json:"resourceValidation"`
	AdditionalMetadata  map[string]interface{} `json:"additionalMetadata"`
}

// CalculateHoursSince calculates hours since the verification time
func (v *VerificationRecord) CalculateHoursSince() float64 {
	if v.VerificationAt == "" {
		return 0
	}
	
	verTime, err := time.Parse(time.RFC3339, v.VerificationAt)
	if err != nil {
		return 0
	}
	
	return time.Since(verTime).Hours()
}

// ToHistoricalContext converts a verification record to a historical context map
func (v *VerificationRecord) ToHistoricalContext() map[string]interface{} {
	// Calculate hours since verification
	hoursSince := v.CalculateHoursSince()
	
	// Extract machine structure from TurnConfig if available
	var machineStructure map[string]interface{}
	if v.TurnConfig != nil {
		machineStructure = map[string]interface{}{
			"maxTurns":           v.TurnConfig.MaxTurns,
			"referenceImageTurn": v.TurnConfig.ReferenceImageTurn,
			"checkingImageTurn":  v.TurnConfig.CheckingImageTurn,
		}
	} else {
		machineStructure = make(map[string]interface{})
	}
	
	// Create verification summary
	verificationSummary := map[string]interface{}{
		"verificationId":   v.VerificationId,
		"verificationType": v.VerificationType,
		"vendingMachineId": v.VendingMachineId,
	}
	
	// Build the historical context response
	result := map[string]interface{}{
		"previousVerificationId":     v.VerificationId,
		"previousVerificationAt":     v.VerificationAt,
		"previousVerificationStatus": v.Status,
		"hoursSinceLastVerification": hoursSince,
		"machineStructure":           machineStructure,
		"verificationSummary":        verificationSummary,
		"vendingMachineId":           v.VendingMachineId,
		"verificationType":           v.VerificationType,
	}
	
	// Add request metadata if available
	if v.RequestMetadata != nil {
		result["requestMetadata"] = map[string]interface{}{
			"requestId":         v.RequestMetadata.RequestId,
			"requestTimestamp":  v.RequestMetadata.RequestTimestamp,
			"processingStarted": v.RequestMetadata.ProcessingStarted,
		}
	}
	
	// Add resource validation info if available
	if v.ResourceValidation != nil {
		result["resourceValidation"] = map[string]interface{}{
			"layoutExists":         v.ResourceValidation.LayoutExists,
			"referenceImageExists": v.ResourceValidation.ReferenceImageExists,
			"checkingImageExists":  v.ResourceValidation.CheckingImageExists,
			"validationTimestamp":  v.ResourceValidation.ValidationTimestamp,
		}
	}
	
	return result
}
