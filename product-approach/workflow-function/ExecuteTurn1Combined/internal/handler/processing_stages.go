package handler

import (
	"time"
	"workflow-function/shared/schema"
)

// ProcessingStagesTracker handles tracking of processing stages throughout the workflow
type ProcessingStagesTracker struct {
	processingStages []schema.ProcessingStage
	startTime        time.Time
}

// NewProcessingStagesTracker creates a new instance of ProcessingStagesTracker
func NewProcessingStagesTracker(startTime time.Time) *ProcessingStagesTracker {
	return &ProcessingStagesTracker{
		processingStages: make([]schema.ProcessingStage, 0),
		startTime:        startTime,
	}
}

// RecordStage records a processing stage with its metadata
func (p *ProcessingStagesTracker) RecordStage(stageName, status string, duration time.Duration, metadata map[string]interface{}) {
	stage := schema.ProcessingStage{
		StageName: stageName,
		StartTime: p.startTime.Add(duration - duration).Format(time.RFC3339), // Approximate start time
		EndTime:   p.startTime.Add(duration).Format(time.RFC3339),
		Duration:  duration.Milliseconds(),
		Status:    status,
		Metadata:  metadata,
	}
	
	p.processingStages = append(p.processingStages, stage)
}

// GetStages returns all recorded processing stages
func (p *ProcessingStagesTracker) GetStages() []schema.ProcessingStage {
	return p.processingStages
}

// GetStageCount returns the number of recorded stages
func (p *ProcessingStagesTracker) GetStageCount() int {
	return len(p.processingStages)
}