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

// TokenUsage represents token usage metrics
type TokenUsage struct {
	InputTokens   int `json:"inputTokens"`
	OutputTokens  int `json:"outputTokens"`
	ThinkingTokens int `json:"thinkingTokens,omitempty"`
	TotalTokens   int `json:"totalTokens"`
}
