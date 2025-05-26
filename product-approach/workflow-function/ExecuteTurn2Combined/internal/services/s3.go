package services

import (
    "context"

    "workflow-function/ExecuteTurn2Combined/internal/models"
)

type S3StateManager interface {
    LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error)
    LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error)
    LoadJSON(ctx context.Context, ref models.S3Reference, target interface{}) error
    StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error)
}

// NewS3StateManager returns a no-op manager for this skeleton implementation
func NewS3StateManager(cfg interface{}, log interface{}) (S3StateManager, error) {
    return &noopS3Manager{}, nil
}

type noopS3Manager struct{}

func (n *noopS3Manager) LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error) { return "", nil }
func (n *noopS3Manager) LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error) { return "", nil }
func (n *noopS3Manager) LoadJSON(ctx context.Context, ref models.S3Reference, target interface{}) error { return nil }
func (n *noopS3Manager) StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error) {
    return models.S3Reference{}, nil
}
