// internal/services/s3.go
package services

import (
	"context"
	"fmt"

	"ExecuteTurn1Combined/internal/models"
	"workflow-function/shared/s3state"
)

// S3StateManager defines S3-based state persistence operations.
type S3StateManager interface {
	// LoadSystemPrompt retrieves the system prompt string from S3.
	LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error)
	// LoadBase64Image retrieves a Base64-encoded reference image from S3.
	LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error)
	// StoreRawResponse writes the raw Bedrock JSON response under "responses/turn1-raw-response".
	StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error)
	// StoreProcessedAnalysis writes the parsed/processed response under "responses/turn1-processed-response".
	StoreProcessedAnalysis(ctx context.Context, verificationID string, analysis interface{}) (models.S3Reference, error)
}

// s3Manager is the concrete implementation of S3StateManager, wrapping s3state.Manager.
type s3Manager struct {
	root s3state.Manager
}

// NewS3StateManager constructs an S3StateManager rooted at the given bucket.
func NewS3StateManager(bucket string) (S3StateManager, error) {
	mgr, err := s3state.New(bucket)
	if err != nil {
		return nil, err
	}
	return &s3Manager{root: mgr}, nil
}

// LoadSystemPrompt retrieves a string payload (JSON-unmarshaled) from the given S3 reference.
func (m *s3Manager) LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error) {
	var prompt string
	if err := m.root.RetrieveJSON(toStateRef(ref), &prompt); err != nil {
		return "", err
	}
	return prompt, nil
}

// LoadBase64Image retrieves a Base64 string payload (JSON-unmarshaled) from the given S3 reference.
func (m *s3Manager) LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error) {
	var img string
	if err := m.root.RetrieveJSON(toStateRef(ref), &img); err != nil {
		return "", err
	}
	return img, nil
}

// StoreRawResponse persists the raw Bedrock response under the "responses" category.
func (m *s3Manager) StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error) {
	key := fmt.Sprintf("%s/turn1-raw-response.json", verificationID)
	ref, err := m.root.StoreJSON("responses", key, data)
	if err != nil {
		return models.S3Reference{}, err
	}
	return fromStateRef(*ref), nil
}

// StoreProcessedAnalysis persists the processed analysis under the "responses" category.
func (m *s3Manager) StoreProcessedAnalysis(ctx context.Context, verificationID string, analysis interface{}) (models.S3Reference, error) {
	key := fmt.Sprintf("%s/turn1-processed-response.json", verificationID)
	ref, err := m.root.StoreJSON("responses", key, analysis)
	if err != nil {
		return models.S3Reference{}, err
	}
	return fromStateRef(*ref), nil
}

// toStateRef converts our model S3Reference into the s3state.Ref type.
func toStateRef(r models.S3Reference) s3state.Ref {
	return s3state.Ref{
		Bucket: r.Bucket,
		Key:    r.Key,
		ETag:   r.ETag,
		Size:   r.Size,
	}
}

// fromStateRef converts an s3state.Ref into our model S3Reference.
func fromStateRef(r s3state.Ref) models.S3Reference {
	return models.S3Reference{
		Bucket: r.Bucket,
		Key:    r.Key,
		ETag:   r.ETag,
		Size:   r.Size,
	}
}
