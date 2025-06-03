package dynamodbhelper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"workflow-function/FinalizeWithErrorFunction/internal/models"
	"workflow-function/shared/schema"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// UpdateVerificationResultOnError updates the VerificationResults table with error status and tracking info.
func UpdateVerificationResultOnError(
	ctx context.Context,
	ddbClient *dynamodb.Client,
	tableName string,
	verificationID string,
	errorStage string,
	parsedErrorCause *models.StepFunctionsErrorCause,
	initData *models.InitializationData,
) error {
	// Extract verificationAt from initData if available
	var verificationAt string
	if initData != nil && initData.VerificationContext.VerificationAt != "" {
		verificationAt = initData.VerificationContext.VerificationAt
	} else {
		// If we don't have initData or verificationAt, try to extract from verificationID
		// VerificationID format: verif-YYYYMMDDHHMMSS-XXXX
		if len(verificationID) >= 20 && strings.HasPrefix(verificationID, "verif-") {
			// Extract timestamp part: verif-20250603103642-f1f2 -> 20250603103642
			timestampPart := verificationID[6:20] // Skip "verif-" prefix
			if len(timestampPart) == 14 {
				// Parse YYYYMMDDHHMMSS format
				if parsedTime, err := time.Parse("20060102150405", timestampPart); err == nil {
					verificationAt = parsedTime.UTC().Format(time.RFC3339)
				}
			}
		}

		// If still empty, use current timestamp as fallback
		if verificationAt == "" {
			verificationAt = time.Now().UTC().Format(time.RFC3339)
		}
	}
	currentTimestamp := time.Now().UTC().Format(time.RFC3339)
	errorStatus := fmt.Sprintf("ERROR_%s", errorStage)

	errorDetails := schema.ErrorInfo{
		Code:    parsedErrorCause.ErrorType,
		Message: parsedErrorCause.ErrorMessage,
		Details: map[string]interface{}{
			"stage": errorStage,
		},
		Timestamp: currentTimestamp,
	}
	if len(parsedErrorCause.StackTrace) > 0 {
		errorDetails.Details["stackTrace"] = strings.Join(parsedErrorCause.StackTrace, "\n")
	}

	errorTracking := schema.ErrorTracking{
		HasErrors:    true,
		CurrentError: &errorDetails,
		LastErrorAt:  currentTimestamp,
	}

	statusHistoryEntry := schema.StatusHistoryEntry{
		Status:           errorStatus,
		Timestamp:        currentTimestamp,
		FunctionName:     "FinalizeWithError",
		ProcessingTimeMs: 0,
		Stage:            errorStage,
		Metrics: map[string]interface{}{
			"errorMessage": parsedErrorCause.ErrorMessage,
			"errorType":    parsedErrorCause.ErrorType,
		},
	}

	updateExpression := "SET #cs = :cs, #vs = :vs, #st = :st, #lua = :lua, #et = :et, #sh = list_append(if_not_exists(#sh, :empty), :entry)"
	exprNames := map[string]string{
		"#cs":  "currentStatus",
		"#vs":  "verificationStatus",
		"#st":  "status",
		"#lua": "lastUpdatedAt",
		"#et":  "errorTracking",
		"#sh":  "statusHistory",
	}
	marshaledTracking, err := attributevalue.MarshalMap(errorTracking)
	if err != nil {
		return fmt.Errorf("marshal errorTracking: %w", err)
	}
	marshaledEntry, err := attributevalue.MarshalMap(statusHistoryEntry)
	if err != nil {
		return fmt.Errorf("marshal statusHistory: %w", err)
	}

	exprValues := map[string]types.AttributeValue{
		":cs":    &types.AttributeValueMemberS{Value: errorStatus},
		":vs":    &types.AttributeValueMemberS{Value: schema.VerificationStatusFailed},
		":st":    &types.AttributeValueMemberS{Value: schema.StatusVerificationFailed},
		":lua":   &types.AttributeValueMemberS{Value: currentTimestamp},
		":et":    &types.AttributeValueMemberM{Value: marshaledTracking},
		":entry": &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberM{Value: marshaledEntry}}},
		":empty": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
	}

	_, err = ddbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &tableName,
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationID},
			"verificationAt": &types.AttributeValueMemberS{Value: verificationAt},
		},
		UpdateExpression:          &updateExpression,
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})
	return err
}

// UpdateConversationHistoryOnError marks the conversation history record as failed.
func UpdateConversationHistoryOnError(ctx context.Context, ddbClient *dynamodb.Client, tableName, verificationID, errorSummary string, initData *models.InitializationData) error {
	// Extract conversationAt from initData if available
	var conversationAt string
	if initData != nil && initData.VerificationContext.VerificationAt != "" {
		// For conversation history, conversationAt is typically the same as verificationAt
		conversationAt = initData.VerificationContext.VerificationAt
	} else {
		// If we don't have initData or conversationAt, try to extract from verificationID
		// VerificationID format: verif-YYYYMMDDHHMMSS-XXXX
		if len(verificationID) >= 20 && strings.HasPrefix(verificationID, "verif-") {
			// Extract timestamp part: verif-20250603103642-f1f2 -> 20250603103642
			timestampPart := verificationID[6:20] // Skip "verif-" prefix
			if len(timestampPart) == 14 {
				// Parse YYYYMMDDHHMMSS format
				if parsedTime, err := time.Parse("20060102150405", timestampPart); err == nil {
					conversationAt = parsedTime.UTC().Format(time.RFC3339)
				}
			}
		}

		// If still empty, use current timestamp as fallback
		if conversationAt == "" {
			conversationAt = time.Now().UTC().Format(time.RFC3339)
		}
	}
	updateExpression := "SET #ts = :ts"
	exprNames := map[string]string{"#ts": "turnStatus"}
	exprValues := map[string]types.AttributeValue{":ts": &types.AttributeValueMemberS{Value: "FAILED_WORKFLOW"}}

	_, err := ddbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &tableName,
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationID},
			"conversationAt": &types.AttributeValueMemberS{Value: conversationAt},
		},
		UpdateExpression:          &updateExpression,
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})
	return err
}
