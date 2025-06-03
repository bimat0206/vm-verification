// Package schema provides system prompt related types and functions
package schema

// CompleteSystemPrompt represents the complete system prompt structure
// that matches the expected JSON schema
type CompleteSystemPrompt struct {
	VerificationId     string                 `json:"verificationId"`
	VerificationType   string                 `json:"verificationType"`
	PromptContent      *PromptContent         `json:"promptContent"`
	BedrockConfiguration *BedrockConfiguration `json:"bedrockConfiguration"`
	ContextInformation *ContextInformation    `json:"contextInformation,omitempty"`
	OutputSpecification *OutputSpecification  `json:"outputSpecification"`
	ProcessingMetadata *ProcessingMetadata    `json:"processingMetadata"`
	Version            string                 `json:"version"`
}

// PromptContent contains the actual prompt content
type PromptContent struct {
	SystemMessage   string `json:"systemMessage"`
	TemplateVersion string `json:"templateVersion,omitempty"`
	PromptType      string `json:"promptType"`
}

// BedrockConfiguration contains Bedrock-specific configuration
type BedrockConfiguration struct {
	AnthropicVersion string    `json:"anthropicVersion"`
	MaxTokens        int       `json:"maxTokens"`
	Thinking         *Thinking `json:"thinking,omitempty"`
	ModelId          string    `json:"modelId"`
}

// ContextInformation contains context information for the prompt
type ContextInformation struct {
	MachineStructure  *MachineStructure  `json:"machineStructure,omitempty"`
	LayoutInformation *LayoutInformation `json:"layoutInformation,omitempty"`
	HistoricalContext *HistoricalContext `json:"historicalContext,omitempty"`
}

// We're using the existing MachineStructure type from types.go

// LayoutInformation contains information about the layout
type LayoutInformation struct {
	LayoutId     int    `json:"layoutId"`
	LayoutPrefix string `json:"layoutPrefix"`
	ProductCount int    `json:"productCount"`
}

// HistoricalContext contains information about previous verifications
type HistoricalContext struct {
	PreviousVerificationId     string  `json:"previousVerificationId"`
	HoursSinceLastVerification float64 `json:"hoursSinceLastVerification"`
	KnownIssuesCount           int     `json:"knownIssuesCount"`
}

// OutputSpecification contains information about the expected output format
type OutputSpecification struct {
	ExpectedFormat           string `json:"expectedFormat"`
	RequiresMandatoryStructure bool   `json:"requiresMandatoryStructure"`
	ConversationTurns        int    `json:"conversationTurns"`
}

// ProcessingMetadata contains metadata about the prompt generation process
type ProcessingMetadata struct {
	CreatedAt          string   `json:"createdAt"`
	GenerationTimeMs   int64    `json:"generationTimeMs"`
	TemplateSource     string   `json:"templateSource,omitempty"`
	ContextEnrichment  []string `json:"contextEnrichment,omitempty"`
}

// ConvertToCompleteSystemPrompt converts a SystemPrompt to a CompleteSystemPrompt
func ConvertToCompleteSystemPrompt(prompt *SystemPrompt, verificationContext *VerificationContext) *CompleteSystemPrompt {
	// Determine prompt type based on verification type
	promptType := "LAYOUT_VERIFICATION"
	if verificationContext.VerificationType == VerificationTypePreviousVsCurrent {
		promptType = "TEMPORAL_COMPARISON"
	}

	// Create the complete system prompt
	completePrompt := &CompleteSystemPrompt{
		VerificationId:   verificationContext.VerificationId,
		VerificationType: verificationContext.VerificationType,
		PromptContent: &PromptContent{
			SystemMessage:   prompt.Content,
			TemplateVersion: prompt.PromptVersion,
			PromptType:      promptType,
		},
		BedrockConfiguration: &BedrockConfiguration{
			AnthropicVersion: prompt.BedrockConfig.AnthropicVersion,
			MaxTokens:        prompt.BedrockConfig.MaxTokens,
			Thinking: &Thinking{
				Type:         prompt.BedrockConfig.Thinking.Type,
				BudgetTokens: prompt.BedrockConfig.Thinking.BudgetTokens,
			},
			ModelId: "anthropic.claude-3-7-sonnet-20250219-v1:0", // Default - use ConvertToCompleteSystemPromptWithConfig for configurable model
		},
		OutputSpecification: &OutputSpecification{
			ExpectedFormat:           "STRUCTURED_TEXT",
			RequiresMandatoryStructure: true,
			ConversationTurns:        2,
		},
		ProcessingMetadata: &ProcessingMetadata{
			CreatedAt:        FormatISO8601(),
			GenerationTimeMs: 0, // This will be set by the caller
			TemplateSource:   "DYNAMIC_GENERATION",
			ContextEnrichment: []string{
				"MACHINE_STRUCTURE_INJECTION",
			},
		},
		Version: "1.0",
	}

	// Create machine structure
	machineStructure := &MachineStructure{
		RowCount:      6,
		ColumnsPerRow: 10,
		ColumnOrder:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		RowOrder:      []string{"A", "B", "C", "D", "E", "F"},
	}

	// Create context information based on verification type
	contextInfo := &ContextInformation{
		MachineStructure: machineStructure,
	}

	// Add layout information for LAYOUT_VS_CHECKING
	if verificationContext.VerificationType == VerificationTypeLayoutVsChecking {
		contextInfo.LayoutInformation = &LayoutInformation{
			LayoutId:     verificationContext.LayoutId,
			LayoutPrefix: verificationContext.LayoutPrefix,
			ProductCount: 3, // Default value
		}
	} else {
		// Add historical context for PREVIOUS_VS_CURRENT
		contextInfo.HistoricalContext = &HistoricalContext{
			PreviousVerificationId:     verificationContext.PreviousVerificationId,
			HoursSinceLastVerification: 24, // Default value
			KnownIssuesCount:           0,  // Default value
		}
	}

	completePrompt.ContextInformation = contextInfo

	return completePrompt
}

// ConvertToCompleteSystemPromptWithConfig converts a SystemPrompt to a CompleteSystemPrompt with configurable model ID
func ConvertToCompleteSystemPromptWithConfig(prompt *SystemPrompt, verificationContext *VerificationContext, modelId string) *CompleteSystemPrompt {
	// Determine prompt type based on verification type
	promptType := "LAYOUT_VERIFICATION"
	if verificationContext.VerificationType == VerificationTypePreviousVsCurrent {
		promptType = "TEMPORAL_COMPARISON"
	}

	// Create the complete system prompt
	completePrompt := &CompleteSystemPrompt{
		VerificationId:   verificationContext.VerificationId,
		VerificationType: verificationContext.VerificationType,
		PromptContent: &PromptContent{
			SystemMessage:   prompt.Content,
			TemplateVersion: prompt.PromptVersion,
			PromptType:      promptType,
		},
		BedrockConfiguration: &BedrockConfiguration{
			AnthropicVersion: prompt.BedrockConfig.AnthropicVersion,
			MaxTokens:        prompt.BedrockConfig.MaxTokens,
			Thinking: &Thinking{
				Type:         prompt.BedrockConfig.Thinking.Type,
				BudgetTokens: prompt.BedrockConfig.Thinking.BudgetTokens,
			},
			ModelId: modelId,
		},
		OutputSpecification: &OutputSpecification{
			ExpectedFormat:           "STRUCTURED_TEXT",
			RequiresMandatoryStructure: true,
			ConversationTurns:        2,
		},
		ProcessingMetadata: &ProcessingMetadata{
			CreatedAt:        FormatISO8601(),
			GenerationTimeMs: 0, // This will be set by the caller
			TemplateSource:   "DYNAMIC_GENERATION",
			ContextEnrichment: []string{
				"MACHINE_STRUCTURE_INJECTION",
			},
		},
		Version: "1.0",
	}

	// Create machine structure
	machineStructure := &MachineStructure{
		RowCount:      6,
		ColumnsPerRow: 10,
		ColumnOrder:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
		RowOrder:      []string{"A", "B", "C", "D", "E", "F"},
	}

	// Create context information based on verification type
	contextInfo := &ContextInformation{
		MachineStructure: machineStructure,
	}

	// Add layout information for LAYOUT_VS_CHECKING
	if verificationContext.VerificationType == VerificationTypeLayoutVsChecking {
		contextInfo.LayoutInformation = &LayoutInformation{
			LayoutId:     verificationContext.LayoutId,
			LayoutPrefix: verificationContext.LayoutPrefix,
			ProductCount: 3, // Default value
		}
	} else {
		// Add historical context for PREVIOUS_VS_CURRENT
		contextInfo.HistoricalContext = &HistoricalContext{
			PreviousVerificationId:     verificationContext.PreviousVerificationId,
			HoursSinceLastVerification: 24, // Default value
			KnownIssuesCount:           0,  // Default value
		}
	}

	completePrompt.ContextInformation = contextInfo

	return completePrompt
}

// ADD: Enhanced system prompt with design-specific fields
type EnhancedSystemPrompt struct {
    *CompleteSystemPrompt
    
    // Add fields from design document
    RowProtocol         []RowProtocolEntry `json:"rowProtocol,omitempty"`
    ProductCatalog      []ProductInfo      `json:"productCatalog,omitempty"`
    ValidationFramework *ValidationFramework `json:"validationFramework,omitempty"`
    TokenOptimization   *TokenConfig       `json:"tokenOptimization,omitempty"`
}

type RowProtocolEntry struct {
    RowCode     string `json:"rowCode"`
    Description string `json:"description"`
}

type ProductInfo struct {
    ProductId   int    `json:"productId"`
    ProductName string `json:"productName"`
    Description string `json:"description"`
}

type ValidationFramework struct {
    CriticalRequirements []string `json:"criticalRequirements"`
    OutputFormat         string   `json:"outputFormat"`
    QualityChecks        []string `json:"qualityChecks"`
}

type TokenConfig struct {
    MaxInputTokens  int `json:"maxInputTokens"`
    MaxOutputTokens int `json:"maxOutputTokens"`
}
