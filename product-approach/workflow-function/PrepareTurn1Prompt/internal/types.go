package internal

import (
	"text/template"
	
	"workflow-function/shared/schema"
)

// Input represents the Lambda input event
type Input struct {
	VerificationContext *VerificationContext `json:"verificationContext"`
	LayoutMetadata      *LayoutMetadata      `json:"layoutMetadata,omitempty"`
	HistoricalContext   *HistoricalContext   `json:"historicalContext,omitempty"`
	SystemPrompt        *SystemPrompt        `json:"systemPrompt,omitempty"`
	BedrockConfig       *BedrockConfig       `json:"bedrockConfig,omitempty"`
	TurnNumber          int                  `json:"turnNumber"`
	IncludeImage        string               `json:"includeImage"`
	Images              *ImageData           `json:"images,omitempty"`
}

// Response represents the Lambda response
type Response struct {
	VerificationContext *VerificationContext `json:"verificationContext"`
	LayoutMetadata      *LayoutMetadata      `json:"layoutMetadata,omitempty"`
	HistoricalContext   *HistoricalContext   `json:"historicalContext,omitempty"`
	CurrentPrompt       CurrentPrompt        `json:"currentPrompt"`
	BedrockConfig       BedrockConfig        `json:"bedrockConfig"`
}

// VerificationContext contains verification metadata
type VerificationContext struct {
	VerificationID       string `json:"verificationId"`
	VerificationAt       string `json:"verificationAt"`
	Status               string `json:"status"`
	VerificationType     string `json:"verificationType"`
	VendingMachineID     string `json:"vendingMachineId,omitempty"`
	LayoutID             int    `json:"layoutId,omitempty"`
	LayoutPrefix         string `json:"layoutPrefix,omitempty"`
	ReferenceImageURL    string `json:"referenceImageUrl,omitempty"`
	CheckingImageURL     string `json:"checkingImageUrl,omitempty"`
	NotificationEnabled  bool   `json:"notificationEnabled,omitempty"`
}

// LayoutMetadata contains layout-specific information
type LayoutMetadata struct {
	MachineStructure    *MachineStructure        `json:"machineStructure"`
	ProductPositionMap  map[string]ProductInfo   `json:"productPositionMap,omitempty"`
	RowProductMapping   map[string]interface{}   `json:"rowProductMapping,omitempty"`
	Location            string                   `json:"location,omitempty"`
}

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

// HistoricalContext contains previous verification data
type HistoricalContext struct {
	PreviousVerificationID     string             `json:"previousVerificationId"`
	PreviousVerificationAt     string             `json:"previousVerificationAt"`
	PreviousVerificationStatus string             `json:"previousVerificationStatus"`
	HoursSinceLastVerification float64            `json:"hoursSinceLastVerification"`
	MachineStructure           *MachineStructure  `json:"machineStructure,omitempty"`
	CheckingStatus             map[string]string  `json:"checkingStatus,omitempty"`
	VerificationSummary        *VerificationSummary `json:"verificationSummary,omitempty"`
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

// ImageData contains image information
type ImageData struct {
	ReferenceImageBase64 string                 `json:"referenceImageBase64,omitempty"`
	CheckingImageBase64  string                 `json:"checkingImageBase64,omitempty"`
	ReferenceImageMeta   *ImageMetadata         `json:"referenceImageMeta,omitempty"`
	CheckingImageMeta    *ImageMetadata         `json:"checkingImageMeta,omitempty"`
}

// ImageMetadata contains metadata about an image in S3
type ImageMetadata struct {
	ContentType   string `json:"contentType,omitempty"`
	Size          int64  `json:"size,omitempty"`
	LastModified  string `json:"lastModified,omitempty"`
	ETag          string `json:"etag,omitempty"`
	BucketOwner   string `json:"bucketOwner,omitempty"`
	Bucket        string `json:"bucket,omitempty"`
	Key           string `json:"key,omitempty"`
	StorageMethod    string `json:"storageMethod,omitempty"`
    Base64S3Bucket   string `json:"base64S3Bucket,omitempty"`
    Base64S3Key      string `json:"base64S3Key,omitempty"`
}

// SystemPrompt represents the generated system prompt
type SystemPrompt struct {
	Content       string `json:"content"`
	PromptID      string `json:"promptId"`
	CreatedAt     string `json:"createdAt"`
	PromptVersion string `json:"promptVersion"`
}

// CurrentPrompt represents the current turn's prompt
type CurrentPrompt struct {
	Messages      []schema.BedrockMessage `json:"messages"`
	TurnNumber    int                     `json:"turnNumber"`
	PromptID      string                  `json:"promptId"`
	CreatedAt     string                  `json:"createdAt"`
	PromptVersion string                  `json:"promptVersion"`
	ImageIncluded string                  `json:"imageIncluded"`
}

// BedrockConfig contains configuration for the Bedrock API
type BedrockConfig struct {
	AnthropicVersion string         `json:"anthropic_version"`
	MaxTokens        int            `json:"max_tokens"`
	Thinking         ThinkingConfig `json:"thinking"`
}

// ThinkingConfig configures Claude's thinking process
type ThinkingConfig struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

// ProductMapping represents a formatted product mapping for templates
type ProductMapping struct {
	Position    string `json:"position"`
	ProductID   int    `json:"productId"`
	ProductName string `json:"productName"`
}

// TemplateData contains all data needed for template rendering
type TemplateData struct {
	// Common verification data
	VerificationType   string              `json:"verificationType"`
	VerificationID     string              `json:"verificationId"`
	VerificationAt     string              `json:"verificationAt"`
	VendingMachineID   string              `json:"vendingMachineId"`
	Location           string              `json:"location,omitempty"`
	
	// Turn-specific data
	TurnNumber         int                 `json:"turnNumber"`
	
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

// TemplateManager handles loading and caching templates
type TemplateManager struct {
	baseDir         string
	templates       map[string]*template.Template
	turn1Templates  map[string]*template.Template
	versions        map[string]string
	turn1Versions   map[string]string
}

// ConvertToWorkflowState converts the Input to a schema.WorkflowState
func ConvertToWorkflowState(input *Input) *schema.WorkflowState {
	if input == nil {
		return nil
	}
	
	state := &schema.WorkflowState{
		SchemaVersion: schema.SchemaVersion,
	}
	
	// Convert VerificationContext
	if input.VerificationContext != nil {
		state.VerificationContext = &schema.VerificationContext{
			VerificationId:   input.VerificationContext.VerificationID,
			VerificationAt:   input.VerificationContext.VerificationAt,
			Status:           input.VerificationContext.Status,
			VerificationType: input.VerificationContext.VerificationType,
			VendingMachineId: input.VerificationContext.VendingMachineID,
			LayoutId:         input.VerificationContext.LayoutID,
			LayoutPrefix:     input.VerificationContext.LayoutPrefix,
			ReferenceImageUrl: input.VerificationContext.ReferenceImageURL,
			CheckingImageUrl:  input.VerificationContext.CheckingImageURL,
			NotificationEnabled: input.VerificationContext.NotificationEnabled,
		}
	}
	
	// Convert LayoutMetadata to map[string]interface{}
	if input.LayoutMetadata != nil {
		state.LayoutMetadata = make(map[string]interface{})
		
		if input.LayoutMetadata.MachineStructure != nil {
			state.LayoutMetadata["machineStructure"] = map[string]interface{}{
				"rowCount":      input.LayoutMetadata.MachineStructure.RowCount,
				"columnsPerRow": input.LayoutMetadata.MachineStructure.ColumnsPerRow,
				"rowOrder":      input.LayoutMetadata.MachineStructure.RowOrder,
				"columnOrder":   input.LayoutMetadata.MachineStructure.ColumnOrder,
			}
		}
		
		if input.LayoutMetadata.ProductPositionMap != nil {
			productMap := make(map[string]interface{})
			for pos, info := range input.LayoutMetadata.ProductPositionMap {
				productMap[pos] = map[string]interface{}{
					"productId":   info.ProductID,
					"productName": info.ProductName,
				}
				if info.ProductImage != "" {
					productMap[pos].(map[string]interface{})["productImage"] = info.ProductImage
				}
			}
			state.LayoutMetadata["productPositionMap"] = productMap
		}
		
		if input.LayoutMetadata.RowProductMapping != nil {
			state.LayoutMetadata["rowProductMapping"] = input.LayoutMetadata.RowProductMapping
		}
		
		if input.LayoutMetadata.Location != "" {
			state.LayoutMetadata["location"] = input.LayoutMetadata.Location
		}
	}
	
	// Convert HistoricalContext to map[string]interface{}
	if input.HistoricalContext != nil {
		state.HistoricalContext = make(map[string]interface{})
		
		state.HistoricalContext["previousVerificationId"] = input.HistoricalContext.PreviousVerificationID
		state.HistoricalContext["previousVerificationAt"] = input.HistoricalContext.PreviousVerificationAt
		state.HistoricalContext["previousVerificationStatus"] = input.HistoricalContext.PreviousVerificationStatus
		state.HistoricalContext["hoursSinceLastVerification"] = input.HistoricalContext.HoursSinceLastVerification
		
		if input.HistoricalContext.MachineStructure != nil {
			state.HistoricalContext["machineStructure"] = map[string]interface{}{
				"rowCount":      input.HistoricalContext.MachineStructure.RowCount,
				"columnsPerRow": input.HistoricalContext.MachineStructure.ColumnsPerRow,
				"rowOrder":      input.HistoricalContext.MachineStructure.RowOrder,
				"columnOrder":   input.HistoricalContext.MachineStructure.ColumnOrder,
			}
		}
		
		if input.HistoricalContext.CheckingStatus != nil {
			state.HistoricalContext["checkingStatus"] = input.HistoricalContext.CheckingStatus
		}
		
		if input.HistoricalContext.VerificationSummary != nil {
			vs := input.HistoricalContext.VerificationSummary
			state.HistoricalContext["verificationSummary"] = map[string]interface{}{
				"totalPositionsChecked": vs.TotalPositionsChecked,
				"correctPositions":      vs.CorrectPositions,
				"discrepantPositions":   vs.DiscrepantPositions,
				"missingProducts":       vs.MissingProducts,
				"incorrectProductTypes": vs.IncorrectProductTypes,
				"unexpectedProducts":    vs.UnexpectedProducts,
				"emptyPositionsCount":   vs.EmptyPositionsCount,
				"overallAccuracy":       vs.OverallAccuracy,
				"overallConfidence":     vs.OverallConfidence,
				"verificationStatus":    vs.VerificationStatus,
				"verificationOutcome":   vs.VerificationOutcome,
			}
		}
	}
	
	// Convert SystemPrompt
	if input.SystemPrompt != nil {
		state.SystemPrompt = &schema.SystemPrompt{
			Content:       input.SystemPrompt.Content,
			PromptId:      input.SystemPrompt.PromptID,
			PromptVersion: input.SystemPrompt.PromptVersion,
		}
	}
	
	// Convert BedrockConfig
	if input.BedrockConfig != nil {
		state.BedrockConfig = &schema.BedrockConfig{
			AnthropicVersion: input.BedrockConfig.AnthropicVersion,
			MaxTokens:        input.BedrockConfig.MaxTokens,
			Thinking: &schema.Thinking{
				Type:         input.BedrockConfig.Thinking.Type,
				BudgetTokens: input.BedrockConfig.Thinking.BudgetTokens,
			},
		}
	}
	
	// Set CurrentPrompt
	state.CurrentPrompt = &schema.CurrentPrompt{
		TurnNumber:    input.TurnNumber,
		IncludeImage:  input.IncludeImage,
	}
	
	// Convert Images
	if input.Images != nil {
		state.Images = &schema.ImageData{}
		
		// Convert reference image
		if input.Images.ReferenceImageMeta != nil {
			refMeta := input.Images.ReferenceImageMeta
			state.Images.Reference = &schema.ImageInfo{
				URL:          input.VerificationContext.ReferenceImageURL,
				S3Bucket:     refMeta.Bucket,
				S3Key:        refMeta.Key,
				ContentType:  refMeta.ContentType,
				Size:         refMeta.Size,
				LastModified: refMeta.LastModified,
				ETag:         refMeta.ETag,
			}
			
			if input.Images.ReferenceImageBase64 != "" {
				state.Images.Reference.Base64Data = input.Images.ReferenceImageBase64
				state.Images.Reference.Base64Generated = true
				state.Images.Reference.StorageMethod = schema.StorageMethodInline
			}
			
			// Also set in legacy format
			state.Images.ReferenceImage = state.Images.Reference
		}
		
		// Convert checking image
		if input.Images.CheckingImageMeta != nil {
			checkMeta := input.Images.CheckingImageMeta
			state.Images.Checking = &schema.ImageInfo{
				URL:          input.VerificationContext.CheckingImageURL,
				S3Bucket:     checkMeta.Bucket,
				S3Key:        checkMeta.Key,
				ContentType:  checkMeta.ContentType,
				Size:         checkMeta.Size,
				LastModified: checkMeta.LastModified,
				ETag:         checkMeta.ETag,
			}
			
			if input.Images.CheckingImageBase64 != "" {
				state.Images.Checking.Base64Data = input.Images.CheckingImageBase64
				state.Images.Checking.Base64Generated = true
				state.Images.Checking.StorageMethod = schema.StorageMethodInline
			}
			
			// Also set in legacy format
			state.Images.CheckingImage = state.Images.Checking
		}
	}
	
	return state
}

// ConvertToResponse converts a schema.WorkflowState to a Response
func ConvertToResponse(state *schema.WorkflowState) *Response {
	if state == nil {
		return nil
	}
	
	response := &Response{}
	
	// Convert VerificationContext
	if state.VerificationContext != nil {
		response.VerificationContext = &VerificationContext{
			VerificationID:      state.VerificationContext.VerificationId,
			VerificationAt:      state.VerificationContext.VerificationAt,
			Status:              state.VerificationContext.Status,
			VerificationType:    state.VerificationContext.VerificationType,
			VendingMachineID:    state.VerificationContext.VendingMachineId,
			LayoutID:            state.VerificationContext.LayoutId,
			LayoutPrefix:        state.VerificationContext.LayoutPrefix,
			ReferenceImageURL:   state.VerificationContext.ReferenceImageUrl,
			CheckingImageURL:    state.VerificationContext.CheckingImageUrl,
			NotificationEnabled: state.VerificationContext.NotificationEnabled,
		}
	}
	
	// Convert CurrentPrompt
	if state.CurrentPrompt != nil {
		response.CurrentPrompt = CurrentPrompt{
			Messages:      state.CurrentPrompt.Messages,
			TurnNumber:    state.CurrentPrompt.TurnNumber,
			PromptID:      state.CurrentPrompt.PromptId,
			CreatedAt:     state.CurrentPrompt.CreatedAt,
			PromptVersion: state.CurrentPrompt.PromptVersion,
			ImageIncluded: state.CurrentPrompt.IncludeImage,
		}
	}
	
	// Convert BedrockConfig
	if state.BedrockConfig != nil {
		response.BedrockConfig = BedrockConfig{
			AnthropicVersion: state.BedrockConfig.AnthropicVersion,
			MaxTokens:        state.BedrockConfig.MaxTokens,
		}
		
		if state.BedrockConfig.Thinking != nil {
			response.BedrockConfig.Thinking = ThinkingConfig{
				Type:         state.BedrockConfig.Thinking.Type,
				BudgetTokens: state.BedrockConfig.Thinking.BudgetTokens,
			}
		}
	}
	
	// Convert LayoutMetadata if available
	if state.LayoutMetadata != nil {
		response.LayoutMetadata = &LayoutMetadata{}
		
		// Extract machine structure
		if machineStructure, ok := state.LayoutMetadata["machineStructure"].(map[string]interface{}); ok {
			ms := &MachineStructure{}
			
			if rowCount, ok := machineStructure["rowCount"].(int); ok {
				ms.RowCount = rowCount
			}
			
			if columnsPerRow, ok := machineStructure["columnsPerRow"].(int); ok {
				ms.ColumnsPerRow = columnsPerRow
			}
			
			if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok {
				rowOrder := make([]string, 0, len(rowOrderInterface))
				for _, row := range rowOrderInterface {
					if rowStr, ok := row.(string); ok {
						rowOrder = append(rowOrder, rowStr)
					}
				}
				ms.RowOrder = rowOrder
			}
			
			if columnOrderInterface, ok := machineStructure["columnOrder"].([]interface{}); ok {
				columnOrder := make([]string, 0, len(columnOrderInterface))
				for _, col := range columnOrderInterface {
					if colStr, ok := col.(string); ok {
						columnOrder = append(columnOrder, colStr)
					}
				}
				ms.ColumnOrder = columnOrder
			}
			
			response.LayoutMetadata.MachineStructure = ms
		}
		
		// Extract product position map
		if productPositionMap, ok := state.LayoutMetadata["productPositionMap"].(map[string]interface{}); ok {
			response.LayoutMetadata.ProductPositionMap = make(map[string]ProductInfo)
			
			for pos, infoInterface := range productPositionMap {
				if info, ok := infoInterface.(map[string]interface{}); ok {
					productInfo := ProductInfo{}
					
					if productID, ok := info["productId"].(int); ok {
						productInfo.ProductID = productID
					}
					
					if productName, ok := info["productName"].(string); ok {
						productInfo.ProductName = productName
					}
					
					if productImage, ok := info["productImage"].(string); ok {
						productInfo.ProductImage = productImage
					}
					
					response.LayoutMetadata.ProductPositionMap[pos] = productInfo
				}
			}
		}
		
		// Extract row product mapping
		if rowProductMapping, ok := state.LayoutMetadata["rowProductMapping"].(map[string]interface{}); ok {
			response.LayoutMetadata.RowProductMapping = rowProductMapping
		}
		
		// Extract location
		if location, ok := state.LayoutMetadata["location"].(string); ok {
			response.LayoutMetadata.Location = location
		}
	}
	
	// Convert HistoricalContext if available
	if state.HistoricalContext != nil {
		response.HistoricalContext = &HistoricalContext{}
		
		if previousVerificationId, ok := state.HistoricalContext["previousVerificationId"].(string); ok {
			response.HistoricalContext.PreviousVerificationID = previousVerificationId
		}
		
		if previousVerificationAt, ok := state.HistoricalContext["previousVerificationAt"].(string); ok {
			response.HistoricalContext.PreviousVerificationAt = previousVerificationAt
		}
		
		if previousVerificationStatus, ok := state.HistoricalContext["previousVerificationStatus"].(string); ok {
			response.HistoricalContext.PreviousVerificationStatus = previousVerificationStatus
		}
		
		if hoursSinceLastVerification, ok := state.HistoricalContext["hoursSinceLastVerification"].(float64); ok {
			response.HistoricalContext.HoursSinceLastVerification = hoursSinceLastVerification
		}
		
		// Extract machine structure
		if machineStructure, ok := state.HistoricalContext["machineStructure"].(map[string]interface{}); ok {
			ms := &MachineStructure{}
			
			if rowCount, ok := machineStructure["rowCount"].(int); ok {
				ms.RowCount = rowCount
			}
			
			if columnsPerRow, ok := machineStructure["columnsPerRow"].(int); ok {
				ms.ColumnsPerRow = columnsPerRow
			}
			
			if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok {
				rowOrder := make([]string, 0, len(rowOrderInterface))
				for _, row := range rowOrderInterface {
					if rowStr, ok := row.(string); ok {
						rowOrder = append(rowOrder, rowStr)
					}
				}
				ms.RowOrder = rowOrder
			}
			
			if columnOrderInterface, ok := machineStructure["columnOrder"].([]interface{}); ok {
				columnOrder := make([]string, 0, len(columnOrderInterface))
				for _, col := range columnOrderInterface {
					if colStr, ok := col.(string); ok {
						columnOrder = append(columnOrder, colStr)
					}
				}
				ms.ColumnOrder = columnOrder
			}
			
			response.HistoricalContext.MachineStructure = ms
		}
		
		// Extract checking status
		if checkingStatus, ok := state.HistoricalContext["checkingStatus"].(map[string]string); ok {
			response.HistoricalContext.CheckingStatus = checkingStatus
		}
		
		// Extract verification summary
		if verificationSummary, ok := state.HistoricalContext["verificationSummary"].(map[string]interface{}); ok {
			vs := &VerificationSummary{}
			
			if totalPositionsChecked, ok := verificationSummary["totalPositionsChecked"].(int); ok {
				vs.TotalPositionsChecked = totalPositionsChecked
			}
			
			if correctPositions, ok := verificationSummary["correctPositions"].(int); ok {
				vs.CorrectPositions = correctPositions
			}
			
			if discrepantPositions, ok := verificationSummary["discrepantPositions"].(int); ok {
				vs.DiscrepantPositions = discrepantPositions
			}
			
			if missingProducts, ok := verificationSummary["missingProducts"].(int); ok {
				vs.MissingProducts = missingProducts
			}
			
			if incorrectProductTypes, ok := verificationSummary["incorrectProductTypes"].(int); ok {
				vs.IncorrectProductTypes = incorrectProductTypes
			}
			
			if unexpectedProducts, ok := verificationSummary["unexpectedProducts"].(int); ok {
				vs.UnexpectedProducts = unexpectedProducts
			}
			
			if emptyPositionsCount, ok := verificationSummary["emptyPositionsCount"].(int); ok {
				vs.EmptyPositionsCount = emptyPositionsCount
			}
			
			if overallAccuracy, ok := verificationSummary["overallAccuracy"].(float64); ok {
				vs.OverallAccuracy = overallAccuracy
			}
			
			if overallConfidence, ok := verificationSummary["overallConfidence"].(float64); ok {
				vs.OverallConfidence = overallConfidence
			}
			
			if verificationStatus, ok := verificationSummary["verificationStatus"].(string); ok {
				vs.VerificationStatus = verificationStatus
			}
			
			if verificationOutcome, ok := verificationSummary["verificationOutcome"].(string); ok {
				vs.VerificationOutcome = verificationOutcome
			}
			
			response.HistoricalContext.VerificationSummary = vs
		}
	}
	
	return response
}
