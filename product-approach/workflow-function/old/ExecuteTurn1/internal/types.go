// Package internal provides shared types across ExecuteTurn1 functions
package internal

import (
	"fmt"
	"strings"
	"strconv"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// S3 State Management Category Structure - matching design specifications
const (
	// Categories
	CategoryProcessing = "processing"
	CategoryImages     = "images"
	CategoryPrompts    = "prompts" 
	CategoryResponses  = "responses"
)

// Standard filenames for each category - matching design specifications
// In state/saver.go, update the file name constants
const (
    // Processing category
    FileInitialization   = "initialization.json"
    FileLayoutMetadata   = "layout-metadata.json"
    FileHistoricalContext = "historical-context.json"
    
    // Images category
    FileImageMetadata    = "metadata.json"
    FileReferenceBase64  = "reference-base64.base64"
    FileCheckingBase64   = "checking-base64.base64"
    
    // Prompts category
    FileSystemPrompt     = "system-prompt.json"
    
    // Responses category
    FileTurn1Response    = "turn1-raw-response.json" // Updated filename
)

// S3 path date components
const (
	DatePathFormat = "%04d/%02d/%02d/%s" // YYYY/MM/DD/verificationId
)

// Storage structure information
const (
	Base64Extension = ".base64"
	JsonExtension = ".json"
)

// StateReferences contains all S3 object references needed for the workflow
// Updated to match the correct folder structure
type StateReferences struct {
	VerificationId      string              `json:"verificationId"`
	DatePartition       string              `json:"datePartition"` // YYYY/MM/DD format
	
	// Processing category references
	Initialization      *s3state.Reference  `json:"initialization"`
	LayoutMetadata      *s3state.Reference  `json:"layoutMetadata,omitempty"`
	HistoricalContext   *s3state.Reference  `json:"historicalContext,omitempty"`
	ConversationState   *s3state.Reference  `json:"conversationState,omitempty"`
	
	// Images category references
	ImageMetadata       *s3state.Reference  `json:"imageMetadata"`
	ReferenceBase64     *s3state.Reference  `json:"referenceBase64,omitempty"`
	CheckingBase64      *s3state.Reference  `json:"checkingBase64,omitempty"`
	
	// Prompts category references
	SystemPrompt        *s3state.Reference  `json:"systemPrompt"`
	Turn1Prompt         *s3state.Reference  `json:"turn1Prompt,omitempty"`
	
	// Responses category references
	Turn1Response       *s3state.Reference  `json:"turn1Response,omitempty"`
	Turn2Response       *s3state.Reference  `json:"turn2Response,omitempty"`
	Turn2Thinking       *s3state.Reference  `json:"turn2Thinking,omitempty"`
	FinalResults        *s3state.Reference  `json:"finalResults,omitempty"`
	StorageResult       *s3state.Reference  `json:"storageResult,omitempty"`
	NotificationResult  *s3state.Reference  `json:"notificationResult,omitempty"`
}

// HybridStorageConfig defines configuration for hybrid storage of Base64 data
type HybridStorageConfig struct {
	EnableHybridStorage     bool   `json:"enableHybridStorage"`
	TempBase64Bucket        string `json:"tempBase64Bucket"`
	Base64SizeThreshold     int64  `json:"base64SizeThreshold"`
	Base64RetrievalTimeout  int    `json:"base64RetrievalTimeout"`
	MaxInlineBase64Size     int64  `json:"maxInlineBase64Size"`
	Base64RetrievalRetry    int    `json:"base64RetrievalRetry"`
	Base64RetrievalBackoff  int    `json:"base64RetrievalBackoff"`
}

// StepFunctionInput represents the input to the step function
// Updated to capture raw JSON structure for S3References
type StepFunctionInput struct {
	S3References    map[string]map[string]interface{} `json:"s3References"` // Changed to capture raw JSON
	VerificationId  string                            `json:"verificationId"`
	Status          string                            `json:"status"`
	Config          map[string]interface{}            `json:"config,omitempty"`
}

// StepFunctionOutput represents the output from the step function
type StepFunctionOutput struct {
	S3References    *StateReferences       `json:"s3References"`
	Status          string                 `json:"status"`
	Summary         map[string]interface{} `json:"summary,omitempty"`
	Error           *schema.ErrorInfo      `json:"error,omitempty"`
}

// MapS3References maps dynamic S3 references to structured StateReferences
func (input *StepFunctionInput) MapS3References() *StateReferences {
    // If no S3References map, can't proceed with mapping
    if input.S3References == nil || len(input.S3References) == 0 {
        return nil
    }
    
    // Create new StateReferences
    refs := &StateReferences{
        VerificationId: input.VerificationId,
    }
    
    // Extract date partition if present in any reference
    for _, refMap := range input.S3References {
        if refKey, ok := refMap["key"].(string); ok {
            parts := strings.Split(refKey, "/")
            if len(parts) >= 4 {
                // Check if first three parts form a date pattern (YYYY/MM/DD)
                if isYearMonthDay(parts[0], parts[1], parts[2]) {
                    refs.DatePartition = fmt.Sprintf("%s/%s/%s", parts[0], parts[1], parts[2])
                    break
                }
            }
        }
    }
    
    // Convert each raw map to an s3state.Reference and map to specific fields
    for key, refMap := range input.S3References {
        ref := convertMapToReference(refMap)
        if ref != nil {
            mapReferenceToField(refs, key, ref)
        }
    }
    
    return refs
}

// Helper to check if parts form a date pattern
func isYearMonthDay(year, month, day string) bool {
    // Simple check if these could be a date
    if len(year) != 4 || len(month) != 2 || len(day) != 2 {
        return false
    }
    
    // Check if all are numeric
    _, yearErr := strconv.Atoi(year)
    _, monthErr := strconv.Atoi(month)
    _, dayErr := strconv.Atoi(day)
    
    return yearErr == nil && monthErr == nil && dayErr == nil
}

// Improved conversion from map to reference with better type handling
func convertMapToReference(refMap map[string]interface{}) *s3state.Reference {
    ref := &s3state.Reference{}
    
    // Extract bucket
    if bucket, ok := refMap["bucket"].(string); ok {
        ref.Bucket = bucket
    } else {
        return nil // Bucket is required
    }
    
    // Extract key
    if key, ok := refMap["key"].(string); ok {
        ref.Key = key
    } else {
        return nil // Key is required
    }
    
    // Extract size with flexible type handling
    if size, ok := refMap["size"].(float64); ok {
        ref.Size = int64(size)
    } else if size, ok := refMap["size"].(int64); ok {
        ref.Size = size
    } else if size, ok := refMap["size"].(int); ok {
        ref.Size = int64(size)
    } else if sizeStr, ok := refMap["size"].(string); ok {
        if sizeVal, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
            ref.Size = sizeVal
        }
    }
    
    return ref
}


// mapReferenceToField maps a reference to the appropriate field in StateReferences
func mapReferenceToField(refs *StateReferences, key string, ref *s3state.Reference) {
	switch key {
	case "processing_initialization":
		refs.Initialization = ref
	case "images_metadata":
		refs.ImageMetadata = ref
	case "processing_layout-metadata":
		refs.LayoutMetadata = ref
	case "processing_historical-context":
		refs.HistoricalContext = ref
	case "processing_conversation-state":
		refs.ConversationState = ref
	case "images_reference-base64":
		refs.ReferenceBase64 = ref
	case "images_checking-base64":
		refs.CheckingBase64 = ref
	case "prompts_system":
		refs.SystemPrompt = ref
	case "prompts_turn1-prompt":
		refs.Turn1Prompt = ref
	case "responses_turn1-response":
		refs.Turn1Response = ref
	case "responses_turn2-response":
		refs.Turn2Response = ref
	case "responses_turn2-thinking":
		refs.Turn2Thinking = ref
	case "processing_final-results":
		refs.FinalResults = ref
	case "processing_storage-result":
		refs.StorageResult = ref
	case "processing_notification-result":
		refs.NotificationResult = ref
	}
}