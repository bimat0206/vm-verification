package models

import (
	"encoding/json"
	"strings"
)

// MachineStructure describes the vending machine physical layout
type MachineStructure struct {
	RowCount      int      `json:"rowCount"`
	ColumnsPerRow int      `json:"columnsPerRow"`
	RowOrder      []string `json:"rowOrder"`
	ColumnOrder   []string `json:"columnOrder"`
}

// ProductInfo contains product details for a specific position
type ProductInfo struct {
	ProductID    int    `json:"productId"`
	ProductName  string `json:"productName"`
	ProductImage string `json:"productImage,omitempty"`
}

// ProductMapping represents a formatted product mapping for templates
type ProductMapping struct {
	Position    string `json:"position"`
	ProductID   int    `json:"productId"`
	ProductName string `json:"productName"`
}

// VerificationSummary contains summary statistics from a verification
type VerificationSummary struct {
	TotalPositionsChecked  int     `json:"totalPositionsChecked"`
	CorrectPositions       int     `json:"correctPositions"`
	DiscrepantPositions    int     `json:"discrepantPositions"`
	MissingProducts        int     `json:"missingProducts"`
	IncorrectProductTypes  int     `json:"incorrectProductTypes"`
	UnexpectedProducts     int     `json:"unexpectedProducts"`
	EmptyPositionsCount    int     `json:"emptyPositionsCount"`
	OverallAccuracy        float64 `json:"overallAccuracy"`
	OverallConfidence      float64 `json:"overallConfidence"`
	VerificationStatus     string  `json:"verificationStatus"`
	VerificationOutcome    string  `json:"verificationOutcome"`
}

// TemplateData contains all data needed for template rendering
type TemplateData struct {
	// Common verification data
	VerificationType   string              `json:"verificationType"`
	VerificationID     string              `json:"verificationId"`
	VerificationAt     string              `json:"verificationAt"`
	VendingMachineID   string              `json:"vendingMachineId"`
	Location           string              `json:"location,omitempty"`
	
	// Machine structure data
	MachineStructure   *MachineStructure   `json:"machineStructure,omitempty"`
	RowCount           int                 `json:"rowCount"`
	ColumnCount        int                 `json:"columnCount"`
	RowLabels          string              `json:"rowLabels"`
	ColumnLabels       string              `json:"columnLabels"`
	TotalPositions     int                 `json:"totalPositions"`
	
	// Layout-specific data
	ProductMappings    []ProductMapping    `json:"productMappings,omitempty"`
	
	// Historical context data
	PreviousVerificationID     string      `json:"previousVerificationId,omitempty"`
	PreviousVerificationAt     string      `json:"previousVerificationAt,omitempty"`
	PreviousVerificationStatus string      `json:"previousVerificationStatus,omitempty"`
	HoursSinceLastVerification float64     `json:"hoursSinceLastVerification,omitempty"`
	VerificationSummary        *VerificationSummary `json:"verificationSummary,omitempty"`
}

// ExtractMachineStructure extracts machine structure from layout metadata
func ExtractMachineStructure(layoutMetadata map[string]interface{}) (*MachineStructure, error) {
	if layoutMetadata == nil {
		return nil, nil
	}
	
	msData, ok := layoutMetadata["machineStructure"]
	if !ok {
		return nil, nil
	}
	
	// Convert to JSON and then to struct
	msBytes, err := json.Marshal(msData)
	if err != nil {
		return nil, err
	}
	
	var ms MachineStructure
	if err := json.Unmarshal(msBytes, &ms); err != nil {
		return nil, err
	}
	
	return &ms, nil
}

// ExtractProductPositionMap extracts product position map from layout metadata
func ExtractProductPositionMap(layoutMetadata map[string]interface{}) (map[string]ProductInfo, error) {
	if layoutMetadata == nil {
		return nil, nil
	}
	
	ppmData, ok := layoutMetadata["productPositionMap"]
	if !ok {
		return nil, nil
	}
	
	// Convert to JSON and then to struct
	ppmBytes, err := json.Marshal(ppmData)
	if err != nil {
		return nil, err
	}
	
	var ppm map[string]ProductInfo
	if err := json.Unmarshal(ppmBytes, &ppm); err != nil {
		return nil, err
	}
	
	return ppm, nil
}

// ExtractVerificationSummary extracts verification summary from historical context
func ExtractVerificationSummary(historicalContext map[string]interface{}) (*VerificationSummary, error) {
	if historicalContext == nil {
		return nil, nil
	}
	
	vsData, ok := historicalContext["verificationSummary"]
	if !ok {
		return nil, nil
	}
	
	// Convert to JSON and then to struct
	vsBytes, err := json.Marshal(vsData)
	if err != nil {
		return nil, err
	}
	
	var vs VerificationSummary
	if err := json.Unmarshal(vsBytes, &vs); err != nil {
		return nil, err
	}
	
	return &vs, nil
}

// FormatArrayToString formats a string array to a comma-separated string
func FormatArrayToString(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return strings.Join(arr, ", ")
}

// FormatProductMappings converts a product position map to a formatted array
func FormatProductMappings(positionMap map[string]ProductInfo) []ProductMapping {
	if positionMap == nil {
		return []ProductMapping{}
	}
	
	mappings := make([]ProductMapping, 0, len(positionMap))
	for position, info := range positionMap {
		mappings = append(mappings, ProductMapping{
			Position:    position,
			ProductID:   info.ProductID,
			ProductName: info.ProductName,
		})
	}
	
	return mappings
}