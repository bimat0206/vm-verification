// Package schema provides standardized types for workflow functions
package schema

import (
	"time"
)

// MachineStructure represents the vending machine structure
type MachineStructure struct {
	RowCount      int      `json:"rowCount"`
	ColumnsPerRow int      `json:"columnsPerRow"`
	RowOrder      []string `json:"rowOrder"`
	ColumnOrder   []string `json:"columnOrder"`
}

// ImageMetadata represents image metadata from S3
type ImageMetadata struct {
	ContentType   string    `json:"contentType"`
	Size          int64     `json:"size"`
	LastModified  time.Time `json:"lastModified"`
	ETag          string    `json:"etag"`
	BucketOwner   string    `json:"bucketOwner"`
}

// Images represents the image metadata structure
type Images struct {
	ReferenceImageMeta ImageMetadata `json:"referenceImageMeta"`
	CheckingImageMeta  ImageMetadata `json:"checkingImageMeta"`
	BucketOwner        string        `json:"bucketOwner,omitempty"`
}

// TurnHistory represents a single turn in conversation history
type TurnHistory struct {
	TurnID        int        `json:"turnId"`
	Timestamp     time.Time  `json:"timestamp"`
	Prompt        string     `json:"prompt"`
	Response      string     `json:"response"`
	LatencyMs     int64      `json:"latencyMs"`
	TokenUsage    TokenUsage `json:"tokenUsage"`
	AnalysisStage string     `json:"analysisStage"`
}

// TokenUsage is defined in bedrock.go
// Add these structures for template support

// PromptTemplate represents a prompt template
type PromptTemplate struct {
    TemplateId      string                 `json:"templateId"`
    TemplateVersion string                 `json:"templateVersion"`
    TemplateType    string                 `json:"templateType"` // "turn1-layout-vs-checking", etc.
    Content         string                 `json:"content"`
    Variables       map[string]interface{} `json:"variables,omitempty"`
    CreatedAt       string                 `json:"createdAt"`
    UpdatedAt       string                 `json:"updatedAt"`
}

// TemplateProcessor handles template processing
type TemplateProcessor struct {
    Template        *PromptTemplate        `json:"template"`
    ContextData     map[string]interface{} `json:"contextData"`
    ProcessedPrompt string                 `json:"processedPrompt"`
    ProcessingTime  int64                  `json:"processingTimeMs"`
    TokenEstimate   int                    `json:"tokenEstimate"`
}

// ConversationTracker tracks conversation progress
type ConversationTracker struct {
    ConversationId   string                 `json:"conversationId"`
    CurrentTurn      int                    `json:"currentTurn"`
    MaxTurns         int                    `json:"maxTurns"`
    TurnStatus       string                 `json:"turnStatus"`
    ConversationAt   string                 `json:"conversationAt"`
    History          []interface{}          `json:"history"`
    Metadata         map[string]interface{} `json:"metadata,omitempty"`
}