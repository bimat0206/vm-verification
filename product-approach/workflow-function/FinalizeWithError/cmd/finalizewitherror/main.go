package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/FinalizeWithErrorFunction/internal/config"
	"workflow-function/FinalizeWithErrorFunction/internal/dynamodbhelper"
	"workflow-function/FinalizeWithErrorFunction/internal/models"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

var (
	appConfig  *config.LambdaConfig
	awsClients *config.AWSClients
	log        logger.Logger
)

func init() {
	log = logger.New("kootoro-verification", "FinalizeWithError")
	var err error
	appConfig, err = config.LoadEnvConfig()
	if err != nil {
		log.Error("Failed to load configuration", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	awsClients, err = config.NewAWSClients(context.Background())
	if err != nil {
		log.Error("Failed to initialize AWS clients", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

// parseStepFunctionsErrorCause parses the Step Functions cause string into a struct.
func parseStepFunctionsErrorCause(causeString string) (*models.StepFunctionsErrorCause, error) {
	var cause models.StepFunctionsErrorCause
	if err := json.Unmarshal([]byte(causeString), &cause); err == nil {
		if cause.ErrorMessage == "" {
			cause.ErrorMessage = causeString
		}
		return &cause, nil
	}
	return &models.StepFunctionsErrorCause{ErrorMessage: causeString}, nil
}

func HandleRequest(ctx context.Context, event models.LambdaInput) (models.LambdaOutput, error) {
	log.LogReceivedEvent(event)

	var stepCause *models.StepFunctionsErrorCause
	rawMessage := "Unknown error"

	if cause, ok := event.Error["Cause"].(string); ok {
		parsed, _ := parseStepFunctionsErrorCause(cause)
		stepCause = parsed
		rawMessage = parsed.ErrorMessage
	} else if msg, ok := event.Error["ErrorMessage"].(string); ok {
		stepCause = &models.StepFunctionsErrorCause{ErrorMessage: msg}
		rawMessage = msg
	} else {
		stepCause = &models.StepFunctionsErrorCause{ErrorMessage: "Unknown error"}
	}

	var initData *schema.InitializationData
	if ref := event.PartialS3References.InitializationS3Ref; ref.Bucket != "" && ref.Key != "" {
		var data schema.InitializationData
		err := s3state.GetS3ObjectAsJSON(ctx, awsClients.S3Client, ref.Bucket, ref.Key, &data)
		if err != nil {
			log.Warn("failed to load initialization.json", map[string]interface{}{"error": err.Error(), "bucket": ref.Bucket, "key": ref.Key})
		} else {
			initData = &data
		}
	}

	err := dynamodbhelper.UpdateVerificationResultOnError(ctx, awsClients.DynamoDBClient,
		appConfig.VerificationResultsTable, event.VerificationID, event.ErrorStage, stepCause, initData)
	if err != nil {
		log.Error("failed to update verification result", map[string]interface{}{"error": err.Error(), "verificationId": event.VerificationID})
		return models.LambdaOutput{}, fmt.Errorf("critical update failure: %w", err)
	}

	if appConfig.ConversationHistoryTable != "" {
		if err := dynamodbhelper.UpdateConversationHistoryOnError(ctx, awsClients.DynamoDBClient,
			appConfig.ConversationHistoryTable, event.VerificationID, rawMessage); err != nil {
			log.Warn("failed to update conversation history", map[string]interface{}{"error": err.Error(), "verificationId": event.VerificationID})
		}
	}

	status := fmt.Sprintf("ERROR_%s", strings.ToUpper(event.ErrorStage))
	if len(status) > 256 {
		status = "FAILED"
	}

	output := models.LambdaOutput{
		VerificationID: event.VerificationID,
		Status:         status,
		ErrorStage:     event.ErrorStage,
		ErrorMessage:   rawMessage,
		Message:        fmt.Sprintf("Verification failed at stage '%s'", event.ErrorStage),
	}
	log.LogOutputEvent(output)
	return output, nil
}

func main() {
	lambda.Start(HandleRequest)
}
