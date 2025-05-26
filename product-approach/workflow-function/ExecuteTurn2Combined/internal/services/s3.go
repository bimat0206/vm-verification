package services

import (
	"context"
	"fmt"

	"workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
)

// S3StateManager defines minimal S3 operations required for Turn 2.
type S3StateManager interface {
	LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error)
	LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error)
	LoadJSON(ctx context.Context, ref models.S3Reference, target interface{}) error
	StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error)
}

type s3Manager struct {
	mgr    s3state.Manager
	bucket string
	cfg    config.Config
	log    logger.Logger
}

// NewS3StateManager creates an S3 manager using the shared state package.
func NewS3StateManager(cfg config.Config, log logger.Logger) (S3StateManager, error) {
	mgr, err := s3state.New(cfg.AWS.S3Bucket)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeS3, "init s3 manager", false)
	}
	return &s3Manager{mgr: mgr, bucket: cfg.AWS.S3Bucket, cfg: cfg, log: log}, nil
}

func (s *s3Manager) refToState(ref models.S3Reference) *s3state.Reference {
	return &s3state.Reference{Bucket: ref.Bucket, Key: ref.Key}
}

func (s *s3Manager) toModelRef(ref *s3state.Reference) models.S3Reference {
	if ref == nil {
		return models.S3Reference{}
	}
	return models.S3Reference{Bucket: ref.Bucket, Key: ref.Key, Size: ref.Size}
}

func (s *s3Manager) LoadSystemPrompt(ctx context.Context, ref models.S3Reference) (string, error) {
	var wrapper struct {
		PromptContent struct {
			SystemMessage string `json:"systemMessage"`
		} `json:"promptContent"`
	}
	if err := s.mgr.RetrieveJSON(s.refToState(ref), &wrapper); err != nil {
		return "", err
	}
	return wrapper.PromptContent.SystemMessage, nil
}

func (s *s3Manager) LoadBase64Image(ctx context.Context, ref models.S3Reference) (string, error) {
	data, err := s.mgr.Retrieve(s.refToState(ref))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *s3Manager) LoadJSON(ctx context.Context, ref models.S3Reference, target interface{}) error {
	return s.mgr.RetrieveJSON(s.refToState(ref), target)
}

func (s *s3Manager) StoreRawResponse(ctx context.Context, verificationID string, data interface{}) (models.S3Reference, error) {
	key := fmt.Sprintf("%s/%s/responses/turn2-raw-response.json", s.cfg.CurrentDatePartition(), verificationID)
	ref, err := s.mgr.StoreJSON("", key, data)
	if err != nil {
		return models.S3Reference{}, err
	}
	return s.toModelRef(ref), nil
}
