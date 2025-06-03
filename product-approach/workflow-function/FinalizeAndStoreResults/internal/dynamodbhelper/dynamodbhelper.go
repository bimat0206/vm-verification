package dynamodbhelper

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"workflow-function/FinalizeAndStoreResults/internal/models"
	"workflow-function/shared/errors"
)

func StoreVerificationResult(ctx context.Context, client *dynamodb.Client, tableName string, item models.VerificationResultItem) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeDynamoDB, "failed to marshal verification result item", false)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tableName,
		Item:      av,
	})
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeDynamoDB, "failed to put item to DynamoDB", false)
	}
	return nil
}

func UpdateConversationHistory(ctx context.Context, client *dynamodb.Client, tableName, verificationID string) error {
	_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &tableName,
		Key: map[string]types.AttributeValue{
			"verificationId": &types.AttributeValueMemberS{Value: verificationID},
		},
		UpdateExpression: aws.String("SET turnStatus = :ts"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":ts": &types.AttributeValueMemberS{Value: "WORKFLOW_COMPLETED"},
		},
	})
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeDynamoDB, "failed to update conversation history", false)
	}
	return nil
}
