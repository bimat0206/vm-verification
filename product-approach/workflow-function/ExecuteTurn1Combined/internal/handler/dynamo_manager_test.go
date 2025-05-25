package handler

import (
	"context"
	"errors"
	"testing"

	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

type mockDynamo struct {
	statusErr error
	turnErr   error
}

func (m *mockDynamo) UpdateVerificationStatus(ctx context.Context, verificationID string, status models.VerificationStatus, metrics models.TokenUsage) error {
	return nil
}
func (m *mockDynamo) RecordConversationTurn(ctx context.Context, turn *models.ConversationTurn) error {
	return nil
}
func (m *mockDynamo) UpdateVerificationStatusEnhanced(ctx context.Context, verificationID string, initialVerificationAt string, entry schema.StatusHistoryEntry) error {
	return m.statusErr
}
func (m *mockDynamo) RecordConversationHistory(ctx context.Context, ct *schema.ConversationTracker) error {
	return nil
}
func (m *mockDynamo) UpdateProcessingMetrics(ctx context.Context, verificationID string, metrics *schema.ProcessingMetrics) error {
	return nil
}
func (m *mockDynamo) UpdateStatusHistory(ctx context.Context, verificationID string, statusHistory []schema.StatusHistoryEntry) error {
	return nil
}
func (m *mockDynamo) UpdateErrorTracking(ctx context.Context, verificationID string, errorTracking *schema.ErrorTracking) error {
	return nil
}
func (m *mockDynamo) InitializeVerificationRecord(ctx context.Context, verificationContext *schema.VerificationContext) error {
	return nil
}
func (m *mockDynamo) UpdateCurrentStatus(ctx context.Context, verificationID, currentStatus, lastUpdatedAt string, metrics map[string]interface{}) error {
	return nil
}
func (m *mockDynamo) GetVerificationStatus(ctx context.Context, verificationID string) (*services.VerificationStatusInfo, error) {
	return nil, nil
}
func (m *mockDynamo) InitializeConversationHistory(ctx context.Context, verificationID string, maxTurns int, metadata map[string]interface{}) error {
	return nil
}
func (m *mockDynamo) UpdateConversationTurn(ctx context.Context, verificationID string, turnData *schema.TurnResponse) error {
	return m.turnErr
}
func (m *mockDynamo) CompleteConversation(ctx context.Context, verificationID string, conversationAt string, finalStatus string) error {
	return nil
}
func (m *mockDynamo) QueryPreviousVerification(ctx context.Context, checkingImageUrl string) (*schema.VerificationContext, error) {
	return nil, nil
}
func (m *mockDynamo) GetLayoutMetadata(ctx context.Context, layoutID int, layoutPrefix string) (*schema.LayoutMetadata, error) {
	return nil, nil
}

func TestDynamoManagerUpdateSuccess(t *testing.T) {
	mgr := NewDynamoManager(&mockDynamo{}, config.Config{}, logger.New("test", "test"))
	ok := mgr.Update(context.Background(), "id", "2025-05-30T00:00:00Z", schema.StatusHistoryEntry{}, &schema.TurnResponse{})
	if !ok {
		t.Errorf("expected true on success")
	}
}

func TestDynamoManagerUpdateFailure(t *testing.T) {
	mgr := NewDynamoManager(&mockDynamo{statusErr: errors.New("fail")}, config.Config{}, logger.New("test", "test"))
	ok := mgr.Update(context.Background(), "id", "2025-05-30T00:00:00Z", schema.StatusHistoryEntry{}, &schema.TurnResponse{})
	if ok {
		t.Errorf("expected false on failure")
	}
}
