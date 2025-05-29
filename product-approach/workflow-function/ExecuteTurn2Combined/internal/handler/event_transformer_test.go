package handler

import (
	"context"
	"encoding/json"
	"testing"

	"workflow-function/ExecuteTurn2Combined/internal/bedrockparser"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	sharedlogger "workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// fakeS3Manager is a minimal implementation of services.S3StateManager for tests.
type fakeS3Manager struct {
	initData  *services.InitializationData
	imageMeta *services.ImageMetadata
}

func (f *fakeS3Manager) LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error) {
	return "", nil
}
func (f *fakeS3Manager) LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error) {
	return "", nil
}
func (f *fakeS3Manager) LoadJSON(ctx context.Context, ref models.S3Reference, target interface{}) error {
	return nil
}
func (f *fakeS3Manager) StoreJSONAtReference(ctx context.Context, ref models.S3Reference, data interface{}) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) LoadInitializationData(ctx context.Context, ref models.S3Reference) (*services.InitializationData, error) {
	return f.initData, nil
}
func (f *fakeS3Manager) LoadImageMetadata(ctx context.Context, ref models.S3Reference) (*services.ImageMetadata, error) {
	return f.imageMeta, nil
}
func (f *fakeS3Manager) LoadLayoutMetadata(ctx context.Context, ref models.S3Reference) (*schema.LayoutMetadata, error) {
	return nil, nil
}
func (f *fakeS3Manager) ValidateInitializationData(ctx context.Context, data *services.InitializationData) error {
	return nil
}
func (f *fakeS3Manager) ValidateImageMetadata(ctx context.Context, metadata *services.ImageMetadata) error {
	return nil
}
func (f *fakeS3Manager) StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreProcessedAnalysis(ctx context.Context, verificationID string, analysis interface{}) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StorePrompt(ctx context.Context, verificationID string, turn int, prompt interface{}) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreProcessedTurn1Response(ctx context.Context, verificationID string, analysisData *bedrockparser.ParsedTurn1Data) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreProcessedTurn1Markdown(ctx context.Context, verificationID string, markdownContent string) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreConversationTurn(ctx context.Context, verificationID string, turnData *schema.TurnResponse) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreTemplateProcessor(ctx context.Context, verificationID string, processor *schema.TemplateProcessor) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreProcessingMetrics(ctx context.Context, verificationID string, metrics *schema.ProcessingMetrics) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) LoadProcessingState(ctx context.Context, verificationID string, stateType string) (interface{}, error) {
	return nil, nil
}
func (f *fakeS3Manager) LoadTurn1ProcessedResponse(ctx context.Context, ref models.S3Reference) (*schema.Turn1ProcessedResponse, error) {
	return nil, nil
}
func (f *fakeS3Manager) LoadTurn1RawResponse(ctx context.Context, ref models.S3Reference) (json.RawMessage, error) {
	return nil, nil
}
func (f *fakeS3Manager) StoreTurn2Response(ctx context.Context, verificationID string, response *bedrockparser.ParsedTurn2Data) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreTurn2Markdown(ctx context.Context, verificationID string, markdownContent string) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreTurn2RawResponse(ctx context.Context, verificationID string, raw interface{}) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreTurn2ProcessedResponse(ctx context.Context, verificationID string, processed *bedrockparser.ParsedTurn2Data) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) StoreWorkflowState(ctx context.Context, verificationID string, state *schema.WorkflowState) (models.S3Reference, error) {
	return models.S3Reference{}, nil
}
func (f *fakeS3Manager) LoadWorkflowState(ctx context.Context, verificationID string) (*schema.WorkflowState, error) {
	return nil, nil
}

func TestTransformStepFunctionEvent_AdjustsInitPath(t *testing.T) {
	payload := `{"schemaVersion":"2.1.0","s3References":{"processing_initialization":{"bucket":"kootoro-dev-s3-state-f6d3xl","key":"2025/05/29/verif-20250529035808-9533/initialization.json"},"images_metadata":{"bucket":"b","key":"m"},"prompts_system":{"bucket":"b","key":"p"}},"verificationId":"verif-20250529035808-9533","status":"TURN1_COMPLETED"}`
	var event StepFunctionEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	initData := &services.InitializationData{
		SchemaVersion:       schema.SchemaVersion,
		VerificationContext: schema.VerificationContext{VerificationId: event.VerificationID, VerificationType: schema.VerificationTypeLayoutVsChecking},
		SystemPrompt:        services.SystemPromptData{Content: ""},
	}
	imageMeta := &services.ImageMetadata{
		CheckingImage: services.ImageInfoEnhanced{StorageMetadata: services.StorageImageMetadata{Bucket: "b", Key: "checking", StoredSize: 1}, OriginalMetadata: services.OriginalImageMetadata{ContentType: "image/png", SourceKey: "c.png"}},
	}

	s3 := &fakeS3Manager{initData: initData, imageMeta: imageMeta}
	log := sharedlogger.New("test", "transform")
	transformer := NewEventTransformer(s3, log)

	req, err := transformer.TransformStepFunctionEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("transform error: %v", err)
	}

	expectedKey := "2025/05/29/verif-20250529035808-9533/processing/initialization.json"
	if req.InputInitializationFileRef.Key != expectedKey {
		t.Errorf("expected key %s, got %s", expectedKey, req.InputInitializationFileRef.Key)
	}
	if req.InputInitializationFileRef.Bucket != "kootoro-dev-s3-state-f6d3xl" {
		t.Errorf("bucket changed")
	}
}
