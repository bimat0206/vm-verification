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
	initData *schema.InitializationData,
) error {
	currentTimestamp := time.Now().UTC().Format(time.RFC3339)
	errorStatus := fmt.Sprintf("ERROR_%s", errorStage)

	errorDetails := schema.ErrorDetails{
		Type:    parsedErrorCause.ErrorType,
		Message: parsedErrorCause.ErrorMessage,
		Stage:   errorStage,
	}
	if len(parsedErrorCause.StackTrace) > 0 {
		errorDetails.StackTrace = strings.Join(parsedErrorCause.StackTrace, "\n")
	}

	errorTracking := schema.ErrorTracking{
		HasErrors:    true,
		CurrentError: &errorDetails,
		LastErrorAt:  currentTimestamp,
	}

	statusHistoryEntry := schema.StatusHistoryEntry{
		Status:       errorStatus,
		Timestamp:    currentTimestamp,
		ErrorStage:   errorStage,
		ErrorMessage: parsedErrorCause.ErrorMessage,
	}

	updateExpression := "SET #cs = :cs, #vs = :vs, #lua = :lua, #et = :et, #sh = list_append(if_not_exists(#sh, :empty), :entry)"
	exprNames := map[string]string{
		"#cs":  "currentStatus",
		"#vs":  "verificationStatus",
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
		":vs":    &types.AttributeValueMemberS{Value: "FAILED"},
		":lua":   &types.AttributeValueMemberS{Value: currentTimestamp},
		":et":    &types.AttributeValueMemberM{Value: marshaledTracking},
		":entry": &types.AttributeValueMemberL{Value: []types.AttributeValue{&types.AttributeValueMemberM{Value: marshaledEntry}}},
		":empty": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
	}

	_, err = ddbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &tableName,
		Key:                       map[string]types.AttributeValue{"verificationId": &types.AttributeValueMemberS{Value: verificationID}},
		UpdateExpression:          &updateExpression,
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})
	return err
}

// UpdateConversationHistoryOnError marks the conversation history record as failed.
func UpdateConversationHistoryOnError(ctx context.Context, ddbClient *dynamodb.Client, tableName, verificationID, errorSummary string) error {
	updateExpression := "SET #ts = :ts"
	exprNames := map[string]string{"#ts": "turnStatus"}
	exprValues := map[string]types.AttributeValue{":ts": &types.AttributeValueMemberS{Value: "FAILED_WORKFLOW"}}

	_, err := ddbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &tableName,
		Key:                       map[string]types.AttributeValue{"verificationId": &types.AttributeValueMemberS{Value: verificationID}},
		UpdateExpression:          &updateExpression,
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})
	return err
}
