package internal

import (
	"text/template"
)

// Input represents the Lambda input event
type Input struct {
	VerificationContext *VerificationContext `json:"verificationContext"`
	LayoutMetadata      *LayoutMetadata      `json:"layoutMetadata,omitempty"`
	HistoricalContext   *HistoricalContext   `json:"historicalContext,omitempty"`
	Images              *ImageData           `json:"images,omitempty"`
	BedrockConfig       *BedrockConfig       `json:"bedrockConfig,omitempty"`
	TurnNumber          int                  `json:"turnNumber,omitempty"`
	IncludeImage        string               `json:"includeImage,omitempty"`
}

// Response represents the Lambda response
type Response struct {
	VerificationContext *VerificationContext `json:"verificationContext"`
	LayoutMetadata      *LayoutMetadata      `json:"layoutMetadata,omitempty"`
	HistoricalContext   *HistoricalContext   `json:"historicalContext,omitempty"`
	SystemPrompt        SystemPrompt         `json:"systemPrompt"`
	BedrockConfig       BedrockConfig        `json:"bedrockConfig"`
}

// VerificationContext contains verification metadata
type VerificationContext struct {
	VerificationID     string `json:"verificationId"`
	VerificationAt     string `json:"verificationAt"`
	Status             string `json:"status"`
	VerificationType   string `json:"verificationType"`
	VendingMachineID   string `json:"vendingMachineId,omitempty"`
	LayoutID           int    `json:"layoutId,omitempty"`
	LayoutPrefix       string `json:"layoutPrefix,omitempty"`
	ReferenceImageURL  string `json:"referenceImageUrl,omitempty"`
	CheckingImageURL   string `json:"checkingImageUrl,omitempty"`
	NotificationEnabled bool   `json:"notificationEnabled,omitempty"`
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
	PreviousVerificationID     string            `json:"previousVerificationId"`
	PreviousVerificationAt     string            `json:"previousVerificationAt"`
	PreviousVerificationStatus string            `json:"previousVerificationStatus"`
	HoursSinceLastVerification float64           `json:"hoursSinceLastVerification"`
	MachineStructure           *MachineStructure `json:"machineStructure,omitempty"`
	CheckingStatus             map[string]string `json:"checkingStatus,omitempty"`
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
	ReferenceImageBase64 string `json:"referenceImageBase64,omitempty"`
	CheckingImageBase64  string `json:"checkingImageBase64,omitempty"`
}

// SystemPrompt represents the generated system prompt
type SystemPrompt struct {
	Content       string `json:"content"`
	PromptID      string `json:"promptId"`
	CreatedAt     string `json:"createdAt"`
	PromptVersion string `json:"promptVersion"`
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
	baseDir    string
	templates  map[string]*template.Template
	versions   map[string]string
}