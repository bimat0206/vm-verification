package services

import "context"

// DynamoDBService abstracts DynamoDB operations

type DynamoDBService interface {
    UpdateTurn2Completion(ctx context.Context, verificationID string) error
}

func NewDynamoDBService(cfg interface{}) (DynamoDBService, error) {
    return &noopDynamo{}, nil
}

type noopDynamo struct{}

func (n *noopDynamo) UpdateTurn2Completion(ctx context.Context, verificationID string) error { return nil }
