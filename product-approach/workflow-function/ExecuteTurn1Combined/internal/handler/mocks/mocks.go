package mocks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	
	"ExecuteTurn1Combined/internal/models"
	"ExecuteTurn1Combined/internal/services"
	"github.com/vending-machine-verification/shared/schema"
	"github.com/vending-machine-verification/shared/logger"
)

// MockSet contains all mock services needed for testing
type MockSet struct {
	S3           *MockS3StateManager
	Bedrock      *MockBedrockService
	Dynamo       *MockDynamoDBService
	Prompt       *MockPromptService
	Template     *MockTemplateLoader
	Logger       *MockLogger
	callRecords  sync.Map // thread-safe storage of all calls
}

// New creates a new set of mocks
func New() *MockSet {
	return &MockSet{
		S3:       NewMockS3StateManager(),
		Bedrock:  NewMockBedrockService(),
		Dynamo:   NewMockDynamoDBService(),
		Prompt:   NewMockPromptService(),
		Template: NewMockTemplateLoader(),
		Logger:   NewMockLogger(),
	}
}

// AssertCalled verifies a method was called with matching arguments
func (m *MockSet) AssertCalled(t *testing.T, methodName string, args ...interface{}) {
	found := false
	m.callRecords.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok && strings.HasSuffix(k, methodName) {
			found = true
			return false
		}
		return true
	})
	assert.True(t, found, "Method %s was not called", methodName)
}

// AssertLogContains verifies the logger recorded a specific message
func (m *MockSet) AssertLogContains(t *testing.T, substring string) {
	found := false
	m.Logger.logs.Range(func(key, value interface{}) bool {
		if entry, ok := value.(logEntry); ok {
			if strings.Contains(entry.message, substring) {
				found = true
				return false
			}
		}
		return true
	})
	assert.True(t, found, "Log does not contain: %s", substring)
}

// GetCallRecords returns all recorded calls for diagnostics
func (m *MockSet) GetCallRecords() map[string]interface{} {
	records := make(map[string]interface{})
	m.callRecords.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			records[k] = value
		}
		return true
	})
	return records
}

// MockS3StateManager implements services.S3StateManager
type MockS3StateManager struct {
	mock.Mock
	fixtures map[string][]byte
	parent   *MockSet
}

func NewMockS3StateManager() *MockS3StateManager {
	return &MockS3StateManager{
		fixtures: make(map[string][]byte),
	}
}

func (m *MockS3StateManager) SetParent(parent *MockSet) {
	m.parent = parent
}

func (m *MockS3StateManager) AddFixture(key string, data []byte) {
	m.fixtures[key] = data
}

func (m *MockS3StateManager) LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error) {
	args := m.Called(ctx, ref)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.LoadSystemPrompt", ref)
	}
	if data, ok := m.fixtures["system-prompt"]; ok {
		var prompt map[string]interface{}
		json.Unmarshal(data, &prompt)
		if content, ok := prompt["promptContent"].(map[string]interface{}); ok {
			if msg, ok := content["systemMessage"].(string); ok {
				return msg, nil
			}
		}
	}
	return args.String(0), args.Error(1)
}

func (m *MockS3StateManager) LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error) {
	args := m.Called(ctx, ref)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.LoadBase64Image", ref)
	}
	// Return mock base64 image data
	return "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==", args.Error(1)
}

func (m *MockS3StateManager) LoadJSON(ctx context.Context, ref models.S3Reference, target interface{}) error {
	args := m.Called(ctx, ref, target)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.LoadJSON", ref)
	}
	
	// Check fixtures based on the key
	key := ref.Key
	if strings.Contains(key, "metadata.json") {
		key = "image-metadata"
	} else if strings.Contains(key, "layout-metadata") {
		key = "layout-metadata"
	} else if strings.Contains(key, "initialization") {
		key = "initialization"
	}
	
	if data, ok := m.fixtures[key]; ok {
		return json.Unmarshal(data, target)
	}
	
	return args.Error(0)
}

func (m *MockS3StateManager) LoadInitializationData(ctx context.Context, ref models.S3Reference) (*services.InitializationData, error) {
	args := m.Called(ctx, ref)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.LoadInitializationData", ref)
	}
	if data, ok := m.fixtures["initialization"]; ok {
		var init services.InitializationData
		json.Unmarshal(data, &init)
		return &init, nil
	}
	return args.Get(0).(*services.InitializationData), args.Error(1)
}

func (m *MockS3StateManager) LoadImageMetadata(ctx context.Context, ref models.S3Reference) (*services.ImageMetadata, error) {
	args := m.Called(ctx, ref)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.LoadImageMetadata", ref)
	}
	if data, ok := m.fixtures["image-metadata"]; ok {
		var meta services.ImageMetadata
		json.Unmarshal(data, &meta)
		return &meta, nil
	}
	return args.Get(0).(*services.ImageMetadata), args.Error(1)
}

func (m *MockS3StateManager) LoadLayoutMetadata(ctx context.Context, ref models.S3Reference) (*schema.LayoutMetadata, error) {
	args := m.Called(ctx, ref)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.LoadLayoutMetadata", ref)
	}
	if data, ok := m.fixtures["layout-metadata"]; ok {
		var layout schema.LayoutMetadata
		json.Unmarshal(data, &layout)
		return &layout, nil
	}
	return args.Get(0).(*schema.LayoutMetadata), args.Error(1)
}

func (m *MockS3StateManager) ValidateInitializationData(ctx context.Context, data *services.InitializationData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockS3StateManager) ValidateImageMetadata(ctx context.Context, metadata *services.ImageMetadata) error {
	args := m.Called(ctx, metadata)
	return args.Error(0)
}

func (m *MockS3StateManager) StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error) {
	args := m.Called(ctx, verificationID, data)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.StoreRawResponse", map[string]interface{}{
			"verificationID": verificationID,
			"data": data,
		})
	}
	return args.Get(0).(models.S3Reference), args.Error(1)
}

func (m *MockS3StateManager) StoreProcessedAnalysis(ctx context.Context, verificationID string, analysis interface{}) (models.S3Reference, error) {
	args := m.Called(ctx, verificationID, analysis)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.StoreProcessedAnalysis", map[string]interface{}{
			"verificationID": verificationID,
			"analysis": analysis,
		})
	}
	return args.Get(0).(models.S3Reference), args.Error(1)
}

func (m *MockS3StateManager) StoreConversationTurn(ctx context.Context, verificationID string, turnData *schema.TurnResponse) (models.S3Reference, error) {
	args := m.Called(ctx, verificationID, turnData)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.StoreConversationTurn", map[string]interface{}{
			"verificationID": verificationID,
			"turnData": turnData,
		})
	}
	return args.Get(0).(models.S3Reference), args.Error(1)
}

func (m *MockS3StateManager) StoreTemplateProcessor(ctx context.Context, verificationID string, processor *schema.TemplateProcessor) (models.S3Reference, error) {
	args := m.Called(ctx, verificationID, processor)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.StoreTemplateProcessor", map[string]interface{}{
			"verificationID": verificationID,
			"processor": processor,
		})
	}
	return args.Get(0).(models.S3Reference), args.Error(1)
}

func (m *MockS3StateManager) StoreProcessingMetrics(ctx context.Context, verificationID string, metrics *schema.ProcessingMetrics) (models.S3Reference, error) {
	args := m.Called(ctx, verificationID, metrics)
	return args.Get(0).(models.S3Reference), args.Error(1)
}

func (m *MockS3StateManager) LoadProcessingState(ctx context.Context, verificationID string, stateType string) (interface{}, error) {
	args := m.Called(ctx, verificationID, stateType)
	return args.Get(0), args.Error(1)
}

func (m *MockS3StateManager) StoreWorkflowState(ctx context.Context, verificationID string, state *schema.WorkflowState) (models.S3Reference, error) {
	args := m.Called(ctx, verificationID, state)
	if m.parent != nil {
		m.parent.callRecords.Store("S3.StoreWorkflowState", map[string]interface{}{
			"verificationID": verificationID,
			"state": state,
		})
	}
	return args.Get(0).(models.S3Reference), args.Error(1)
}

func (m *MockS3StateManager) LoadWorkflowState(ctx context.Context, verificationID string) (*schema.WorkflowState, error) {
	args := m.Called(ctx, verificationID)
	return args.Get(0).(*schema.WorkflowState), args.Error(1)
}

// MockBedrockService implements services.BedrockService
type MockBedrockService struct {
	mock.Mock
	parent *MockSet
}

func NewMockBedrockService() *MockBedrockService {
	return &MockBedrockService{}
}

func (m *MockBedrockService) SetParent(parent *MockSet) {
	m.parent = parent
}

func (m *MockBedrockService) Converse(ctx context.Context, systemPrompt, turnPrompt, base64Image string) (*models.BedrockResponse, error) {
	args := m.Called(ctx, systemPrompt, turnPrompt, base64Image)
	if m.parent != nil {
		m.parent.callRecords.Store("Bedrock.Converse", map[string]interface{}{
			"systemPrompt": systemPrompt,
			"turnPrompt": turnPrompt,
			"hasImage": base64Image != "",
		})
	}
	// Return default happy path response
	return &models.BedrockResponse{
		Content: `{"overallAccuracy": 0.98}`,
		Usage: models.TokenUsage{
			InputTokens:  500,
			OutputTokens: 42,
		},
	}, args.Error(1)
}

// MockDynamoDBService implements services.DynamoDBService
type MockDynamoDBService struct {
	mock.Mock
	parent *MockSet
}

func NewMockDynamoDBService() *MockDynamoDBService {
	return &MockDynamoDBService{}
}

func (m *MockDynamoDBService) SetParent(parent *MockSet) {
	m.parent = parent
}

func (m *MockDynamoDBService) UpdateVerificationStatus(ctx context.Context, verificationID string, status models.VerificationStatus, metrics models.TokenUsage) error {
	args := m.Called(ctx, verificationID, status, metrics)
	if m.parent != nil {
		m.parent.callRecords.Store("Dynamo.UpdateVerificationStatus", map[string]interface{}{
			"verificationID": verificationID,
			"status": status,
			"metrics": metrics,
		})
	}
	return args.Error(0)
}

func (m *MockDynamoDBService) RecordConversationTurn(ctx context.Context, turn *models.ConversationTurn) error {
	args := m.Called(ctx, turn)
	return args.Error(0)
}

func (m *MockDynamoDBService) UpdateVerificationStatusEnhanced(ctx context.Context, verificationID string, statusEntry schema.StatusHistoryEntry) error {
	args := m.Called(ctx, verificationID, statusEntry)
	if m.parent != nil {
		m.parent.callRecords.Store("Dynamo.UpdateVerificationStatusEnhanced", map[string]interface{}{
			"verificationID": verificationID,
			"status": statusEntry.Status,
		})
	}
	return args.Error(0)
}

func (m *MockDynamoDBService) RecordConversationHistory(ctx context.Context, conversationTracker *schema.ConversationTracker) error {
	args := m.Called(ctx, conversationTracker)
	return args.Error(0)
}

func (m *MockDynamoDBService) UpdateProcessingMetrics(ctx context.Context, verificationID string, metrics *schema.ProcessingMetrics) error {
	args := m.Called(ctx, verificationID, metrics)
	return args.Error(0)
}

func (m *MockDynamoDBService) UpdateStatusHistory(ctx context.Context, verificationID string, statusHistory []schema.StatusHistoryEntry) error {
	args := m.Called(ctx, verificationID, statusHistory)
	return args.Error(0)
}

func (m *MockDynamoDBService) UpdateErrorTracking(ctx context.Context, verificationID string, errorTracking *schema.ErrorTracking) error {
	args := m.Called(ctx, verificationID, errorTracking)
	return args.Error(0)
}

func (m *MockDynamoDBService) InitializeVerificationRecord(ctx context.Context, verificationContext *schema.VerificationContext) error {
	args := m.Called(ctx, verificationContext)
	return args.Error(0)
}

func (m *MockDynamoDBService) UpdateCurrentStatus(ctx context.Context, verificationID, currentStatus, lastUpdatedAt string, metrics map[string]interface{}) error {
	args := m.Called(ctx, verificationID, currentStatus, lastUpdatedAt, metrics)
	return args.Error(0)
}

func (m *MockDynamoDBService) GetVerificationStatus(ctx context.Context, verificationID string) (*services.VerificationStatusInfo, error) {
	args := m.Called(ctx, verificationID)
	return args.Get(0).(*services.VerificationStatusInfo), args.Error(1)
}

func (m *MockDynamoDBService) InitializeConversationHistory(ctx context.Context, verificationID string, maxTurns int, metadata map[string]interface{}) error {
	args := m.Called(ctx, verificationID, maxTurns, metadata)
	return args.Error(0)
}

func (m *MockDynamoDBService) UpdateConversationTurn(ctx context.Context, verificationID string, turnData *schema.TurnResponse) error {
	args := m.Called(ctx, verificationID, turnData)
	return args.Error(0)
}

func (m *MockDynamoDBService) CompleteConversation(ctx context.Context, verificationID string, finalStatus string) error {
	args := m.Called(ctx, verificationID, finalStatus)
	return args.Error(0)
}

func (m *MockDynamoDBService) QueryPreviousVerification(ctx context.Context, checkingImageUrl string) (*schema.VerificationContext, error) {
	args := m.Called(ctx, checkingImageUrl)
	return args.Get(0).(*schema.VerificationContext), args.Error(1)
}

func (m *MockDynamoDBService) GetLayoutMetadata(ctx context.Context, layoutID int, layoutPrefix string) (*schema.LayoutMetadata, error) {
	args := m.Called(ctx, layoutID, layoutPrefix)
	return args.Get(0).(*schema.LayoutMetadata), args.Error(1)
}

// MockPromptService implements services.PromptService
type MockPromptService struct {
	mock.Mock
	parent *MockSet
}

func NewMockPromptService() *MockPromptService {
	return &MockPromptService{}
}

func (m *MockPromptService) SetParent(parent *MockSet) {
	m.parent = parent
}

func (m *MockPromptService) GenerateTurn1Prompt(ctx context.Context, vCtx models.VerificationContext, systemPrompt string) (string, error) {
	args := m.Called(ctx, vCtx, systemPrompt)
	return args.String(0), args.Error(1)
}

func (m *MockPromptService) GenerateTurn1PromptWithMetrics(ctx context.Context, vCtx models.VerificationContext, systemPrompt string) (string, *schema.TemplateProcessor, error) {
	args := m.Called(ctx, vCtx, systemPrompt)
	if m.parent != nil {
		m.parent.callRecords.Store("Prompt.GenerateTurn1PromptWithMetrics", vCtx)
	}
	// Return mock prompt and metrics
	metrics := &schema.TemplateProcessor{
		TemplateType:    "turn1-layout-vs-checking",
		TemplateVersion: "v1.0",
		ProcessingTime:  100,
		InputTokens:     500,
		OutputTokens:    42,
		Variables: map[string]interface{}{
			"verificationId": vCtx.VerificationID,
		},
	}
	return "MOCK_PROMPT", metrics, args.Error(2)
}

// MockTemplateLoader implements shared/templateloader.TemplateLoader
type MockTemplateLoader struct {
	mock.Mock
	parent *MockSet
}

func NewMockTemplateLoader() *MockTemplateLoader {
	return &MockTemplateLoader{}
}

func (m *MockTemplateLoader) SetParent(parent *MockSet) {
	m.parent = parent
}

func (m *MockTemplateLoader) LoadTemplate(templateType string) (*template.Template, error) {
	args := m.Called(templateType)
	return args.Get(0).(*template.Template), args.Error(1)
}

func (m *MockTemplateLoader) LoadTemplateWithVersion(templateType, version string) (*template.Template, error) {
	args := m.Called(templateType, version)
	return args.Get(0).(*template.Template), args.Error(1)
}

func (m *MockTemplateLoader) RenderTemplate(templateType string, data interface{}) (string, error) {
	args := m.Called(templateType, data)
	if m.parent != nil {
		m.parent.callRecords.Store("Template.RenderTemplate", map[string]interface{}{
			"templateType": templateType,
			"data": data,
		})
	}
	return "MOCK_PROMPT", args.Error(1)
}

func (m *MockTemplateLoader) RenderTemplateWithVersion(templateType, version string, data interface{}) (string, error) {
	args := m.Called(templateType, version, data)
	return args.String(0), args.Error(1)
}

func (m *MockTemplateLoader) GetLatestVersion(templateType string) string {
	args := m.Called(templateType)
	return args.String(0)
}

func (m *MockTemplateLoader) ListVersions(templateType string) []string {
	args := m.Called(templateType)
	return args.Get(0).([]string)
}

func (m *MockTemplateLoader) ClearCache() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTemplateLoader) RefreshVersions() error {
	args := m.Called()
	return args.Error(0)
}

// MockLogger implements logger.Logger
type MockLogger struct {
	mock.Mock
	parent *MockSet
	logs   sync.Map
}

type logEntry struct {
	level   string
	message string
	details map[string]interface{}
}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) SetParent(parent *MockSet) {
	m.parent = parent
}

func (m *MockLogger) Debug(message string, details map[string]interface{}) {
	m.Called(message, details)
	m.logs.Store(fmt.Sprintf("debug_%d", m.logCount()), logEntry{"debug", message, details})
}

func (m *MockLogger) Info(message string, details map[string]interface{}) {
	m.Called(message, details)
	m.logs.Store(fmt.Sprintf("info_%d", m.logCount()), logEntry{"info", message, details})
	if m.parent != nil && strings.Contains(message, "bedrock_invoke_completed") {
		m.parent.callRecords.Store("Logger.bedrock_invoke_completed", details)
	}
}

func (m *MockLogger) Warn(message string, details map[string]interface{}) {
	m.Called(message, details)
	m.logs.Store(fmt.Sprintf("warn_%d", m.logCount()), logEntry{"warn", message, details})
}

func (m *MockLogger) Error(message string, details map[string]interface{}) {
	m.Called(message, details)
	m.logs.Store(fmt.Sprintf("error_%d", m.logCount()), logEntry{"error", message, details})
}

func (m *MockLogger) LogReceivedEvent(event interface{}) {
	m.Called(event)
}

func (m *MockLogger) LogOutputEvent(event interface{}) {
	m.Called(event)
}

func (m *MockLogger) WithCorrelationId(correlationId string) logger.Logger {
	args := m.Called(correlationId)
	return m
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	args := m.Called(fields)
	return m
}

func (m *MockLogger) logCount() int {
	count := 0
	m.logs.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// ConnectMocks sets up parent references for all mocks
func (m *MockSet) ConnectMocks() {
	m.S3.SetParent(m)
	m.Bedrock.SetParent(m)
	m.Dynamo.SetParent(m)
	m.Prompt.SetParent(m)
	m.Template.SetParent(m)
	m.Logger.SetParent(m)
}