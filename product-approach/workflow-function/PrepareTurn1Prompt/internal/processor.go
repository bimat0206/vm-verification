package internal

import (
	"strings"
	"fmt"
)

// BuildTemplateData creates the data object for template rendering
func BuildTemplateData(input *Input) (TemplateData, error) {
	vCtx := input.VerificationContext
	data := TemplateData{
		VerificationType: vCtx.VerificationType,
		VerificationID:   vCtx.VerificationID,
		VerificationAt:   vCtx.VerificationAt,
		VendingMachineID: vCtx.VendingMachineID,
		TurnNumber:       input.TurnNumber, // Always 1 for Turn1Prompt
	}
	
	// Set machine structure based on verification type
	if vCtx.VerificationType == "LAYOUT_VS_CHECKING" && input.LayoutMetadata != nil {
		ms := input.LayoutMetadata.MachineStructure
		data.MachineStructure = ms
		data.RowCount = ms.RowCount
		data.ColumnCount = ms.ColumnsPerRow
		data.RowLabels = FormatArrayToString(ms.RowOrder)
		data.ColumnLabels = FormatArrayToString(ms.ColumnOrder)
		data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
		
		// Extract product mappings if available
		if input.LayoutMetadata.ProductPositionMap != nil {
			data.ProductMappings = FormatProductMappings(input.LayoutMetadata.ProductPositionMap)
		}
		
		if input.LayoutMetadata.Location != "" {
			data.Location = input.LayoutMetadata.Location
		}
	} else if vCtx.VerificationType == "PREVIOUS_VS_CURRENT" && input.HistoricalContext != nil {
		hCtx := input.HistoricalContext
		data.PreviousVerificationID = hCtx.PreviousVerificationID
		data.PreviousVerificationAt = hCtx.PreviousVerificationAt
		data.PreviousVerificationStatus = hCtx.PreviousVerificationStatus
		data.HoursSinceLastVerification = hCtx.HoursSinceLastVerification
		
		// Set machine structure from historical context if available
		if hCtx.MachineStructure != nil {
			ms := hCtx.MachineStructure
			data.MachineStructure = ms
			data.RowCount = ms.RowCount
			data.ColumnCount = ms.ColumnsPerRow
			data.RowLabels = FormatArrayToString(ms.RowOrder)
			data.ColumnLabels = FormatArrayToString(ms.ColumnOrder)
			data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
		}
		
		// Include previous verification summary if available
		if hCtx.VerificationSummary != nil {
			data.VerificationSummary = hCtx.VerificationSummary
		}
	}
	
	return data, nil
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

// FormatRowHighlights creates a map of row labels to descriptions based on historical context
func FormatRowHighlights(checkingStatus map[string]string) map[string]string {
	if checkingStatus == nil {
		return nil
	}
	
	// Create a cleaned-up version of the checking status
	highlights := make(map[string]string)
	for row, status := range checkingStatus {
		// Extract the key information and remove verbose parts
		cleaned := strings.ReplaceAll(status, "Current: ", "")
		cleaned = strings.ReplaceAll(cleaned, "Status: ", "")
		
		// Add to highlights
		highlights[row] = cleaned
	}
	
	return highlights
}

// CreateTurn1MessageContent generates the text content for the Turn 1 message
func CreateTurn1MessageContent(verificationType string, promptText string, rowCount int, topRow, bottomRow string) string {
	if promptText != "" {
		return promptText
	}
	
	// Fallback to basic message if template didn't provide content
	base := "Please analyze the FIRST image (Reference Image)\n\n"
	
	if verificationType == "LAYOUT_VS_CHECKING" {
		base += "This image shows the approved product arrangement according to the planogram.\n\n"
	} else if verificationType == "PREVIOUS_VS_CURRENT" {
		base += "This image shows the previous state of the vending machine.\n\n"
	}
	
	base += fmt.Sprintf("Focus exclusively on analyzing this Reference Image in detail. Your goal is to identify the exact contents of all %d rows (%s-%s).\n\n", 
		rowCount, topRow, bottomRow)
	
	base += "Important reminders:\n"
	base += fmt.Sprintf("1. Row identification is CRITICAL - Row %s is ALWAYS the topmost physical shelf, Row %s is ALWAYS the bottommost physical shelf.\n", 
		topRow, bottomRow)
	base += "2. Be thorough and descriptive in your analysis of each row status (Full/Partial/Empty).\n"
	base += "3. DO NOT compare with any other image at this stage - just analyze this Reference Image.\n"
	
	return base
}

