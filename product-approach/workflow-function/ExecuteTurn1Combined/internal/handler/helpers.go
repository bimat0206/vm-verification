// internal/handler/helpers.go
package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"workflow-function/ExecuteTurn1Combined/internal/models"
)

// S3ReferenceTree represents a complete tree of S3 references for a verification
type S3ReferenceTree struct {
	Initialization models.S3Reference   `json:"initialization,omitempty"`
	Images         ImageReferences      `json:"images,omitempty"`
	Processing     ProcessingReferences `json:"processing,omitempty"`
	Prompts        PromptReferences     `json:"prompts,omitempty"`
	Responses      ResponseReferences   `json:"responses,omitempty"`
}

// PromptReferences holds S3 references for prompt artifacts
type PromptReferences struct {
	SystemPrompt models.S3Reference `json:"systemPrompt,omitempty"`
	Turn1Prompt  models.S3Reference `json:"turn1Prompt,omitempty"`
	Turn2Prompt  models.S3Reference `json:"turn2Prompt,omitempty"`
}

// ImageReferences holds S3 references for image artifacts
type ImageReferences struct {
	Metadata  models.S3Reference `json:"metadata,omitempty"`
	Reference models.S3Reference `json:"reference,omitempty"`
	Checking  models.S3Reference `json:"checking,omitempty"`
}

// ProcessingReferences holds S3 references for processing artifacts
type ProcessingReferences struct {
	HistoricalContext models.S3Reference `json:"historicalContext,omitempty"`
	LayoutMetadata    models.S3Reference `json:"layoutMetadata,omitempty"`
}

// ResponseReferences holds S3 references for response artifacts
type ResponseReferences struct {
	Turn1Raw       models.S3Reference `json:"turn1Raw,omitempty"`
	Turn1Processed models.S3Reference `json:"turn1Processed,omitempty"`
	Turn2Raw       models.S3Reference `json:"turn2Raw,omitempty"`
	Turn2Processed models.S3Reference `json:"turn2Processed,omitempty"`
}

// ExecutionSummary contains metrics and identifiers for execution
type ExecutionSummary struct {
	AnalysisStage       string             `json:"analysisStage"`
	VerificationType    string             `json:"verificationType"`
	ProcessingTimeMs    int64              `json:"processingTimeMs"`
	TokenUsage          TokenUsageDetailed `json:"tokenUsage"`
	BedrockLatencyMs    int64              `json:"bedrockLatencyMs"`
	BedrockRequestId    string             `json:"bedrockRequestId"`
	DynamodbUpdated     bool               `json:"dynamodbUpdated"`
	ConversationTracked bool               `json:"conversationTracked"`
	S3StorageCompleted  bool               `json:"s3StorageCompleted"`
}

// TokenUsageDetailed provides a detailed breakdown of token usage
type TokenUsageDetailed struct {
	Input    int `json:"input"`
	Output   int `json:"output"`
	Thinking int `json:"thinking"`
	Total    int `json:"total"`
}

// buildS3RefTree constructs a unified S3 reference tree from various sources
func buildS3RefTree(
	base models.Turn1RequestS3Refs,
	promptRef, rawRef, procRef models.S3Reference,
) S3ReferenceTree {
	// Extract verification ID from the key pattern
	verificationID := extractVerificationID(rawRef.Key)
	datePartition := extractDatePartitionFromKey(rawRef.Key)

	prefix := func(key string) string {
		if datePartition != "" {
			return fmt.Sprintf("%s/%s", datePartition, key)
		}
		return key
	}

	// Create initialization reference
	initRefKey := prefix(fmt.Sprintf("%s/initialization.json", verificationID))
	initRef := models.S3Reference{
		Bucket: rawRef.Bucket,
		Key:    initRefKey,
	}

	// Create images metadata reference
	imagesMetadataKey := prefix(fmt.Sprintf("%s/images/metadata.json", verificationID))
	imagesMetadataRef := models.S3Reference{
		Bucket: rawRef.Bucket,
		Key:    imagesMetadataKey,
	}

	// Create layout metadata reference for LAYOUT_VS_CHECKING
	layoutMetadataKey := prefix(fmt.Sprintf("%s/processing/layout-metadata.json", verificationID))
	layoutMetadataRef := models.S3Reference{
		Bucket: rawRef.Bucket,
		Key:    layoutMetadataKey,
	}

	// Create historical context reference for PREVIOUS_VS_CURRENT
	historicalContextKey := prefix(fmt.Sprintf("%s/processing/historical-context.json", verificationID))
	historicalContextRef := models.S3Reference{
		Bucket: rawRef.Bucket,
		Key:    historicalContextKey,
	}

	tree := S3ReferenceTree{
		Initialization: initRef,
		Images: ImageReferences{
			Metadata:  imagesMetadataRef,
			Reference: base.Images.ReferenceBase64,
		},
		Processing: ProcessingReferences{
			HistoricalContext: historicalContextRef,
			LayoutMetadata:    layoutMetadataRef,
		},
		Prompts: PromptReferences{
			SystemPrompt: base.Prompts.System,
		},
		Responses: ResponseReferences{
			Turn1Raw:       rawRef,
			Turn1Processed: procRef,
		},
	}

	// Only include promptRef if the key is not empty
	if promptRef.Key != "" {
		tree.Prompts.Turn1Prompt = promptRef
	}

	return tree
}

// extractVerificationID extracts the verification ID from an S3 key
func extractVerificationID(key string) string {
	// Expected format: verif-YYYYMMDDHHMMSS/responses/turn1-raw-response.json
	// or: YYYY/MM/DD/verif-YYYYMMDDHHMMSS/responses/turn1-raw-response.json

	parts := strings.Split(key, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "verif-") {
			return part
		}
	}

	// If we can't find a verification ID, return a placeholder
	return "unknown-verification-id"
}

// extractDatePartitionFromKey extracts the YYYY/MM/DD prefix from an S3 key if present.
func extractDatePartitionFromKey(key string) string {
	parts := strings.Split(key, "/")
	if len(parts) >= 4 {
		year, month, day := parts[0], parts[1], parts[2]
		if len(year) == 4 && len(month) == 2 && len(day) == 2 &&
			isAllDigits(year) && isAllDigits(month) && isAllDigits(day) {
			return fmt.Sprintf("%s/%s/%s", year, month, day)
		}
	}
	return ""
}

// isAllDigits returns true if all characters in the string are digits
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// buildSummary creates a summary of the turn execution
func buildSummary(
	totalDurationMs int64,
	invoke *models.BedrockResponse,
	verificationType string,
	bedrockLatencyMs int64,
	dynamoOK bool,
) ExecutionSummary {
	// Convert TokenUsage to TokenUsageDetailed
	tokenUsage := TokenUsageDetailed{
		Input:    invoke.TokenUsage.InputTokens,
		Output:   invoke.TokenUsage.OutputTokens,
		Thinking: invoke.TokenUsage.ThinkingTokens,
		Total:    invoke.TokenUsage.TotalTokens,
	}

	// Default to true for conversation tracked and S3 storage completed
	// These could be parameterized in the future if needed
	conversationTracked := true
	s3StorageCompleted := true

	// Default to true for dynamoOK if not provided
	dynamodbUpdated := dynamoOK

	// Use the provided bedrock latency

	return ExecutionSummary{
		AnalysisStage:       "REFERENCE_ANALYSIS",
		VerificationType:    verificationType,
		ProcessingTimeMs:    totalDurationMs,
		TokenUsage:          tokenUsage,
		BedrockLatencyMs:    bedrockLatencyMs,
		BedrockRequestId:    invoke.RequestID,
		DynamodbUpdated:     dynamodbUpdated,
		ConversationTracked: conversationTracked,
		S3StorageCompleted:  s3StorageCompleted,
	}
}

// extractCheckingImageUrl extracts the checking image URL from the S3 key
func extractCheckingImageUrl(s3Key string) string {
	if s3Key == "" {
		return ""
	}

	// In a real implementation, this would parse the S3 key to extract the URL
	// For now, we'll just return the key as a placeholder
	return s3Key
}

// calculateHoursSince calculates hours between a timestamp and now
func calculateHoursSince(timestamp string) float64 {
	if timestamp == "" {
		return 0
	}

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return 0
	}

	duration := time.Since(t)
	return duration.Hours()
}

// Helper function to extract image reference URL from request
func (h *Handler) ImageRef(req *models.Turn1Request) string {
	if req != nil && req.S3Refs.Images.ReferenceBase64.Key != "" {
		return req.S3Refs.Images.ReferenceBase64.Key
	}
	return ""
}
