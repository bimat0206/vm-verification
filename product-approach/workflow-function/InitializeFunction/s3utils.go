// Package main provides the entry point for the initialization Lambda function
// This file contains adapters for the shared s3utils package
package main

import (
	"context"
	"workflow-function/shared/s3utils"
)

// S3URL type alias for backward compatibility 
type S3URL = s3utils.S3URL

// S3Utils type alias for backward compatibility
type S3Utils = s3utils.S3Utils

// NewS3Utils creates an instance of S3Utils for backward compatibility
func NewS3Utils(client interface{}, logger Logger) *S3Utils {
	return s3utils.New(client, logger)
}

// SetConfig is a compatibility wrapper that will be removed in future versions
func (u *S3Utils) SetConfig(config ConfigVars) {
	// This method is no longer needed as the s3utils package doesn't
	// require these specific configurations, but we keep it for compatibility
}

// Wrapper for backward compatibility to match old API signature
func (u *S3Utils) ValidateImageExists(ctx context.Context, s3Url string) error {
	// Default maximum size to 10MB
	maxSize := int64(10 * 1024 * 1024)
	
	// Use the shared implementation but convert the return value
	exists, err := u.ValidateImageExistsWithSize(ctx, s3Url, maxSize)
	if err != nil {
		return err
	}
	
	if !exists {
		return context.DeadlineExceeded // Just to satisfy old signature
	}
	
	return nil
}

// Adapter method for validation that matches shared package signature
func (u *S3Utils) ValidateImageExistsWithSize(ctx context.Context, s3Url string, maxSize int64) (bool, error) {
	return s3utils.S3Utils.ValidateImageExists(u, ctx, s3Url, maxSize)
}