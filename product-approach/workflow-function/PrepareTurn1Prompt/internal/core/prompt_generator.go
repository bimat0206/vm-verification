package core

import (
	"fmt"
	"strings"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
	"prepare-turn1/internal/integration"
)

// PromptGenerator handles generation of the Turn 1 prompt
type PromptGenerator struct {
	log       logger.Logger
	processor *TemplateProcessor
}

// NewPromptGenerator creates a new prompt generator with the given logger and template processor
func NewPromptGenerator(log logger.Logger, processor *TemplateProcessor) *PromptGenerator {
	return &PromptGenerator{
		log:       log,
		processor: processor,
	}
}

// GeneratePrompt creates the Turn 1 prompt based on the workflow state
func (p *PromptGenerator) GeneratePrompt(state *schema.WorkflowState) (string, error) {
	// Build template data from the workflow state
	templateData, err := p.buildTemplateData(state)
	if err != nil {
		return "", fmt.Errorf("failed to build template data: %w", err)
	}

	// Determine template name based on verification type
	templateName := p.buildTemplateName(state.VerificationContext.VerificationType)

	// Process the template to generate the prompt text
	promptText, err := p.processor.ProcessTemplate(templateName, templateData)
	if err != nil {
		return "", fmt.Errorf("failed to process template: %w", err)
	}

	// If template processing failed to generate any content, use fallback
	if promptText == "" {
		p.log.Warn("Template processing returned empty text, using fallback", map[string]interface{}{
			"templateName": templateName,
		})
		promptText = p.createFallbackPrompt(state)
	}

	return promptText, nil
}

// buildTemplateData creates the data structure for template rendering
func (p *PromptGenerator) buildTemplateData(state *schema.WorkflowState) (map[string]interface{}, error) {
	if state == nil {
		return nil, errors.NewValidationError("Workflow state is nil", nil)
	}

	vCtx := state.VerificationContext
	if vCtx == nil {
		return nil, errors.NewValidationError("Verification context is nil", nil)
	}

	// Create base template data
	data := map[string]interface{}{
		"VerificationType": vCtx.VerificationType,
		"VerificationID":   vCtx.VerificationId,
		"VerificationAt":   vCtx.VerificationAt,
		"VendingMachineID": vCtx.VendingMachineId,
		"TurnNumber":       state.CurrentPrompt.TurnNumber,
	}

	// Set machine structure based on verification type
	if vCtx.VerificationType == schema.VerificationTypeLayoutVsChecking && state.LayoutMetadata != nil {
		p.addLayoutMetadataToTemplateData(state, data)
	} else if vCtx.VerificationType == schema.VerificationTypePreviousVsCurrent && state.HistoricalContext != nil {
		p.addHistoricalContextToTemplateData(state, data)
	}

	return data, nil
}

// addLayoutMetadataToTemplateData adds layout metadata to the template data
func (p *PromptGenerator) addLayoutMetadataToTemplateData(state *schema.WorkflowState, data map[string]interface{}) {
	// Extract machine structure
	if machineStructure, ok := state.LayoutMetadata["machineStructure"].(map[string]interface{}); ok {
		// Set machine structure in template data
		data["MachineStructure"] = machineStructure

		// Extract and set basic machine properties
		if rowCount, ok := machineStructure["rowCount"].(int); ok {
			data["RowCount"] = rowCount
		}

		if columnsPerRow, ok := machineStructure["columnsPerRow"].(int); ok {
			data["ColumnCount"] = columnsPerRow
		}

		// Format row and column labels
		if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok {
			rowOrder := make([]string, 0, len(rowOrderInterface))
			for _, row := range rowOrderInterface {
				if rowStr, ok := row.(string); ok {
					rowOrder = append(rowOrder, rowStr)
				}
			}
			data["RowLabels"] = integration.FormatArrayToString(rowOrder)
		}

		if columnOrderInterface, ok := machineStructure["columnOrder"].([]interface{}); ok {
			columnOrder := make([]string, 0, len(columnOrderInterface))
			for _, col := range columnOrderInterface {
				if colStr, ok := col.(string); ok {
					columnOrder = append(columnOrder, colStr)
				}
			}
			data["ColumnLabels"] = integration.FormatArrayToString(columnOrder)
		}

		// Calculate total positions
		if rowCount, ok := data["RowCount"].(int); ok {
			if columnCount, ok := data["ColumnCount"].(int); ok {
				data["TotalPositions"] = rowCount * columnCount
			}
		}
	}

	// Extract product mappings
	if productPositionMap, ok := state.LayoutMetadata["productPositionMap"].(map[string]interface{}); ok {
		data["ProductMappings"] = p.formatProductMappings(productPositionMap)
	}

	// Extract location
	if location, ok := state.LayoutMetadata["location"].(string); ok {
		data["Location"] = location
	}
}

// addHistoricalContextToTemplateData adds historical context to the template data
func (p *PromptGenerator) addHistoricalContextToTemplateData(state *schema.WorkflowState, data map[string]interface{}) {
	// Extract previous verification information
	if previousVerificationId, ok := state.HistoricalContext["previousVerificationId"].(string); ok {
		data["PreviousVerificationID"] = previousVerificationId
	}

	if previousVerificationAt, ok := state.HistoricalContext["previousVerificationAt"].(string); ok {
		data["PreviousVerificationAt"] = previousVerificationAt
	}

	if previousVerificationStatus, ok := state.HistoricalContext["previousVerificationStatus"].(string); ok {
		data["PreviousVerificationStatus"] = previousVerificationStatus
	}

	if hoursSinceLastVerification, ok := state.HistoricalContext["hoursSinceLastVerification"].(float64); ok {
		data["HoursSinceLastVerification"] = hoursSinceLastVerification
	}

	// Extract machine structure
	if machineStructure, ok := state.HistoricalContext["machineStructure"].(map[string]interface{}); ok {
		// Set machine structure in template data
		data["MachineStructure"] = machineStructure

		// Extract and set basic machine properties
		if rowCount, ok := machineStructure["rowCount"].(int); ok {
			data["RowCount"] = rowCount
		}

		if columnsPerRow, ok := machineStructure["columnsPerRow"].(int); ok {
			data["ColumnCount"] = columnsPerRow
		}

		// Format row and column labels
		if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok {
			rowOrder := make([]string, 0, len(rowOrderInterface))
			for _, row := range rowOrderInterface {
				if rowStr, ok := row.(string); ok {
					rowOrder = append(rowOrder, rowStr)
				}
			}
			data["RowLabels"] = integration.FormatArrayToString(rowOrder)
		}

		if columnOrderInterface, ok := machineStructure["columnOrder"].([]interface{}); ok {
			columnOrder := make([]string, 0, len(columnOrderInterface))
			for _, col := range columnOrderInterface {
				if colStr, ok := col.(string); ok {
					columnOrder = append(columnOrder, colStr)
				}
			}
			data["ColumnLabels"] = integration.FormatArrayToString(columnOrder)
		}

		// Calculate total positions
		if rowCount, ok := data["RowCount"].(int); ok {
			if columnCount, ok := data["ColumnCount"].(int); ok {
				data["TotalPositions"] = rowCount * columnCount
			}
		}
	}

	// Extract verification summary
	if verificationSummary, ok := state.HistoricalContext["verificationSummary"].(map[string]interface{}); ok {
		data["VerificationSummary"] = verificationSummary
	}
}

// formatProductMappings formats product mappings for the template
func (p *PromptGenerator) formatProductMappings(positionMap map[string]interface{}) []map[string]interface{} {
	if positionMap == nil {
		return []map[string]interface{}{}
	}

	mappings := make([]map[string]interface{}, 0, len(positionMap))
	for position, infoInterface := range positionMap {
		if info, ok := infoInterface.(map[string]interface{}); ok {
			productID, _ := info["productId"].(int)
			productName, _ := info["productName"].(string)

			mappings = append(mappings, map[string]interface{}{
				"Position":    position,
				"ProductID":   productID,
				"ProductName": productName,
			})
		}
	}

	return mappings
}

// buildTemplateName constructs the template name based on verification type
func (p *PromptGenerator) buildTemplateName(verificationType string) string {
	// Convert LAYOUT_VS_CHECKING to layout-vs-checking (replace underscores with hyphens)
	formattedType := strings.ReplaceAll(strings.ToLower(verificationType), "_", "-")
	return fmt.Sprintf("turn1-%s", formattedType)
}

// BuildTemplateName is the exported version of buildTemplateName
func (p *PromptGenerator) BuildTemplateName(verificationType string) string {
	return p.buildTemplateName(verificationType)
}

// createFallbackPrompt generates a basic Turn 1 prompt if template processing fails
func (p *PromptGenerator) createFallbackPrompt(state *schema.WorkflowState) string {
	// Basic fallback text
	verificationType := state.VerificationContext.VerificationType
	rowCount := 0
	topRow := ""
	bottomRow := ""

	// Get machine structure information if available
	if verificationType == schema.VerificationTypeLayoutVsChecking && state.LayoutMetadata != nil {
		if machineStructure, ok := state.LayoutMetadata["machineStructure"].(map[string]interface{}); ok {
			if count, ok := machineStructure["rowCount"].(int); ok {
				rowCount = count
			}

			if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok && len(rowOrderInterface) > 0 {
				if rowStr, ok := rowOrderInterface[0].(string); ok {
					topRow = rowStr
				}
				if len(rowOrderInterface) > 1 {
					if rowStr, ok := rowOrderInterface[len(rowOrderInterface)-1].(string); ok {
						bottomRow = rowStr
					}
				}
			}
		}
	} else if verificationType == schema.VerificationTypePreviousVsCurrent && state.HistoricalContext != nil {
		if machineStructure, ok := state.HistoricalContext["machineStructure"].(map[string]interface{}); ok {
			if count, ok := machineStructure["rowCount"].(int); ok {
				rowCount = count
			}

			if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok && len(rowOrderInterface) > 0 {
				if rowStr, ok := rowOrderInterface[0].(string); ok {
					topRow = rowStr
				}
				if len(rowOrderInterface) > 1 {
					if rowStr, ok := rowOrderInterface[len(rowOrderInterface)-1].(string); ok {
						bottomRow = rowStr
					}
				}
			}
		}
	}

	// Fallback to sensible defaults if we couldn't extract the information
	if rowCount == 0 {
		rowCount = 5
	}
	if topRow == "" {
		topRow = "A"
	}
	if bottomRow == "" {
		bottomRow = "E"
	}

	// Create basic prompt text
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