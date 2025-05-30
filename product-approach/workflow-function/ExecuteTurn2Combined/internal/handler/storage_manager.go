package handler

import (
	"context"
	"encoding/json"

	"workflow-function/ExecuteTurn2Combined/internal/models"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
)

// StorageManager handles saving Turn2 artifacts to S3 using an envelope
// and exposes the resulting references in models.S3Reference format.
type StorageManager struct {
	manager s3state.Manager
	log     logger.Logger
}

func NewStorageManager(manager s3state.Manager, log logger.Logger) *StorageManager {
	return &StorageManager{manager: manager, log: log}
}

// SaveTurn2Outputs stores the raw Bedrock response and the processed analysis
// into the provided envelope. The references are also mapped to concise keys
// (turn2Raw and turn2Processed) for downstream use.
func (s *StorageManager) SaveTurn2Outputs(ctx context.Context, envelope *s3state.Envelope, raw []byte, processed interface{}) (models.S3Reference, models.S3Reference, error) {
	if envelope == nil {
		return models.S3Reference{}, models.S3Reference{}, nil
	}

	// Save raw bytes
	if err := s.manager.SaveToEnvelope(envelope, "responses", "turn2-raw-response.json", json.RawMessage(raw)); err != nil {
		return models.S3Reference{}, models.S3Reference{}, err
	}
	rawRef := envelope.GetReference("responses_turn2-raw-response")
	if rawRef != nil {
		envelope.AddReference("turn2Raw", rawRef)
	}

	// Save processed structure
	if err := s.manager.SaveToEnvelope(envelope, "responses", "turn2-processed-response.json", processed); err != nil {
		return models.S3Reference{}, models.S3Reference{}, err
	}
	procRef := envelope.GetReference("responses_turn2-processed-response")
	if procRef != nil {
		envelope.AddReference("turn2Processed", procRef)
	}

	return toModelRef(rawRef), toModelRef(procRef), nil
}

// SaveTurn2Prompt stores the rendered Turn2 prompt into the envelope and returns the reference.
func (s *StorageManager) SaveTurn2Prompt(ctx context.Context, envelope *s3state.Envelope, prompt string) (models.S3Reference, error) {
	if envelope == nil {
		return models.S3Reference{}, nil
	}

	if err := s.manager.SaveToEnvelope(envelope, "prompts", "turn2-prompt.json", json.RawMessage([]byte(prompt))); err != nil {
		return models.S3Reference{}, err
	}

	ref := envelope.GetReference("prompts_turn2-prompt")
	if ref != nil {
		envelope.AddReference("prompts_turn2", ref)
		return toModelRef(ref), nil
	}
	return models.S3Reference{}, nil
}

func toModelRef(ref *s3state.Reference) models.S3Reference {
	if ref == nil {
		return models.S3Reference{}
	}
	return models.S3Reference{Bucket: ref.Bucket, Key: ref.Key, Size: ref.Size}
}
