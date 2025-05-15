package internal

import (
	"encoding/json"
	"text/template"
	
	"workflow-function/shared/schema"
)

// Input adapts the Lambda input event to our internal structures
type Input struct {
	State              *schema.WorkflowState  `json:"-"`
	VerificationContext *schema.VerificationContext `json:"verificationContext"`
	LayoutMetadata      map[string]interface{} `json:"layoutMetadata,omitempty"`
	HistoricalContext   map[string]interface{} `json:"historicalContext,omitempty"`
	Images              *schema.ImageData     `json:"images,omitempty"`
	TurnNumber          int                   `json:"turnNumber,omitempty"`
	IncludeImage        string                `json:"includeImage,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Input
func (i *Input) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary struct to get the verification context
	var temp struct {
		VerificationContext *schema.VerificationContext `json:"verificationContext"`
		LayoutMetadata      json.RawMessage            `json:"layoutMetadata,omitempty"`
		HistoricalContext   json.RawMessage            `json:"historicalContext,omitempty"`
		Images              *schema.ImageData          `json:"images,omitempty"`
		TurnNumber          int                        `json:"turnNumber,omitempty"`
		IncludeImage        string                     `json:"includeImage,omitempty"`
	}
	
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	
	// Create a new workflow state
	state := &schema.WorkflowState{
		SchemaVersion:      "1.0.0",
		VerificationContext: temp.VerificationContext,
		Images:              temp.Images,
	}
	
	// Parse layout metadata if present
	if len(temp.LayoutMetadata) > 0 {
		var layoutMetadata map[string]interface{}
		if err := json.Unmarshal(temp.LayoutMetadata, &layoutMetadata); err != nil {
			return err
		}
		state.LayoutMetadata = layoutMetadata
		i.LayoutMetadata = layoutMetadata
	}
	
	// Parse historical context if present
	if len(temp.HistoricalContext) > 0 {
		var historicalContext map[string]interface{}
		if err := json.Unmarshal(temp.HistoricalContext, &historicalContext); err != nil {
			return err
		}
		state.HistoricalContext = historicalContext
		i.HistoricalContext = historicalContext
	}
	
	// Set the state and other fields
	i.State = state
	i.VerificationContext = temp.VerificationContext
	i.Images = temp.Images
	i.TurnNumber = temp.TurnNumber
	i.IncludeImage = temp.IncludeImage
	
	return nil
}

// Response represents the Lambda response
type Response struct {
	State              *schema.WorkflowState   `json:"-"`
	VerificationContext *schema.VerificationContext `json:"verificationContext"`
	LayoutMetadata      map[string]interface{} `json:"layoutMetadata,omitempty"`
	HistoricalContext   map[string]interface{} `json:"historicalContext,omitempty"`
	SystemPrompt        *SystemPromptContent   `json:"systemPrompt"`
	BedrockConfig       *schema.BedrockConfig  `json:"bedrockConfig"`
}

// SystemPromptContent represents the content of the system prompt without BedrockConfig
type SystemPromptContent struct {
	Content       string `json:"content"`
	PromptId      string `json:"promptId,omitempty"`
	PromptVersion string `json:"promptVersion,omitempty"`
}

// ToJSON converts the response to JSON
func (r *Response) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// FromWorkflowState creates a Response from a WorkflowState
func ResponseFromWorkflowState(state *schema.WorkflowState) *Response {
	resp := &Response{
		State:              state,
		VerificationContext: state.VerificationContext,
		SystemPrompt:        &SystemPromptContent{
			Content:       state.SystemPrompt.Content,
			PromptId:      state.SystemPrompt.PromptId,
			PromptVersion: state.SystemPrompt.PromptVersion,
		},
		BedrockConfig:       state.SystemPrompt.BedrockConfig,
	}
	
	if state.LayoutMetadata != nil {
		resp.LayoutMetadata = state.LayoutMetadata
	}
	
	if state.HistoricalContext != nil {
		resp.HistoricalContext = state.HistoricalContext
	}
	
	return resp
}

// MachineStructure describes the vending machine physical layout
type MachineStructure struct {
	RowCount      int      `json:"rowCount"`
	ColumnsPerRow int      `json:"columnsPerRow"`
	RowOrder      []string `json:"rowOrder"`
	ColumnOrder   []string `json:"columnOrder"`
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

// ProductInfo contains product details for a specific position
type ProductInfo struct {
	ProductID    int    `json:"productId"`
	ProductName  string `json:"productName"`
	ProductImage string `json:"productImage,omitempty"`
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