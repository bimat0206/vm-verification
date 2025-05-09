package main

import (
    "context"
    "fmt"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// LayoutMetadata holds layout information from DynamoDB.
type LayoutMetadata struct {
    LayoutId      int64  `json:"layoutId"`
    LayoutPrefix  string `json:"layoutPrefix"`
    // Add more fields as needed
    MachineStructure map[string]interface{} `json:"machineStructure"`
    ProductMapping   map[string]interface{} `json:"productMapping"`
}

// HistoricalContext holds previous verification summary.
type HistoricalContext struct {
    PreviousVerificationId string `json:"previousVerificationId"`
    Summary                map[string]interface{} `json:"summary"`
    // Add more fields as needed
}

// FetchLayoutMetadata retrieves layout metadata from DynamoDB.
func FetchLayoutMetadata(ctx context.Context, layoutId int64, layoutPrefix string) (*LayoutMetadata, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    client := dynamodb.NewFromConfig(cfg)

    tableName := "LayoutMetadata" // Replace with your table name

    out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "layoutId":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", layoutId)},
            "layoutPrefix": &types.AttributeValueMemberS{Value: layoutPrefix},
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to fetch layout metadata: %w", err)
    }
    if out.Item == nil {
        return nil, fmt.Errorf("layout metadata not found")
    }

    // Example: Unmarshal to struct (expand as needed)
    meta := &LayoutMetadata{
        LayoutId:     layoutId,
        LayoutPrefix: layoutPrefix,
        // Fill in more fields as needed
    }
    // TODO: Unmarshal machineStructure, productMapping, etc.
    return meta, nil
}

// FetchHistoricalContext retrieves previous verification from DynamoDB.
func FetchHistoricalContext(ctx context.Context, verificationId string) (*HistoricalContext, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    client := dynamodb.NewFromConfig(cfg)

    tableName := "VerificationResults" // Replace with your table name

    out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "verificationId": &types.AttributeValueMemberS{Value: verificationId},
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to fetch historical context: %w", err)
    }
    if out.Item == nil {
        return nil, fmt.Errorf("historical verification not found")
    }

    ctxObj := &HistoricalContext{
        PreviousVerificationId: verificationId,
        // TODO: Unmarshal summary, etc.
    }
    return ctxObj, nil
}

// Optionally, update verification status in DynamoDB.
func UpdateVerificationStatus(ctx context.Context, verificationId, status string) error {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return fmt.Errorf("failed to load AWS config: %w", err)
    }
    client := dynamodb.NewFromConfig(cfg)

    tableName := "VerificationResults" // Replace with your table name

    _, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "verificationId": &types.AttributeValueMemberS{Value: verificationId},
        },
        UpdateExpression: aws.String("SET verificationStatus = :status, updatedAt = :now"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":status": &types.AttributeValueMemberS{Value: status},
            ":now":    &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
        },
    })
    if err != nil {
        return fmt.Errorf("failed to update verification status: %w", err)
    }
    return nil
}
