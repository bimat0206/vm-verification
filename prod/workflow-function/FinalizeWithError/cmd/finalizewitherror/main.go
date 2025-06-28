package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/FinalizeWithErrorFunction/internal/config"
	"workflow-function/FinalizeWithErrorFunction/internal/dynamodbhelper"
	"workflow-function/FinalizeWithErrorFunction/internal/models"
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

	stateManager, err = s3state.New(appConfig.StateBucket)
	if err != nil {
		log.Error("Failed to initialize S3 state manager", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

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

func inferErrorStage(errorMessage string) string {
	errorMessageLower := strings.ToLower(errorMessage)
	if strings.Contains(errorMessageLower, "turn2") {
		return "TURN2_PROCESSING"
	}
	if strings.Contains(errorMessageLower, "turn1") {
		return "TURN1_PROCESSING"
	}
	if strings.Contains(errorMessageLower, "initialization") || strings.Contains(errorMessageLower, "initialize") {
		return "INITIALIZATION"
	}
	if strings.Contains(errorMessageLower, "fetch") && strings.Contains(errorMessageLower, "image") {
		return "IMAGE_FETCH"
	}
	if strings.Contains(errorMessageLower, "prepare") && strings.Contains(errorMessageLower, "prompt") {
		return "PROMPT_PREPARATION"
	}
	if strings.Contains(errorMessageLower, "bedrock") {
		return "BEDROCK_PROCESSING"
	}
	return "UNKNOWN"
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

	errorStage := event.ErrorStage
	if errorStage == "" {
		errorStage = inferErrorStage(rawMessage)
		log.Info("Inferred error stage", map[string]interface{}{
			"originalStage": event.ErrorStage,
			"inferredStage": errorStage,
			"errorMessage":  rawMessage,
		})
	}

	var initData *models.InitializationData
	if ref := event.S3References.InitializationS3Ref; ref.Bucket != "" && ref.Key != "" {
		var data models.InitializationData
		if err := stateManager.RetrieveJSON(&s3state.Reference{
			Bucket: ref.Bucket,
			Key:    ref.Key,
			Size:   ref.Size,
		}, &data); err == nil {
			initData = &data
		} else {
			log.Warn("Failed to load initialization data", map[string]interface{}{"error": err.Error()})
		}
	}

	// Create envelope for error storage
	envelope := s3state.NewEnvelope(event.VerificationID)

	// Store error details using SaveToEnvelope for proper date-based path structure
	err := stateManager.SaveToEnvelope(envelope, "error", "error.json", stepCause)
	if err != nil {
		log.Error("Failed to store error details", map[string]interface{}{"error": err.Error()})
		return models.LambdaOutput{}, fmt.Errorf("failed to store error details: %w", err)
	}

	// Update initialization data with error information and save it
	var processingRef *s3state.Reference
	var hasProcessingRef bool
	if initData != nil {
		initData.Status = schema.VerificationStatusFailed
		initData.ErrorStage = errorStage
		initData.ErrorMessage = rawMessage
		initData.FailedAt = time.Now().UTC().Format(time.RFC3339)

		// Save updated initialization data
		err = stateManager.SaveToEnvelope(envelope, "processing", "initialization.json", initData)
		if err != nil {
			log.Error("Failed to update initialization data", map[string]interface{}{"error": err.Error()})
			return models.LambdaOutput{}, fmt.Errorf("failed to update initialization data: %w", err)
		}
		processingRef = envelope.GetReference("processing_initialization")
		if processingRef == nil {
			log.Error("Failed to get processing initialization reference from envelope", nil)
			return models.LambdaOutput{}, fmt.Errorf("failed to get processing initialization reference")
		}
		hasProcessingRef = true
	} else {
		log.Warn("No initialization data available to update - proceeding with error-only output", nil)
		if ref := event.S3References.InitializationS3Ref; ref.Bucket != "" && ref.Key != "" {
			processingRef = &s3state.Reference{
				Bucket: ref.Bucket,
				Key:    ref.Key,
				Size:   ref.Size,
			}
		}
	}

	// Get references from the envelope
	errorRef := envelope.GetReference("error_error")
	if errorRef == nil {
		log.Error("Failed to get error reference from envelope", nil)
		return models.LambdaOutput{}, fmt.Errorf("failed to get error reference")
	}

	// Update databases
	if err := dynamodbhelper.UpdateVerificationResultOnError(ctx, awsClients.DynamoDBClient,
		appConfig.VerificationResultsTable, event.VerificationID, errorStage, stepCause, initData); err != nil {
		return models.LambdaOutput{}, fmt.Errorf("verification update failed: %w", err)
	}

	if appConfig.ConversationHistoryTable != "" {
		if err := dynamodbhelper.UpdateConversationHistoryOnError(ctx, awsClients.DynamoDBClient,
			appConfig.ConversationHistoryTable, event.VerificationID, rawMessage, initData); err != nil {
			log.Warn("Conversation history update failed", map[string]interface{}{"error": err.Error()})
		}
	}

	// Build output with conditional processing reference
	output := models.LambdaOutput{
		VerificationID: event.VerificationID,
		Status:         schema.VerificationStatusFailed,
		ErrorStage:     errorStage,
		ErrorMessage:   rawMessage,
		Message:        fmt.Sprintf("Verification failed at %s stage. Error details logged and persisted.", errorStage),
	}

	// Set S3 references based on availability
	if hasProcessingRef && processingRef != nil {
		output.S3References = struct {
			ProcessingInitialization models.S3Reference      `json:"processing_initialization"`
			Error                    models.S3ErrorReference `json:"error"`
		}{
			ProcessingInitialization: models.S3Reference{
				Bucket: processingRef.Bucket,
				Key:    processingRef.Key,
				Size:   processingRef.Size,
			},
			Error: models.S3ErrorReference{
				Bucket: errorRef.Bucket,
				Key:    errorRef.Key,
				Size:   errorRef.Size,
			},
		}
	} else {
		// Only error reference available
		output.S3References = struct {
			ProcessingInitialization models.S3Reference      `json:"processing_initialization"`
			Error                    models.S3ErrorReference `json:"error"`
		}{
			ProcessingInitialization: models.S3Reference{}, // Empty reference
			Error: models.S3ErrorReference{
				Bucket: errorRef.Bucket,
				Key:    errorRef.Key,
				Size:   errorRef.Size,
			},
		}
	}
	log.LogOutputEvent(output)
	return output, nil
}

func main() {
	lambda.Start(HandleRequest)
}
