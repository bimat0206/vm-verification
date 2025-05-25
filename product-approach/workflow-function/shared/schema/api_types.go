// Package schema provides API-related types and structures
package schema

// APIResponse represents a standardized API response
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Timestamp string      `json:"timestamp"`
    RequestId string      `json:"requestId,omitempty"`
}

// APIError represents a standardized API error
type APIError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
    Total      int `json:"total"`
    Limit      int `json:"limit"`
    Offset     int `json:"offset"`
    NextOffset *int `json:"nextOffset,omitempty"`
}

// VerificationListResponse represents the response for listing verifications
type VerificationListResponse struct {
    Results    []VerificationSummary `json:"results"`
    Pagination PaginationInfo        `json:"pagination"`
}

// VerificationSummary represents a summary of verification results
type VerificationSummary struct {
    VerificationId     string  `json:"verificationId"`
    VerificationAt     string  `json:"verificationAt"`
    VendingMachineId   string  `json:"vendingMachineId"`
    Location           string  `json:"location"`
    VerificationStatus string  `json:"verificationStatus"`
    VerificationOutcome string `json:"verificationOutcome"`
    CorrectPositions   int     `json:"correctPositions"`
    DiscrepantPositions int    `json:"discrepantPositions"`
    EmptyPositionsCount int    `json:"emptyPositionsCount"`
    OverallAccuracy    float64 `json:"overallAccuracy"`
    OverallConfidence  float64 `json:"overallConfidence"`
}

// ConversationResponse represents the response for conversation history
type ConversationResponse struct {
    VerificationId string             `json:"verificationId"`
    ConversationAt string             `json:"conversationAt"`
    Status         string             `json:"status"`
    History        []TurnHistory      `json:"history"`
    Metadata       map[string]interface{} `json:"metadata"`
}

// HealthCheckResponse represents the health check response
type HealthCheckResponse struct {
    Status    string                 `json:"status"`
    Version   string                 `json:"version"`
    Timestamp string                 `json:"timestamp"`
    Services  map[string]ServiceHealth `json:"services"`
}

// ServiceHealth represents the health of a service component
type ServiceHealth struct {
    Status    string `json:"status"`
    Latency   int64  `json:"latency,omitempty"`
    Error     string `json:"error,omitempty"`
}

// ImageBrowseRequest represents request for browsing images
type ImageBrowseRequest struct {
    BucketType string `json:"bucketType"` // "reference" or "checking"
    Path       string `json:"path,omitempty"`
    Limit      int    `json:"limit,omitempty"`
}

// ImageBrowseResponse represents response for browsing images
type ImageBrowseResponse struct {
    CurrentPath string       `json:"currentPath"`
    ParentPath  string       `json:"parentPath,omitempty"`
    Items       []BrowseItem `json:"items"`
}

// BrowseItem represents an item in the browse response
type BrowseItem struct {
    Name          string `json:"name"`
    Type          string `json:"type"` // "file" or "folder"
    Path          string `json:"path"`
    DatePartition string `json:"datePartition,omitempty"`
    Size          int64  `json:"size,omitempty"`
    LastModified  string `json:"lastModified,omitempty"`
}