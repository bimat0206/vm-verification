package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/FinalizeAndStoreResults/internal/config"
	"workflow-function/FinalizeAndStoreResults/internal/dynamodbhelper"
	"workflow-function/FinalizeAndStoreResults/internal/models"
	"workflow-function/FinalizeAndStoreResults/internal/parser"
	"workflow-function/FinalizeAndStoreResults/internal/s3helper"
	"workflow-function/shared/logger"
)

var (
	appConfig  *config.LambdaConfig
	awsClients *config.AWSClients
	log        logger.Logger
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
}

func HandleRequest(ctx context.Context, event models.LambdaInput) (models.LambdaOutput, error) {
	log.LogReceivedEvent(event)

	if event.References.InitializationS3Ref.Bucket == "" || event.References.InitializationS3Ref.Key == "" {
		return models.LambdaOutput{}, fmt.Errorf("initialization S3 reference missing")
	}
	if event.References.Turn2ProcessedS3Ref.Bucket == "" || event.References.Turn2ProcessedS3Ref.Key == "" {
		return models.LambdaOutput{}, fmt.Errorf("turn2 processed S3 reference missing")
	}

	var initData models.InitializationData
	err := s3helper.GetS3ObjectAsJSON(ctx, awsClients.S3Client, event.References.InitializationS3Ref.Bucket, event.References.InitializationS3Ref.Key, &initData)
	if err != nil {
		log.Error("fetch_init_failed", map[string]interface{}{"error": err.Error(), "verificationId": event.VerificationID})
		return models.LambdaOutput{}, err
	}

	turn2Bytes, err := s3helper.GetS3Object(ctx, awsClients.S3Client, event.References.Turn2ProcessedS3Ref.Bucket, event.References.Turn2ProcessedS3Ref.Key)
	if err != nil {
		log.Error("fetch_turn2_failed", map[string]interface{}{"error": err.Error(), "verificationId": event.VerificationID})
		return models.LambdaOutput{}, err
	}

	parsed, err := parser.ParseTurn2ResponseData(turn2Bytes)
	if err != nil {
		log.Error("parse_turn2_failed", map[string]interface{}{"error": err.Error(), "verificationId": event.VerificationID})
		return models.LambdaOutput{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	item := models.VerificationResultItem{
		VerificationID:         event.VerificationID,
		VerificationAt:         now,
		VerificationType:       initData.VerificationType,
		LayoutID:               &initData.LayoutID,
		LayoutPrefix:           initData.LayoutPrefix,
		VendingMachineID:       initData.VendingMachineID,
		ReferenceImageUrl:      initData.ReferenceImageUrl,
		CheckingImageUrl:       initData.CheckingImageUrl,
		VerificationStatus:     parsed.VerificationStatus,
		CurrentStatus:          "COMPLETED",
		LastUpdatedAt:          now,
		ProcessingStartedAt:    initData.ProcessingStartedAt,
		InitialConfirmation:    parsed.InitialConfirmation,
		VerificationSummary:    parsed.VerificationSummary,
		PreviousVerificationID: initData.PreviousVerificationID,
	}
	if initData.LayoutID == 0 {
		item.LayoutID = nil
	}

	err = dynamodbhelper.StoreVerificationResult(ctx, awsClients.DynamoDBClient, appConfig.VerificationResultsTable, item)
	if err != nil {
		log.Error("dynamodb_store_failed", map[string]interface{}{"error": err.Error(), "verificationId": event.VerificationID})
		return models.LambdaOutput{}, err
	}

	err = dynamodbhelper.UpdateConversationHistory(ctx, awsClients.DynamoDBClient, appConfig.ConversationHistoryTable, event.VerificationID)
	if err != nil {
		log.Error("conversation_update_failed", map[string]interface{}{"error": err.Error(), "verificationId": event.VerificationID})
		return models.LambdaOutput{}, err
	}

	output := models.LambdaOutput{
		VerificationID:      event.VerificationID,
		VerificationAt:      now,
		Status:              "COMPLETED",
		VerificationStatus:  parsed.VerificationStatus,
		VerificationSummary: parsed.VerificationSummary,
		Message:             "Verification finalized and stored",
	}

	log.LogOutputEvent(output)
	return output, nil
}

func main() {
	lambda.Start(HandleRequest)
}
