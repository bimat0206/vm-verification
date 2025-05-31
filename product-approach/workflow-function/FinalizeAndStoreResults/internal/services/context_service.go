package services

import (
	"context"
	"time"

	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

// ContextService provides helpers around VerificationContext persistence
type ContextService struct {
	logger logger.Logger
	s3Mgr  s3state.Manager
}

// NewContextService creates a new ContextService
func NewContextService(log logger.Logger, mgr s3state.Manager) *ContextService {
	return &ContextService{logger: log, s3Mgr: mgr}
}

// LoadOrCreateVerificationContext loads the context from S3 or creates minimal one
func (s *ContextService) LoadOrCreateVerificationContext(ctx context.Context, verificationID string, initRef *schema.S3Reference, defaultStatus string, errInfo *schema.ErrorInfo) (*schema.VerificationContext, error) {
	var vc schema.VerificationContext
	if initRef != nil {
		if err := s.s3Mgr.RetrieveJSON(initRef, &vc); err == nil {
			return &vc, nil
		}
		s.logger.Warn("Initialization context not found or failed to load", map[string]interface{}{
			"verificationId": verificationID,
			"ref":            initRef.Key,
		})
	}
	now := time.Now().UTC().Format(time.RFC3339)
	vc = schema.VerificationContext{
		VerificationId: verificationID,
		VerificationAt: now,
		Status:         defaultStatus,
		CurrentStatus:  defaultStatus,
		LastUpdatedAt:  now,
		Error:          errInfo,
	}
	return &vc, nil
}

// UpdateVerificationContextWithError updates the context with error info
func (s *ContextService) UpdateVerificationContextWithError(vc *schema.VerificationContext, stage string, errInfo schema.ErrorInfo, statusMapper func(string) string) {
	status := statusMapper(stage)
	vc.Status = status
	vc.CurrentStatus = status
	vc.LastUpdatedAt = errInfo.Timestamp
	vc.Error = &errInfo

	if vc.ErrorTracking == nil {
		vc.ErrorTracking = &schema.ErrorTracking{}
	}
	vc.ErrorTracking.HasErrors = true
	vc.ErrorTracking.CurrentError = &errInfo
	vc.ErrorTracking.ErrorHistory = append(vc.ErrorTracking.ErrorHistory, errInfo)
	vc.ErrorTracking.LastErrorAt = errInfo.Timestamp

	vc.StatusHistory = append(vc.StatusHistory, schema.StatusHistoryEntry{
		Status:    status,
		Timestamp: errInfo.Timestamp,
	})
}
