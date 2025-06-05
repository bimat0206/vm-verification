package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/FinalizeAndStoreResults/internal/config"
	"workflow-function/FinalizeAndStoreResults/internal/dynamodbhelper"
	"workflow-function/FinalizeAndStoreResults/internal/models"
	"workflow-function/FinalizeAndStoreResults/internal/parser"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

var (
	appConfig    *config.LambdaConfig
	awsClients   *config.AWSClients
	log          logger.Logger
	stateManager s3state.Manager
)

// extractNestedReference extracts a reference from nested JSON structure
// This handles cases where the input has nested s3References like:
// {"s3References": {"responses": {"turn2Processed": {...}}}}
func extractNestedReference(input interface{}, category, refName string) *s3state.Reference {
	// Convert input to map[string]interface{}
	var inputMap map[string]interface{}

	switch v := input.(type) {
	case map[string]interface{}:
		inputMap = v
	case []byte:
		if err := json.Unmarshal(v, &inputMap); err != nil {
			return nil
		}
	case string:
		if err := json.Unmarshal([]byte(v), &inputMap); err != nil {
			return nil
		}
	default:
		return nil
	}

	// Navigate to s3References
	s3Refs, ok := inputMap["s3References"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Navigate to category (e.g., "responses")
	categoryRefs, ok := s3Refs[category].(map[string]interface{})
	if !ok {
		return nil
	}

	// Get the specific reference (e.g., "turn2Processed")
	refData, ok := categoryRefs[refName].(map[string]interface{})
	if !ok {
		return nil
	}

	// Extract bucket, key, and size
	bucket, _ := refData["bucket"].(string)
	key, _ := refData["key"].(string)
	var size int64
	if sizeVal, ok := refData["size"]; ok {
		switch s := sizeVal.(type) {
		case float64:
			size = int64(s)
		case int64:
			size = s
		case int:
			size = int64(s)
		}
	}

	if bucket == "" || key == "" {
		return nil
	}

	return &s3state.Reference{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	}
}

// parseInputAndExtractReferences parses the input and extracts required references
func parseInputAndExtractReferences(input interface{}) (*s3state.Envelope, *s3state.Reference, *s3state.Reference, error) {
	// Convert input to map[string]interface{}
	var inputMap map[string]interface{}

	switch v := input.(type) {
	case map[string]interface{}:
		inputMap = v
	case []byte:
		if err := json.Unmarshal(v, &inputMap); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to unmarshal input bytes: %w", err)
		}
	case string:
		if err := json.Unmarshal([]byte(v), &inputMap); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to unmarshal input string: %w", err)
		}
	default:
		return nil, nil, nil, fmt.Errorf("unsupported input type: %T", input)
	}

	// Extract basic envelope fields
	verificationID, _ := inputMap["verificationId"].(string)
	status, _ := inputMap["status"].(string)
	if verificationID == "" {
		return nil, nil, nil, fmt.Errorf("verificationId is required")
	}

	// Navigate to s3References
	s3Refs, ok := inputMap["s3References"].(map[string]interface{})
	if !ok {
		return nil, nil, nil, fmt.Errorf("s3References not found or invalid")
	}

	// Extract processing_initialization reference (top-level)
	initRef := extractReferenceFromMap(s3Refs, "processing_initialization")
	if initRef == nil {
		return nil, nil, nil, fmt.Errorf("processing_initialization reference not found")
	}

	// Extract turn2Processed reference from nested responses structure
	turn2Ref := extractNestedReference(input, "responses", "turn2Processed")
	if turn2Ref == nil {
		return nil, nil, nil, fmt.Errorf("turn2Processed reference not found in responses")
	}

	// Create envelope with flat references structure for compatibility
	envelope := &s3state.Envelope{
		VerificationID: verificationID,
		Status:         status,
		References:     make(map[string]*s3state.Reference),
		Summary:        make(map[string]interface{}),
	}

	// Add references to envelope for compatibility with existing code
	envelope.References["processing_initialization"] = initRef
	envelope.References["turn2Processed"] = turn2Ref

	return envelope, initRef, turn2Ref, nil
}

// extractReferenceFromMap extracts a reference from a map
func extractReferenceFromMap(refMap map[string]interface{}, refName string) *s3state.Reference {
	refData, ok := refMap[refName].(map[string]interface{})
	if !ok {
		return nil
	}

	bucket, _ := refData["bucket"].(string)
	key, _ := refData["key"].(string)
	var size int64
	if sizeVal, ok := refData["size"]; ok {
		switch s := sizeVal.(type) {
		case float64:
			size = int64(s)
		case int64:
			size = s
		case int:
			size = int64(s)
		}
	}

	if bucket == "" || key == "" {
		return nil
	}

	return &s3state.Reference{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	}
}

func init() {
	var err error
	log = logger.New("FinalizeAndStoreResults", "main")

	appConfig, err = config.LoadEnvConfig()
	if err != nil {
		log.Error("failed_to_load_config", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	awsClients, err = config.NewAWSClients(context.Background())
	if err != nil {
		log.Error("failed_to_init_aws", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	stateManager, err = s3state.New(appConfig.StateBucket)
	if err != nil {
		log.Error("failed_to_init_s3state", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

func HandleRequest(ctx context.Context, input interface{}) (*s3state.Envelope, error) {
	// Parse input to extract references and create envelope
	envelope, initRef, turn2Ref, err := parseInputAndExtractReferences(input)
	if err != nil {
		wfErr := errors.NewValidationError("failed to parse input", map[string]interface{}{
			"error": err.Error(),
		})
		log.Error("input_parse_failed", map[string]interface{}{"error": wfErr.Error()})
		return nil, wfErr
	}

	log.LogReceivedEvent(envelope)

	log.Info("extracted_references", map[string]interface{}{
		"init_bucket":  initRef.Bucket,
		"init_key":     initRef.Key,
		"turn2_bucket": turn2Ref.Bucket,
		"turn2_key":    turn2Ref.Key,
		"turn2_size":   turn2Ref.Size,
	})

	// Load initialization data using s3state manager
	var initData models.InitializationData
	err = stateManager.RetrieveJSON(initRef, &initData)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeS3, "failed to load initialization data", false)
		wfErr.VerificationID = envelope.VerificationID
		log.Error("fetch_init_failed", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
			"bucket":         initRef.Bucket,
			"key":            initRef.Key,
		})
		return nil, wfErr
	}

	// Load turn2 processed data using s3state manager
	turn2Bytes, err := stateManager.Retrieve(turn2Ref)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeS3, "failed to load turn2 processed data", false)
		wfErr.VerificationID = envelope.VerificationID
		log.Error("fetch_turn2_failed", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
			"bucket":         turn2Ref.Bucket,
			"key":            turn2Ref.Key,
		})
		return nil, wfErr
	}

	// Parse turn2 response data
	parsed, err := parser.ParseTurn2ResponseData(turn2Bytes)
	if err != nil {
		wfErr := errors.NewParsingError("turn2 response data", err)
		wfErr.VerificationID = envelope.VerificationID
		log.Error("parse_turn2_failed", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
		})
		return nil, wfErr
	}

	// Create DynamoDB item using shared schema constants
	now := schema.FormatISO8601()
	item := models.VerificationResultItem{
		VerificationID:         envelope.VerificationID,
		VerificationAt:         now,
		VerificationType:       initData.VerificationType,
		LayoutID:               &initData.LayoutID,
		LayoutPrefix:           initData.LayoutPrefix,
		VendingMachineID:       initData.VendingMachineID,
		ReferenceImageUrl:      initData.ReferenceImageUrl,
		CheckingImageUrl:       initData.CheckingImageUrl,
		VerificationStatus:     parsed.VerificationStatus,
		CurrentStatus:          schema.StatusCompleted,
		LastUpdatedAt:          now,
		ProcessingStartedAt:    initData.ProcessingStartedAt,
		InitialConfirmation:    parsed.InitialConfirmation,
		VerificationSummary:    parsed.VerificationSummary,
		PreviousVerificationID: initData.PreviousVerificationID,
	}
	if initData.LayoutID == 0 {
		item.LayoutID = nil
	}

	// Store verification result in DynamoDB
	err = dynamodbhelper.StoreVerificationResult(ctx, awsClients.DynamoDBClient, appConfig.VerificationResultsTable, item)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeDynamoDB, "failed to store verification result", false)
		wfErr.VerificationID = envelope.VerificationID
		log.Error("dynamodb_store_failed", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
			"table":          appConfig.VerificationResultsTable,
		})
		return nil, wfErr
	}

	// Update conversation history in DynamoDB
	err = dynamodbhelper.UpdateConversationHistory(ctx, awsClients.DynamoDBClient, appConfig.ConversationHistoryTable, envelope.VerificationID, initData.InitialVerificationAt)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeDynamoDB, "failed to update conversation history", false)
		wfErr.VerificationID = envelope.VerificationID
		log.Error("conversation_update_failed", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
			"table":          appConfig.ConversationHistoryTable,
		})
		return nil, wfErr
	}

	// Update envelope status and add summary
	envelope.SetStatus(schema.StatusCompleted)
	if envelope.Summary == nil {
		envelope.Summary = make(map[string]interface{})
	}
	envelope.Summary["verificationStatus"] = parsed.VerificationStatus
	envelope.Summary["verificationAt"] = now
	envelope.Summary["message"] = "Verification finalized and stored"

	log.LogOutputEvent(envelope)
	return envelope, nil
}

func main() {
	lambda.Start(HandleRequest)
}
