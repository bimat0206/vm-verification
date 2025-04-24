package engines

import (
	"context"
	//"fmt"
	"time"

	"verification-service/internal/domain/models"
)

// AIProvider interface for using in VerificationEngine
type AIProvider interface {
	InvokeModel(
		ctx context.Context,
		systemPrompt string,
		userPrompt string,
		images []string, // Base64 encoded
		conversationContext map[string]interface{},
	) (string, map[string]interface{}, error)
}

// VerificationEngine implements the domain.VerificationEngine interface
type VerificationEngine struct {
	aiProvider AIProvider
}

// NewVerificationEngine creates a new verification engine
func NewVerificationEngine(aiProvider AIProvider) *VerificationEngine {
	return &VerificationEngine{
		aiProvider: aiProvider,
	}
}

// AnalyzeReferenceLayout processes the reference layout image (Turn 1)
func (e *VerificationEngine) AnalyzeReferenceLayout(
	ctx context.Context,
	verificationContext models.VerificationContext,
	referenceImage []byte,
	layoutMetadata map[string]interface{},
) (*models.ReferenceAnalysis, error) {
	// In a real implementation, this would use the AI provider to analyze the image
	// For this implementation, we'll return a simplified result
	
	// Create basic machine structure from metadata
	machineStructure := models.MachineStructure{
		RowCount:      6,
		ColumnsPerRow: 10,
		RowOrder:      []string{"A", "B", "C", "D", "E", "F"},
		ColumnOrder:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
	}
	
	// Physical orientation
	machineStructure.PhysicalOrientation.TopRow = "A"
	machineStructure.PhysicalOrientation.LeftColumn = "1"
	machineStructure.PhysicalOrientation.RowDirection = "topToBottom"
	machineStructure.PhysicalOrientation.ColumnDirection = "leftToRight"
	
	// Extract or use layoutMetadata to populate machine structure if available
	if meta, ok := layoutMetadata["machineStructure"].(map[string]interface{}); ok {
		if rowCount, ok := meta["rowCount"].(float64); ok {
			machineStructure.RowCount = int(rowCount)
		}
		if columnsPerRow, ok := meta["columnsPerRow"].(float64); ok {
			machineStructure.ColumnsPerRow = int(columnsPerRow)
		}
		// Additional mapping from metadata would happen here
	}
	
	// Return analysis
	return &models.ReferenceAnalysis{
		TurnNumber:         1,
		MachineStructure:   machineStructure,
		RowAnalysis:        make(map[string]map[string]interface{}),
		ProductPositions:   make(map[string]map[string]interface{}),
		EmptyPositions:     []string{},
		Confidence:         90,
		InitialConfirmation: "Successfully identified machine structure.",
		CompletedAt:        time.Now(),
	}, nil
}

// VerifyCheckingImage compares the checking image to the reference layout (Turn 2)
func (e *VerificationEngine) VerifyCheckingImage(
	ctx context.Context,
	verificationContext models.VerificationContext, 
	checkingImage []byte,
	referenceAnalysis *models.ReferenceAnalysis,
) (*models.CheckingAnalysis, error) {
	// In a real implementation, this would use the AI provider to analyze the image
	// For this implementation, we'll return a simplified result
	
	// Create mock discrepancies for demonstration
	discrepancies := []models.Discrepancy{
		{
			Position:          "A01",
			Expected:          "Mi Hảo Hảo",
			Found:             "Mi modern Lẩu thái",
			Issue:             models.DiscrepancyIncorrectProductType,
			Confidence:        95,
			Evidence:          "Different packaging color and branding visible",
			VerificationResult: models.StatusIncorrect,
			Severity:          models.SeverityHigh,
		},
	}
	
	// Create empty slot report
	emptySlotReport := struct {
		ReferenceEmptyRows      []string `json:"referenceEmptyRows"`
		CheckingEmptyRows       []string `json:"checkingEmptyRows"`
		CheckingPartiallyEmptyRows []string `json:"checkingPartiallyEmptyRows"`
		CheckingEmptyPositions  []string `json:"checkingEmptyPositions"`
		TotalEmpty              int      `json:"totalEmpty"`
	}{
		ReferenceEmptyRows:      []string{},
		CheckingEmptyRows:       []string{"F"},
		CheckingPartiallyEmptyRows: []string{},
		CheckingEmptyPositions:  []string{"F01", "F02", "F03", "F04", "F05", "F06", "F07"},
		TotalEmpty:              7,
	}
	
	// Return analysis
	return &models.CheckingAnalysis{
		TurnNumber:        2,
		VerificationStatus: models.StatusIncorrect,
		Discrepancies:     discrepancies,
		TotalDiscrepancies: len(discrepancies),
		Severity:          models.SeverityHigh,
		RowAnalysis:       make(map[string]map[string]interface{}),
		EmptySlotReport:   emptySlotReport,
		Confidence:        95,
		CompletedAt:       time.Now(),
	}, nil
}