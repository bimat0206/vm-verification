package models

import "workflow-function/shared/schema"

// Turn2Response contains S3 references produced by this function

type Turn2Response struct {
    S3Refs map[string]S3Reference `json:"s3Refs"`
    Status string                 `json:"status"`
    Summary schema.ProcessingMetrics `json:"summary"`
}
