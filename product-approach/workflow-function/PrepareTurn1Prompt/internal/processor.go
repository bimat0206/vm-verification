package internal

import (
	"fmt"
	"strings"

	"workflow-function/shared/schema"
)

// BuildTemplateData creates the data object for template rendering
func BuildTemplateData(input *schema.WorkflowState) (TemplateData, error) {
	vCtx := input.VerificationContext
	data := TemplateData{
		VerificationType: vCtx.VerificationType,
		VerificationID:   vCtx.VerificationId,
		VerificationAt:   vCtx.VerificationAt,
		VendingMachineID: vCtx.VendingMachineId,
		TurnNumber:       input.CurrentPrompt.TurnNumber,
	}
	
	// Set machine structure based on verification type
	if vCtx.VerificationType == schema.VerificationTypeLayoutVsChecking && input.LayoutMetadata != nil {
		// Convert from map to MachineStructure
		var ms *MachineStructure
		if machineStructure, ok := input.LayoutMetadata["machineStructure"].(map[string]interface{}); ok {
			rowCount, _ := machineStructure["rowCount"].(int)
			columnsPerRow, _ := machineStructure["columnsPerRow"].(int)
			
			var rowOrder []string
			if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok {
				for _, row := range rowOrderInterface {
					if rowStr, ok := row.(string); ok {
						rowOrder = append(rowOrder, rowStr)
					}
				}
			}
			
			var columnOrder []string
			if columnOrderInterface, ok := machineStructure["columnOrder"].([]interface{}); ok {
				for _, col := range columnOrderInterface {
					if colStr, ok := col.(string); ok {
						columnOrder = append(columnOrder, colStr)
					}
				}
			}
			
			ms = &MachineStructure{
				RowCount:      rowCount,
				ColumnsPerRow: columnsPerRow,
				RowOrder:      rowOrder,
				ColumnOrder:   columnOrder,
			}
		}
		
		if ms != nil {
			data.MachineStructure = ms
			data.RowCount = ms.RowCount
			data.ColumnCount = ms.ColumnsPerRow
			data.RowLabels = FormatArrayToString(ms.RowOrder)
			data.ColumnLabels = FormatArrayToString(ms.ColumnOrder)
			data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
		}
		
		// Extract product mappings if available
		if productPositionMap, ok := input.LayoutMetadata["productPositionMap"].(map[string]interface{}); ok {
			data.ProductMappings = FormatProductMappingsFromMap(productPositionMap)
		}
		
		if location, ok := input.LayoutMetadata["location"].(string); ok {
			data.Location = location
		}
	} else if vCtx.VerificationType == schema.VerificationTypePreviousVsCurrent && input.HistoricalContext != nil {
		// Extract historical context data
		if previousVerificationId, ok := input.HistoricalContext["previousVerificationId"].(string); ok {
			data.PreviousVerificationID = previousVerificationId
		}
		
		if previousVerificationAt, ok := input.HistoricalContext["previousVerificationAt"].(string); ok {
			data.PreviousVerificationAt = previousVerificationAt
		}
		
		if previousVerificationStatus, ok := input.HistoricalContext["previousVerificationStatus"].(string); ok {
			data.PreviousVerificationStatus = previousVerificationStatus
		}
		
		if hoursSinceLastVerification, ok := input.HistoricalContext["hoursSinceLastVerification"].(float64); ok {
			data.HoursSinceLastVerification = hoursSinceLastVerification
		}
		
		// Set machine structure from historical context if available
		if machineStructure, ok := input.HistoricalContext["machineStructure"].(map[string]interface{}); ok {
			rowCount, _ := machineStructure["rowCount"].(int)
			columnsPerRow, _ := machineStructure["columnsPerRow"].(int)
			
			var rowOrder []string
			if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok {
				for _, row := range rowOrderInterface {
					if rowStr, ok := row.(string); ok {
						rowOrder = append(rowOrder, rowStr)
					}
				}
			}
			
			var columnOrder []string
			if columnOrderInterface, ok := machineStructure["columnOrder"].([]interface{}); ok {
				for _, col := range columnOrderInterface {
					if colStr, ok := col.(string); ok {
						columnOrder = append(columnOrder, colStr)
					}
				}
			}
			
			ms := &MachineStructure{
				RowCount:      rowCount,
				ColumnsPerRow: columnsPerRow,
				RowOrder:      rowOrder,
				ColumnOrder:   columnOrder,
			}
			
			data.MachineStructure = ms
			data.RowCount = ms.RowCount
			data.ColumnCount = ms.ColumnsPerRow
			data.RowLabels = FormatArrayToString(ms.RowOrder)
			data.ColumnLabels = FormatArrayToString(ms.ColumnOrder)
			data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
		}
		
		// Include previous verification summary if available
		if verificationSummary, ok := input.HistoricalContext["verificationSummary"].(map[string]interface{}); ok {
			summary := &VerificationSummary{}
			
			if totalPositionsChecked, ok := verificationSummary["totalPositionsChecked"].(int); ok {
				summary.TotalPositionsChecked = totalPositionsChecked
			}
			
			if correctPositions, ok := verificationSummary["correctPositions"].(int); ok {
				summary.CorrectPositions = correctPositions
			}
			
			if discrepantPositions, ok := verificationSummary["discrepantPositions"].(int); ok {
				summary.DiscrepantPositions = discrepantPositions
			}
			
			if missingProducts, ok := verificationSummary["missingProducts"].(int); ok {
				summary.MissingProducts = missingProducts
			}
			
			if incorrectProductTypes, ok := verificationSummary["incorrectProductTypes"].(int); ok {
				summary.IncorrectProductTypes = incorrectProductTypes
			}
			
			if unexpectedProducts, ok := verificationSummary["unexpectedProducts"].(int); ok {
				summary.UnexpectedProducts = unexpectedProducts
			}
			
			if emptyPositionsCount, ok := verificationSummary["emptyPositionsCount"].(int); ok {
				summary.EmptyPositionsCount = emptyPositionsCount
			}
			
			if overallAccuracy, ok := verificationSummary["overallAccuracy"].(float64); ok {
				summary.OverallAccuracy = overallAccuracy
			}
			
			if overallConfidence, ok := verificationSummary["overallConfidence"].(float64); ok {
				summary.OverallConfidence = overallConfidence
			}
			
			if verificationStatus, ok := verificationSummary["verificationStatus"].(string); ok {
				summary.VerificationStatus = verificationStatus
			}
			
			if verificationOutcome, ok := verificationSummary["verificationOutcome"].(string); ok {
				summary.VerificationOutcome = verificationOutcome
			}
			
			data.VerificationSummary = summary
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

// FormatProductMappingsFromMap converts a product position map to a formatted array
func FormatProductMappingsFromMap(positionMap map[string]interface{}) []ProductMapping {
	if positionMap == nil {
		return []ProductMapping{}
	}
	
	mappings := make([]ProductMapping, 0, len(positionMap))
	for position, infoInterface := range positionMap {
		if info, ok := infoInterface.(map[string]interface{}); ok {
			productID, _ := info["productId"].(int)
			productName, _ := info["productName"].(string)
			
			mappings = append(mappings, ProductMapping{
				Position:    position,
				ProductID:   productID,
				ProductName: productName,
			})
		}
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
	
	if verificationType == schema.VerificationTypeLayoutVsChecking {
		base += "This image shows the approved product arrangement according to the planogram.\n\n"
	} else if verificationType == schema.VerificationTypePreviousVsCurrent {
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
