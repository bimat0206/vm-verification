package handler

import (
	"context"
	"strings"
	"testing"

	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
)

type stubS3Manager struct{}

func (s *stubS3Manager) Store(category, key string, data []byte) (*s3state.Reference, error) {
	return nil, nil
}
func (s *stubS3Manager) StoreWithContentType(category, key string, data []byte, ct string) (*s3state.Reference, error) {
	return nil, nil
}
func (s *stubS3Manager) Retrieve(ref *s3state.Reference) ([]byte, error) { return nil, nil }
func (s *stubS3Manager) StoreJSON(category, key string, data interface{}) (*s3state.Reference, error) {
	return &s3state.Reference{Bucket: "b", Key: category + "/" + key}, nil
}
func (s *stubS3Manager) RetrieveJSON(ref *s3state.Reference, target interface{}) error { return nil }
func (s *stubS3Manager) SaveToEnvelope(env *s3state.Envelope, category, filename string, data interface{}) error {
	ref := &s3state.Reference{Bucket: "b", Key: category + "/" + filename}
	env.AddReference(category+"_"+strings.TrimSuffix(filename, ".json"), ref)
	return nil
}
func (s *stubS3Manager) GetStateBucket() string { return "b" }

func TestSaveTurn2Outputs(t *testing.T) {
	env := s3state.NewEnvelope("verif-1")
	mgr := &stubS3Manager{}
	sm := NewStorageManager(mgr, logger.New("test", "test"))

	raw := []byte("{}")
	proc := map[string]string{"a": "b"}

	rawRef, procRef, err := sm.SaveTurn2Outputs(context.Background(), env, raw, proc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.GetReference("turn2Raw") == nil || env.GetReference("turn2Processed") == nil {
		t.Fatalf("references not stored")
	}
	if rawRef.Key == "" || procRef.Key == "" {
		t.Fatalf("empty refs")
	}
}

func TestSaveTurn2Prompt(t *testing.T) {
	env := s3state.NewEnvelope("verif-1")
	mgr := &stubS3Manager{}
	sm := NewStorageManager(mgr, logger.New("test", "test"))

	ref, err := sm.SaveTurn2Prompt(context.Background(), env, "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.GetReference("prompts_turn2") == nil {
		t.Fatalf("prompt reference not stored")
	}
	if ref.Key == "" {
		t.Fatalf("empty ref")
	}
}
