package mocks

import (
	"github.com/stretchr/testify/mock"
	"workflow-function/ExecuteTurn1Combined/internal/models"
)

// TestFixtures contains all the test data
type TestFixtures struct {
	SystemPrompt     []byte
	ImageMetadata    []byte
	LayoutMetadata   []byte
	Initialization   []byte
	CheckingImageB64 string
	LayoutImageB64   string
}

// GetDefaultFixtures returns the standard test fixtures
func GetDefaultFixtures() *TestFixtures {
	return &TestFixtures{
		SystemPrompt: []byte(`{
			"promptContent": {
				"systemMessage": "You are an AI assistant helping to verify vending machine layouts.",
				"role": "system",
				"taskDescription": "Compare layout and checking images",
				"outputFormat": "JSON with overallAccuracy field"
			},
			"metadata": {
				"version": "1.0",
				"created": "2024-01-01T00:00:00Z"
			}
		}`),
		
		ImageMetadata: []byte(`{
			"images": [
				{
					"name": "layout-image.jpg",
					"url": "s3://test-bucket/images/layout.jpg",
					"width": 1920,
					"height": 1080,
					"size": 512000
				},
				{
					"name": "checking-image.jpg",
					"url": "s3://test-bucket/images/checking.jpg", 
					"width": 1920,
					"height": 1080,
					"size": 498000
				}
			],
			"totalImages": 2,
			"metadata": {
				"captureDate": "2024-01-01",
				"deviceId": "test-device"
			}
		}`),
		
		LayoutMetadata: []byte(`{
			"layoutId": 12345,
			"layoutName": "Standard Layout A",
			"layoutPrefix": "layout-a",
			"planogram": {
				"rows": 5,
				"columns": 6,
				"totalSlots": 30,
				"products": [
					{
						"slotId": "A1",
						"productId": "PROD001",
						"productName": "Coca Cola",
						"row": 0,
						"column": 0
					},
					{
						"slotId": "A2",
						"productId": "PROD002",
						"productName": "Pepsi",
						"row": 0,
						"column": 1
					}
				]
			},
			"metadata": {
				"version": "1.0",
				"lastUpdated": "2024-01-01T00:00:00Z"
			}
		}`),
		
		Initialization: []byte(`{
			"verificationContext": {
				"verificationId": "test-verification-123",
				"verificationType": "LAYOUT_VS_CHECKING",
				"vendingMachineId": "VM-001",
				"layoutId": 12345,
				"layoutPrefix": "layout-a",
				"checkingImageUrl": "s3://test-bucket/images/checking.jpg",
				"referenceImageUrl": "s3://test-bucket/images/layout.jpg"
			},
			"systemPrompt": {
				"promptId": "system-prompt-001",
				"promptVersion": "1.0",
				"s3Reference": {
					"bucket": "test-bucket",
					"key": "prompts/system-prompt.json"
				}
			},
			"turnConfig": {
				"turnNumber": 1,
				"maxTurns": 3,
				"turnType": "turn1"
			}
		}`),
		
		// 1x1 pixel transparent PNG
		CheckingImageB64: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
		LayoutImageB64:   "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
	}
}

// SetupMocksForHappyPath configures all mocks with expected behavior
func SetupMocksForHappyPath(mocks *MockSet) {
	fixtures := GetDefaultFixtures()
	
	// Setup S3 mock with fixtures
	mocks.S3.AddFixture("system-prompt", fixtures.SystemPrompt)
	mocks.S3.AddFixture("image-metadata", fixtures.ImageMetadata)
	mocks.S3.AddFixture("layout-metadata", fixtures.LayoutMetadata)
	mocks.S3.AddFixture("initialization", fixtures.Initialization)
	
	// Setup expected S3 calls
	mocks.S3.On("LoadSystemPrompt", mock.Anything, mock.Anything).Return("", nil)
	mocks.S3.On("LoadBase64Image", mock.Anything, mock.Anything).Return("", nil)
	mocks.S3.On("LoadJSON", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mocks.S3.On("LoadInitializationData", mock.Anything, mock.Anything).Return(nil, nil)
	mocks.S3.On("LoadImageMetadata", mock.Anything, mock.Anything).Return(nil, nil)
	mocks.S3.On("LoadLayoutMetadata", mock.Anything, mock.Anything).Return(nil, nil)
	mocks.S3.On("ValidateInitializationData", mock.Anything, mock.Anything).Return(nil)
	mocks.S3.On("ValidateImageMetadata", mock.Anything, mock.Anything).Return(nil)
	mocks.S3.On("StoreRawResponse", mock.Anything, mock.Anything, mock.Anything).Return(models.S3Reference{
		Bucket: "test-bucket",
		Key:    "processing/turn1-raw-response.json",
	}, nil)
	mocks.S3.On("StoreProcessedAnalysis", mock.Anything, mock.Anything, mock.Anything).Return(models.S3Reference{
		Bucket: "test-bucket",
		Key:    "processing/turn1-analysis.json",
	}, nil)
	mocks.S3.On("StoreConversationTurn", mock.Anything, mock.Anything, mock.Anything).Return(models.S3Reference{
		Bucket: "test-bucket", 
		Key:    "processing/turn1-conversation.json",
	}, nil)
	mocks.S3.On("StoreTemplateProcessor", mock.Anything, mock.Anything, mock.Anything).Return(models.S3Reference{
		Bucket: "test-bucket",
		Key:    "processing/turn1-template-processor.json",
	}, nil)
	mocks.S3.On("StoreWorkflowState", mock.Anything, mock.Anything, mock.Anything).Return(models.S3Reference{
		Bucket: "test-bucket",
		Key:    "processing/workflow-state.json",
	}, nil)
	
	// Setup Bedrock mock
	mocks.Bedrock.On("Converse", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	
	// Setup Dynamo mock
	mocks.Dynamo.On("UpdateVerificationStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mocks.Dynamo.On("UpdateVerificationStatusEnhanced", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mocks.Dynamo.On("UpdateConversationTurn", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	
	// Setup Prompt mock
	mocks.Prompt.On("GenerateTurn1PromptWithMetrics", mock.Anything, mock.Anything, mock.Anything).Return("", nil, nil)
	
	// Setup Template mock
	mocks.Template.On("RenderTemplate", mock.Anything, mock.Anything).Return("", nil)
	
	// Setup Logger mock
	mocks.Logger.On("Debug", mock.Anything, mock.Anything).Return()
	mocks.Logger.On("Info", mock.Anything, mock.Anything).Return()
	mocks.Logger.On("Warn", mock.Anything, mock.Anything).Return()
	mocks.Logger.On("Error", mock.Anything, mock.Anything).Return()
	mocks.Logger.On("WithCorrelationId", mock.Anything).Return(mocks.Logger)
	mocks.Logger.On("WithFields", mock.Anything).Return(mocks.Logger)
	
	// Connect all mocks
	mocks.ConnectMocks()
}