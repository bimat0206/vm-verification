package handler

import (
	"context"
	"errors"
	"testing"
	"time"
	"workflow-function/ExecuteTurn1Combined/internal/config"
	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// MockDynamoDBService is a mock implementation of the DynamoDBService interface
type MockDynamoDBService struct {
	UpdateVerificationStatusEnhancedFunc func(ctx context.Context, verificationID string, statusEntry schema.StatusHistoryEntry) error
	UpdateConversationTurnFunc           func(ctx context.Context, verificationID string, turnData *schema.TurnResponse) error
}

// Implement all required methods of the DynamoDBService interface
func (m *MockDynamoDBService) UpdateVerificationStatus(ctx context.Context, verificationID string, status models.VerificationStatus, metrics models.TokenUsage) error {
	return nil
}

func (m *MockDynamoDBService) RecordConversationTurn(ctx context.Context, turn *models.ConversationTurn) error {
	return nil
}

func (m *MockDynamoDBService) UpdateVerificationStatusEnhanced(ctx context.Context, verificationID string, statusEntry schema.StatusHistoryEntry) error {
	if m.UpdateVerificationStatusEnhancedFunc != nil {
		return m.UpdateVerificationStatusEnhancedFunc(ctx, verificationID, statusEntry)
	}
	return nil
}

func (m *MockDynamoDBService) RecordConversationHistory(ctx context.Context, conversationTracker *schema.ConversationTracker) error {
	return nil
}

func (m *MockDynamoDBService) UpdateProcessingMetrics(ctx context.Context, verificationID string, metrics *schema.ProcessingMetrics) error {
	return nil
}

func (m *MockDynamoDBService) UpdateStatusHistory(ctx context.Context, verificationID string, statusHistory []schema.StatusHistoryEntry) error {
	return nil
}

func (m *MockDynamoDBService) UpdateErrorTracking(ctx context.Context, verificationID string, errorTracking *schema.ErrorTracking) error {
	return nil
}

func (m *MockDynamoDBService) InitializeVerificationRecord(ctx context.Context, verificationContext *schema.VerificationContext) error {
	return nil
}

func (m *MockDynamoDBService) UpdateCurrentStatus(ctx context.Context, verificationID, currentStatus, lastUpdatedAt string, metrics map[string]interface{}) error {
	return nil
}

func (m *MockDynamoDBService) GetVerificationStatus(ctx context.Context, verificationID string) (*services.VerificationStatusInfo, error) {
	return nil, nil
}

func (m *MockDynamoDBService) InitializeConversationHistory(ctx context.Context, verificationID string, maxTurns int, metadata map[string]interface{}) error {
	return nil
}

func (m *MockDynamoDBService) UpdateConversationTurn(ctx context.Context, verificationID string, turnData *schema.TurnResponse) error {
	if m.UpdateConversationTurnFunc != nil {
		return m.UpdateConversationTurnFunc(ctx, verificationID, turnData)
	}
	return nil
}

func (m *MockDynamoDBService) CompleteConversation(ctx context.Context, verificationID string, finalStatus string) error {
	return nil
}

func (m *MockDynamoDBService) QueryPreviousVerification(ctx context.Context, checkingImageUrl string) (*schema.VerificationContext, error) {
	return nil, nil
}

func (m *MockDynamoDBService) GetLayoutMetadata(ctx context.Context, layoutID int, layoutPrefix string) (*schema.LayoutMetadata, error) {
	return nil, nil
}

// MockLogger is a mock implementation of the logger.Logger interface
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields map[string]interface{}) {}
func (m *MockLogger) Info(msg string, fields map[string]interface{})  {}
func (m *MockLogger) Warn(msg string, fields map[string]interface{})  {}
func (m *MockLogger) Error(msg string, fields map[string]interface{}) {}
func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	return m
}
func (m *MockLogger) WithCorrelationId(correlationID string) logger.Logger {
	return m
}
func (m *MockLogger) LogReceivedEvent(event interface{}) {}
func (m *MockLogger) LogOutputEvent(event interface{})   {}

func TestUpdateAsyncWithThrottling(t *testing.T) {
	// Create a mock DynamoDB service that simulates throttling
	mockDynamo := &MockDynamoDBService{
		UpdateVerificationStatusEnhancedFunc: func(ctx context.Context, verificationID string, statusEntry schema.StatusHistoryEntry) error {
			return errors.New("throttling error")
		},
		UpdateConversationTurnFunc: func(ctx context.Context, verificationID string, turnData *schema.TurnResponse) error {
			return errors.New("throttling error")
		},
	}

	// Create a mock logger
	mockLogger := &MockLogger{}

	// Create a test config
	cfg := createDynamoTestConfig()

	// Create a DynamoManager with the mock DynamoDB service
	dynamoManager := NewDynamoManager(mockDynamo, cfg, mockLogger)

	// Create test data
	verificationID := "test-verification-id"
	tokenUsage := models.TokenUsage{
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
	}
	requestID := "test-request-id"
	rawRef := models.S3Reference{
		Bucket: "test-bucket",
		Key:    "test-raw-ref",
		ETag:   "test-etag",
		Size:   200,
	}
	procRef := models.S3Reference{
		Bucket: "test-bucket",
		Key:    "test-proc-ref",
		ETag:   "test-etag",
		Size:   300,
	}

	// Call UpdateAsync
	updateComplete, dynamoOK := dynamoManager.UpdateAsync(
		context.Background(),
		verificationID,
		tokenUsage,
		requestID,
		rawRef,
		procRef,
	)

	// Wait for the goroutine to complete
	dynamoManager.WaitForUpdates(updateComplete, 1*time.Second, mockLogger)

	// Assert that dynamoOK is false due to throttling
	if *dynamoOK {
		t.Errorf("Expected dynamoOK to be false, got true")
	}
	
	// Test that the flag is properly passed to the response builder
	// Create a mock response builder
	responseBuilder := &MockResponseBuilder{
		BuildCombinedTurnResponseFunc: func(
			req *models.Turn1Request,
			renderedPrompt string,
			promptRef, rawRef, procRef models.S3Reference,
			invoke *models.BedrockResponse,
			stages []schema.ProcessingStage,
			totalDurationMs int64,
			dynamoOK *bool,
		) *schema.CombinedTurnResponse {
			// Assert that dynamoOK is false
			if *dynamoOK {
				t.Errorf("Expected dynamoOK passed to response builder to be false, got true")
			}
			
			// Return a minimal response
			return &schema.CombinedTurnResponse{
				TurnResponse: &schema.TurnResponse{
					TurnId: 1,
				},
			}
		},
	}
	
	// Create a test request
	req := &models.Turn1Request{
		VerificationID: verificationID,
		VerificationContext: models.VerificationContext{
			VerificationType: "LAYOUT_VS_CHECKING",
		},
		S3Refs: models.Turn1RequestS3Refs{
			Prompts: models.PromptRefs{
				System: models.S3Reference{
					Bucket: "test-bucket",
					Key:    "system-prompt.json",
				},
			},
			Images: models.ImageRefs{
				ReferenceBase64: models.S3Reference{
					Bucket: "test-bucket",
					Key:    "reference-base64.json",
				},
			},
		},
	}
	
	// Create a test Bedrock response
	invoke := &models.BedrockResponse{
		RequestID: requestID,
		TokenUsage: tokenUsage,
		Raw: []byte("test raw response"),
		Processed: map[string]interface{}{"key": "value"},
	}
	
	// Create a test prompt reference
	testPromptRef := models.S3Reference{
		Bucket: "test-bucket",
		Key:    "test-prompt-ref",
		ETag:   "test-etag",
		Size:   100,
	}
	
	// Call the response builder with the dynamoOK flag
	responseBuilder.BuildCombinedTurnResponse(
		req, "test prompt", testPromptRef, rawRef, procRef,
		invoke, []schema.ProcessingStage{}, 1000, dynamoOK,
	)
}

// MockResponseBuilder is a mock implementation of the ResponseBuilder
type MockResponseBuilder struct {
	BuildCombinedTurnResponseFunc func(
		req *models.Turn1Request,
		renderedPrompt string,
		promptRef, rawRef, procRef models.S3Reference,
		invoke *models.BedrockResponse,
		stages []schema.ProcessingStage,
		totalDurationMs int64,
		dynamoOK *bool,
	) *schema.CombinedTurnResponse
}

// BuildCombinedTurnResponse is the mock implementation
func (m *MockResponseBuilder) BuildCombinedTurnResponse(
	req *models.Turn1Request,
	renderedPrompt string,
	promptRef, rawRef, procRef models.S3Reference,
	invoke *models.BedrockResponse,
	stages []schema.ProcessingStage,
	totalDurationMs int64,
	dynamoOK *bool,
) *schema.CombinedTurnResponse {
	if m.BuildCombinedTurnResponseFunc != nil {
		return m.BuildCombinedTurnResponseFunc(
			req, renderedPrompt, promptRef, rawRef, procRef,
			invoke, stages, totalDurationMs, dynamoOK,
		)
	}
	return nil
}

func TestUpdateAsyncWithSuccess(t *testing.T) {
	// Create a mock DynamoDB service that succeeds
	mockDynamo := &MockDynamoDBService{
		UpdateVerificationStatusEnhancedFunc: func(ctx context.Context, verificationID string, statusEntry schema.StatusHistoryEntry) error {
			return nil
		},
		UpdateConversationTurnFunc: func(ctx context.Context, verificationID string, turnData *schema.TurnResponse) error {
			return nil
		},
	}

	// Create a mock logger
	mockLogger := &MockLogger{}

	// Create a test config
	cfg := createDynamoTestConfig()

	// Create a DynamoManager with the mock DynamoDB service
	dynamoManager := NewDynamoManager(mockDynamo, cfg, mockLogger)

	// Create test data
	verificationID := "test-verification-id"
	tokenUsage := models.TokenUsage{
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
	}
	requestID := "test-request-id"
	rawRef := models.S3Reference{
		Bucket: "test-bucket",
		Key:    "test-raw-ref",
		ETag:   "test-etag",
		Size:   200,
	}
	procRef := models.S3Reference{
		Bucket: "test-bucket",
		Key:    "test-proc-ref",
		ETag:   "test-etag",
		Size:   300,
	}

	// Call UpdateAsync
	updateComplete, dynamoOK := dynamoManager.UpdateAsync(
		context.Background(),
		verificationID,
		tokenUsage,
		requestID,
		rawRef,
		procRef,
	)

	// Wait for the goroutine to complete
	dynamoManager.WaitForUpdates(updateComplete, 1*time.Second, mockLogger)

	// Assert that dynamoOK is true when there are no errors
	if !*dynamoOK {
		t.Errorf("Expected dynamoOK to be true, got false")
	}
}

// Helper function to create test config for dynamo_manager_test
func createDynamoTestConfig() config.Config {
	cfg := config.Config{}
	
	cfg.AWS.Region = "us-west-2"
	cfg.AWS.S3Bucket = "test-bucket"
	cfg.AWS.BedrockModel = "anthropic.claude-v2"
	cfg.AWS.AnthropicVersion = "bedrock-2023-05-31"
	cfg.AWS.DynamoDBVerificationTable = "test-verification-table"
	cfg.AWS.DynamoDBConversationTable = "test-conversation-table"
	
	cfg.Processing.MaxTokens = 24000
	cfg.Processing.BudgetTokens = 16000
	cfg.Processing.ThinkingType = "enable"
	cfg.Processing.MaxRetries = 3
	cfg.Processing.BedrockConnectTimeoutSec = 10
	cfg.Processing.BedrockCallTimeoutSec = 30
	
	cfg.Logging.Level = "INFO"
	cfg.Logging.Format = "json"
	
	cfg.Prompts.TemplateVersion = "1.0"
	cfg.Prompts.TemplateBasePath = "/opt/templates"
	
	return cfg
}
