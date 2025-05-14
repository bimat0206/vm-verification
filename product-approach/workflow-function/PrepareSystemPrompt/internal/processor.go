package internal

import (
	//"strings"

	"workflow-function/shared/schema"
)

// BuildTemplateData creates the data object for template rendering
func BuildTemplateData(input *Input) (TemplateData, error) {
	vCtx := input.VerificationContext
	data := TemplateData{
		VerificationType: vCtx.VerificationType,
		VerificationID:   vCtx.VerificationId,
		VerificationAt:   vCtx.VerificationAt,
		VendingMachineID: vCtx.VendingMachineId,
	}
	
	// Set machine structure based on verification type
	if vCtx.VerificationType == schema.VerificationTypeLayoutVsChecking && input.LayoutMetadata != nil {
		// Extract machine structure from layout metadata
		ms, err := ExtractMachineStructure(input.LayoutMetadata)
		if err != nil || ms == nil {
			// Log warning but continue
			// Note: Validation should have caught this already
			return data, nil
		}
		
		data.MachineStructure = ms
		data.RowCount = ms.RowCount
		data.ColumnCount = ms.ColumnsPerRow
		data.RowLabels = FormatArrayToString(ms.RowOrder)
		data.ColumnLabels = FormatArrayToString(ms.ColumnOrder)
		data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
		
		// Extract product mappings if available
		productMap, err := ExtractProductPositionMap(input.LayoutMetadata)
		if err == nil && productMap != nil {
			data.ProductMappings = FormatProductMappings(productMap)
		}
		
		// Extract location if available
		if locVal, ok := input.LayoutMetadata["location"]; ok {
			if loc, ok := locVal.(string); ok {
				data.Location = loc
			}
		}
	} else if vCtx.VerificationType == schema.VerificationTypePreviousVsCurrent && input.HistoricalContext != nil {
		// Extract previous verification details
		if prevId, ok := input.HistoricalContext["previousVerificationId"]; ok {
			if idStr, ok := prevId.(string); ok {
				data.PreviousVerificationID = idStr
			}
		}
		
		if prevAt, ok := input.HistoricalContext["previousVerificationAt"]; ok {
			if atStr, ok := prevAt.(string); ok {
				data.PreviousVerificationAt = atStr
			}
		}
		
		if prevStatus, ok := input.HistoricalContext["previousVerificationStatus"]; ok {
			if statusStr, ok := prevStatus.(string); ok {
				data.PreviousVerificationStatus = statusStr
			}
		}
		
		if hoursVal, ok := input.HistoricalContext["hoursSinceLastVerification"]; ok {
			if hours, ok := hoursVal.(float64); ok {
				data.HoursSinceLastVerification = hours
			}
		}
		
		// Extract machine structure from historical context if available
		if msVal, ok := input.HistoricalContext["machineStructure"]; ok {
			// Convert to JSON and extract machine structure
			msData, err := ExtractMachineStructure(map[string]interface{}{"machineStructure": msVal})
			if err == nil && msData != nil {
				data.MachineStructure = msData
				data.RowCount = msData.RowCount
				data.ColumnCount = msData.ColumnsPerRow
				data.RowLabels = FormatArrayToString(msData.RowOrder)
				data.ColumnLabels = FormatArrayToString(msData.ColumnOrder)
				data.TotalPositions = msData.RowCount * msData.ColumnsPerRow
			}
		}
		
		// Extract verification summary if available
		vs, err := ExtractVerificationSummary(input.HistoricalContext)
		if err == nil && vs != nil {
			data.VerificationSummary = vs
		}
	}
	
	return data, nil
}

// FormatArrayToString formats a string array to a comma-separated string
// (moved from utils.go for clarity)

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

