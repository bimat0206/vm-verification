package main

import (
	"context"
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
	appConfig   *config.LambdaConfig
	awsClients  *config.AWSClients
	log         logger.Logger
	stateManager s3state.Manager
)

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
	// Load envelope from Step Functions input
	envelope, err := s3state.LoadEnvelope(input)
	if err != nil {
		wfErr := errors.NewValidationError("failed to load envelope", map[string]interface{}{
			"error": err.Error(),
		})
		log.Error("envelope_load_failed", map[string]interface{}{"error": wfErr.Error()})
		return nil, wfErr
	}

	log.LogReceivedEvent(envelope)

	// Validate required references
	initRef := envelope.GetReference("processing_initialization")
	if initRef == nil {
		wfErr := errors.NewMissingFieldError("processing_initialization reference")
		log.Error("missing_init_ref", map[string]interface{}{"error": wfErr.Error()})
		return nil, wfErr
	}

	turn2Ref := envelope.GetReference("turn2Processed")
	if turn2Ref == nil {
		wfErr := errors.NewMissingFieldError("turn2Processed reference")
		log.Error("missing_turn2_ref", map[string]interface{}{"error": wfErr.Error()})
		return nil, wfErr
	}

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
	err = dynamodbhelper.UpdateConversationHistory(ctx, awsClients.DynamoDBClient, appConfig.ConversationHistoryTable, envelope.VerificationID)
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
