// internal/models/bedrock.go
package models

import (
	"encoding/json"
	"workflow-function/shared/schema"
)

// TokenUsage captures the Bedrock usage statistics returned by Claude.
// Using standardized schema type
type TokenUsage = schema.TokenUsage

// BedrockResponse is the in-memory representation of a Claude Converse result.
// Enhanced with schema integration
type BedrockResponse struct {
	Raw        json.RawMessage        `json:"raw"`        // exact JSON payload from Bedrock
	Processed  interface{}            `json:"processed"`  // parsed / summarised analysis object
	TokenUsage schema.TokenUsage      `json:"tokenUsage"` // Use schema type
	RequestID  string                 `json:"requestId"`  // X-Amzn-RequestId header from Bedrock
	Metadata   map[string]interface{} `json:"metadata,omitempty"` // Additional metadata including thinking content
}
