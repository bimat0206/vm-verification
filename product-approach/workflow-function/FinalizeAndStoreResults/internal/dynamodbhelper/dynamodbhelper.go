package dynamodbhelper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"workflow-function/FinalizeAndStoreResults/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// createEnhancedDynamoDBError creates a detailed DynamoDB error with AWS-specific information
func createEnhancedDynamoDBError(operation string, tableName string, awsErr error, itemData interface{}) *errors.WorkflowError {
	details := map[string]interface{}{
		"operation": operation,
		"table":     tableName,
		"awsError":  awsErr.Error(),
	}

	// Add sanitized item data for debugging (remove sensitive fields)
	if itemData != nil {
		if sanitized := sanitizeItemForLogging(itemData); sanitized != nil {
			details["itemStructure"] = sanitized
		}
	}

	// Extract specific AWS error information
	errorCode := "DYNAMODB_ERROR"
	message := fmt.Sprintf("DynamoDB %s operation failed", operation)
	retryable := false
	severity := errors.ErrorSeverityMedium

	errorStr := awsErr.Error()

	if strings.Contains(errorStr, "ValidationException") {
		errorCode = "VALIDATION_EXCEPTION"
		message = fmt.Sprintf("DynamoDB validation error during %s", operation)
		severity = errors.ErrorSeverityHigh
		details["errorType"] = "ValidationException"
		details["troubleshooting"] = "Check item structure, required fields, and data types"
	} else if strings.Contains(errorStr, "ConditionalCheckFailedException") {
		errorCode = "CONDITIONAL_CHECK_FAILED"
		message = fmt.Sprintf("DynamoDB conditional check failed during %s", operation)
		details["errorType"] = "ConditionalCheckFailedException"
		details["troubleshooting"] = "Item may already exist or condition expression failed"
	} else if strings.Contains(errorStr, "ProvisionedThroughputExceededException") {
		errorCode = "THROUGHPUT_EXCEEDED"
		message = fmt.Sprintf("DynamoDB throughput exceeded during %s", operation)
		severity = errors.ErrorSeverityLow
		retryable = true
		details["errorType"] = "ProvisionedThroughputExceededException"
		details["troubleshooting"] = "Consider implementing exponential backoff retry"
	} else if strings.Contains(errorStr, "ResourceNotFoundException") {
		errorCode = "RESOURCE_NOT_FOUND"
		message = fmt.Sprintf("DynamoDB table not found during %s", operation)
		severity = errors.ErrorSeverityHigh
		details["errorType"] = "ResourceNotFoundException"
		details["troubleshooting"] = "Verify table name and AWS region configuration"
	} else if strings.Contains(errorStr, "InternalServerError") {
		errorCode = "INTERNAL_SERVER_ERROR"
		message = fmt.Sprintf("DynamoDB internal server error during %s", operation)
		retryable = true
		details["errorType"] = "InternalServerError"
		details["troubleshooting"] = "AWS service issue - retry may resolve"
	} else if strings.Contains(errorStr, "ServiceUnavailable") {
		errorCode = "SERVICE_UNAVAILABLE"
		message = fmt.Sprintf("DynamoDB service unavailable during %s", operation)
		severity = errors.ErrorSeverityLow
		retryable = true
		details["errorType"] = "ServiceUnavailable"
		details["troubleshooting"] = "AWS service temporarily unavailable - retry recommended"
	} else if strings.Contains(errorStr, "ThrottlingException") {
		errorCode = "THROTTLING_EXCEPTION"
		message = fmt.Sprintf("DynamoDB throttling during %s", operation)
		severity = errors.ErrorSeverityLow
		retryable = true
		details["errorType"] = "ThrottlingException"
		details["troubleshooting"] = "Implement exponential backoff retry strategy"
	}

	wfErr := &errors.WorkflowError{
		Type:      errors.ErrorTypeDynamoDB,
		Message:   message,
		Code:      errorCode,
		Details:   details,
		Retryable: retryable,
		Timestamp: time.Now(),
		Severity:  severity,
		APISource: errors.APISourceUnknown,
	}

	return wfErr
}

// sanitizeItemForLogging removes sensitive data and creates a structure summary for logging
func sanitizeItemForLogging(itemData interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Handle VerificationResultItem specifically
	if item, ok := itemData.(models.VerificationResultItem); ok {
		// Include non-sensitive fields directly
		result["verificationId"] = item.VerificationID
		result["verificationAt"] = item.VerificationAt
		result["verificationType"] = item.VerificationType
		result["currentStatus"] = item.CurrentStatus
		result["verificationStatus"] = item.VerificationStatus

		// Sanitize URLs
		if item.ReferenceImageUrl != "" {
			result["referenceImageUrl"] = "[URL_PROVIDED]"
		} else {
			result["referenceImageUrl"] = "[EMPTY]"
		}

		if item.CheckingImageUrl != "" {
			result["checkingImageUrl"] = "[URL_PROVIDED]"
		} else {
			result["checkingImageUrl"] = "[EMPTY]"
		}

		// Sanitize other fields by showing type/status
		if item.VendingMachineID != "" {
			result["vendingMachineId"] = "[PROVIDED]"
		} else {
			result["vendingMachineId"] = "[EMPTY]"
		}

		if item.LayoutID != nil {
			result["layoutId"] = fmt.Sprintf("[%d]", *item.LayoutID)
		} else {
			result["layoutId"] = "[NULL]"
		}

		if item.LayoutPrefix != "" {
			result["layoutPrefix"] = "[PROVIDED]"
		} else {
			result["layoutPrefix"] = "[EMPTY]"
		}

		// Add summary info
		result["hasStatusHistory"] = len(item.StatusHistory) > 0
		result["hasProcessingMetrics"] = len(item.ProcessingMetrics) > 0
		result["hasErrorTracking"] = len(item.ErrorTracking) > 0

		return result
	}

	// Fallback for other types - convert to JSON and back
	jsonBytes, err := json.Marshal(itemData)
	if err != nil {
		return map[string]interface{}{"error": "failed to serialize item for logging"}
	}

	var itemMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &itemMap); err != nil {
		return map[string]interface{}{"error": "failed to deserialize item for logging"}
	}

	// Create a summary with field names and types, but not actual values for sensitive fields
	for key, value := range itemMap {
		switch key {
		case "verificationId", "verificationAt", "verificationType", "currentStatus", "verificationStatus":
			// Include these non-sensitive fields
			result[key] = value
		case "referenceImageUrl", "checkingImageUrl":
			// Sanitize URLs - show only if they exist
			if value != nil && value != "" {
				result[key] = "[URL_PROVIDED]"
			} else {
				result[key] = "[EMPTY]"
			}
		default:
			// For other fields, just show the type and whether it's populated
			if value == nil {
				result[key] = "[NULL]"
			} else if str, ok := value.(string); ok && str == "" {
				result[key] = "[EMPTY_STRING]"
			} else {
				result[key] = fmt.Sprintf("[%T]", value)
			}
		}
	}

	return result
}

// validateVerificationResultItem performs basic validation on the item before storing
func validateVerificationResultItem(item models.VerificationResultItem) error {
	if item.VerificationID == "" {
		return errors.NewValidationError("verificationId is required", map[string]interface{}{
			"field": "verificationId",
		})
	}

	if item.VerificationAt == "" {
		return errors.NewValidationError("verificationAt is required", map[string]interface{}{
			"field": "verificationAt",
		})
	}

	if item.VerificationType == "" {
		return errors.NewValidationError("verificationType is required", map[string]interface{}{
			"field": "verificationType",
		})
	}

	if item.CurrentStatus == "" {
		return errors.NewValidationError("currentStatus is required", map[string]interface{}{
			"field": "currentStatus",
		})
	}

	// Validate verificationStatus - required for VerificationStatusIndex
	if item.VerificationStatus == "" {
		return errors.NewValidationError("verificationStatus is required for secondary index", map[string]interface{}{
			"field": "verificationStatus",
			"reason": "DynamoDB VerificationStatusIndex does not allow empty string values",
		})
	}

	return nil
}

func StoreVerificationResult(ctx context.Context, client *dynamodb.Client, tableName string, item models.VerificationResultItem, log logger.Logger) error {
	// Validate the item before attempting to store it
	if err := validateVerificationResultItem(item); err != nil {
		log.Error("validation_failed", map[string]interface{}{
			"error":          err.Error(),
			"verificationId": item.VerificationID,
			"table":          tableName,
		})
		return err
	}

	// Log the operation attempt with sanitized data
	log.Info("updating_verification_result", map[string]interface{}{
		"verificationId": item.VerificationID,
		"verificationAt": item.VerificationAt,
		"table":          tableName,
		"operation":      "UpdateItem",
		"itemSummary":    sanitizeItemForLogging(item),
	})

	// Marshal the VerificationSummary struct
	vsmAV, err := attributevalue.Marshal(item.VerificationSummary)
	if err != nil {
		enhancedErr := createEnhancedDynamoDBError("marshal_verification_summary", tableName, err, item)
		enhancedErr.VerificationID = item.VerificationID
		log.Error("marshal_verification_summary_failed", map[string]interface{}{
			"error":          enhancedErr.Error(),
			"verificationId": item.VerificationID,
			"table":          tableName,
			"details":        enhancedErr.Details,
		})
		return enhancedErr
	}

	// Use UpdateItem to update existing record instead of creating new one
	updateInput := &dynamodb.UpdateItemInput{
		TableName: &tableName,
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: item.VerificationID},
			"verificationAt": &types.AttributeValueMemberS{Value: item.VerificationAt},
		},
		UpdateExpression: aws.String(`SET
			verificationType = :vt,
			layoutId = :lid,
			layoutPrefix = :lp,
			vendingMachineId = :vmid,
			referenceImageUrl = :riu,
			checkingImageUrl = :ciu,
			verificationStatus = :vs,
			currentStatus = :cs,
			lastUpdatedAt = :lua,
			processingStartedAt = :psa,
			initialConfirmation = :ic,
			verificationSummary = :vsm,
			previousVerificationId = :pvid,
			statusHistory = if_not_exists(statusHistory, :empty_list),
			processingMetrics = if_not_exists(processingMetrics, :empty_map),
			errorTracking = if_not_exists(errorTracking, :empty_map)`),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":vt":         &types.AttributeValueMemberS{Value: item.VerificationType},
			":lp":         &types.AttributeValueMemberS{Value: item.LayoutPrefix},
			":vmid":       &types.AttributeValueMemberS{Value: item.VendingMachineID},
			":riu":        &types.AttributeValueMemberS{Value: item.ReferenceImageUrl},
			":ciu":        &types.AttributeValueMemberS{Value: item.CheckingImageUrl},
			":vs":         &types.AttributeValueMemberS{Value: item.VerificationStatus},
			":cs":         &types.AttributeValueMemberS{Value: item.CurrentStatus},
			":lua":        &types.AttributeValueMemberS{Value: item.LastUpdatedAt},
			":psa":        &types.AttributeValueMemberS{Value: item.ProcessingStartedAt},
			":ic":         &types.AttributeValueMemberS{Value: item.InitialConfirmation},
			":vsm":        vsmAV,
			":pvid":       &types.AttributeValueMemberS{Value: item.PreviousVerificationID},
			":empty_list": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
			":empty_map":  &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
		},
	}

	// Handle optional layoutId
	if item.LayoutID != nil {
		updateInput.ExpressionAttributeValues[":lid"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", *item.LayoutID)}
	} else {
		updateInput.ExpressionAttributeValues[":lid"] = &types.AttributeValueMemberNULL{Value: true}
	}

	_, err = client.UpdateItem(ctx, updateInput)
	if err != nil {
		enhancedErr := createEnhancedDynamoDBError("UpdateItem", tableName, err, item)
		enhancedErr.VerificationID = item.VerificationID
		log.Error("dynamodb_update_failed", map[string]interface{}{
			"error":          enhancedErr.Error(),
			"verificationId": item.VerificationID,
			"verificationAt": item.VerificationAt,
			"table":          tableName,
			"awsErrorCode":   enhancedErr.Code,
			"retryable":      enhancedErr.Retryable,
			"details":        enhancedErr.Details,
		})
		return enhancedErr
	}

	log.Info("verification_result_updated", map[string]interface{}{
		"verificationId": item.VerificationID,
		"verificationAt": item.VerificationAt,
		"table":          tableName,
		"status":         item.CurrentStatus,
	})

	return nil
}

func UpdateConversationHistory(ctx context.Context, client *dynamodb.Client, tableName, verificationID, expectedConversationAt string, log logger.Logger) error {
	if verificationID == "" {
		validationErr := errors.NewValidationError("verificationID required", map[string]interface{}{
			"field": "verificationID",
		})
		log.Error("conversation_update_validation_failed", map[string]interface{}{
			"error":          validationErr.Error(),
			"verificationId": verificationID,
			"table":          tableName,
		})
		return validationErr
	}

	log.Info("finding_conversation_history", map[string]interface{}{
		"verificationId":           verificationID,
		"expectedConversationAt":   expectedConversationAt,
		"table":                    tableName,
		"operation":                "Query",
	})

	// Query to find the actual conversation record for this verificationID
	queryInput := &dynamodb.QueryInput{
		TableName:              &tableName,
		KeyConditionExpression: aws.String("verificationId = :vid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":vid": &types.AttributeValueMemberS{Value: verificationID},
		},
		ScanIndexForward: aws.Bool(false), // Get most recent first
		Limit:            aws.Int32(1),
		ProjectionExpression: aws.String("verificationId, conversationAt, turnStatus"),
	}

	queryResult, err := client.Query(ctx, queryInput)
	if err != nil {
		enhancedErr := createEnhancedDynamoDBError("Query", tableName, err, map[string]interface{}{
			"verificationId": verificationID,
			"operation": "FindConversationRecord",
		})
		enhancedErr.VerificationID = verificationID
		log.Error("conversation_query_failed", map[string]interface{}{
			"error":          enhancedErr.Error(),
			"verificationId": verificationID,
			"table":          tableName,
			"awsErrorCode":   enhancedErr.Code,
			"details":        enhancedErr.Details,
		})
		return enhancedErr
	}

	if len(queryResult.Items) == 0 {
		validationErr := errors.NewValidationError("no conversation record found", map[string]interface{}{
			"verificationId": verificationID,
			"table": tableName,
		})
		log.Error("conversation_record_not_found", map[string]interface{}{
			"error":          validationErr.Error(),
			"verificationId": verificationID,
			"table":          tableName,
		})
		return validationErr
	}

	// Extract the actual conversationAt from the found record
	conversationAtAttr, exists := queryResult.Items[0]["conversationAt"]
	if !exists {
		validationErr := errors.NewValidationError("conversationAt missing from record", map[string]interface{}{
			"verificationId": verificationID,
			"table": tableName,
		})
		log.Error("conversation_at_missing", map[string]interface{}{
			"error":          validationErr.Error(),
			"verificationId": verificationID,
			"table":          tableName,
		})
		return validationErr
	}

	actualConversationAt := conversationAtAttr.(*types.AttributeValueMemberS).Value

	log.Info("updating_conversation_history", map[string]interface{}{
		"verificationId":           verificationID,
		"expectedConversationAt":   expectedConversationAt,
		"actualConversationAt":     actualConversationAt,
		"table":                    tableName,
		"operation":                "UpdateItem",
		"newStatus":                "WORKFLOW_COMPLETED",
	})

	updateInput := &dynamodb.UpdateItemInput{
		TableName: &tableName,
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationID},
			"conversationAt": &types.AttributeValueMemberS{Value: actualConversationAt},
		},
		UpdateExpression: aws.String("SET turnStatus = :ts"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":ts": &types.AttributeValueMemberS{Value: "WORKFLOW_COMPLETED"},
		},
	}

	_, err = client.UpdateItem(ctx, updateInput)
	if err != nil {
		enhancedErr := createEnhancedDynamoDBError("UpdateItem", tableName, err, map[string]interface{}{
			"verificationId": verificationID,
			"conversationAt": actualConversationAt,
			"updateExpression": "SET turnStatus = :ts",
			"newStatus": "WORKFLOW_COMPLETED",
		})
		enhancedErr.VerificationID = verificationID
		log.Error("conversation_update_failed", map[string]interface{}{
			"error":          enhancedErr.Error(),
			"verificationId": verificationID,
			"actualConversationAt": actualConversationAt,
			"table":          tableName,
			"awsErrorCode":   enhancedErr.Code,
			"retryable":      enhancedErr.Retryable,
			"details":        enhancedErr.Details,
		})
		return enhancedErr
	}

	log.Info("conversation_history_updated", map[string]interface{}{
		"verificationId": verificationID,
		"actualConversationAt": actualConversationAt,
		"table":          tableName,
		"newStatus":      "WORKFLOW_COMPLETED",
	})

	return nil
}
