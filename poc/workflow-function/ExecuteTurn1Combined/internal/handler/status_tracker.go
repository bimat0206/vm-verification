package handler

import (
	"context"
	"time"
	"workflow-function/shared/schema"
)

// StatusTracker manages status updates and history tracking
type StatusTracker struct {
	statusHistory []schema.StatusHistoryEntry
	startTime     time.Time
}

// NewStatusTracker creates a new instance of StatusTracker
func NewStatusTracker(startTime time.Time) *StatusTracker {
	return &StatusTracker{
		statusHistory: make([]schema.StatusHistoryEntry, 0),
		startTime:     startTime,
	}
}

// UpdateStatusWithHistory updates status and maintains history
func (s *StatusTracker) UpdateStatusWithHistory(ctx context.Context, verificationID, status, stage string, metadata map[string]interface{}) error {
	statusEntry := schema.StatusHistoryEntry{
		Status:           status,
		Timestamp:        schema.FormatISO8601(),
		FunctionName:     "ExecuteTurn1Combined",
		ProcessingTimeMs: time.Since(s.startTime).Milliseconds(),
		Stage:            stage,
		Metrics:          metadata,
	}

	s.statusHistory = append(s.statusHistory, statusEntry)

	// In a full implementation, this would also update DynamoDB
	// For now, we just track locally
	return nil
}

// GetHistory returns all status history entries
func (s *StatusTracker) GetHistory() []schema.StatusHistoryEntry {
	return s.statusHistory
}

// GetHistoryCount returns the number of status updates
func (s *StatusTracker) GetHistoryCount() int {
	return len(s.statusHistory)
}
