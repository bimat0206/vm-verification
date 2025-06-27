// Package main provides the entry point for the initialization Lambda function
// This file contains adapters for the shared s3utils package
package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/s3utils"
)

// S3Utils wrapper struct for local extension methods
type S3UtilsWrapper struct {
	*s3utils.S3Utils
}

// NewS3Utils creates an instance of S3Utils for backward compatibility
func NewS3Utils(client *s3.Client, logger Logger) *S3UtilsWrapper {
	return &S3UtilsWrapper{
		S3Utils: s3utils.New(client, logger),
	}
}

// SetConfig is a compatibility wrapper that will be removed in future versions
func (u *S3UtilsWrapper) SetConfig(config ConfigVars) {
	// This method is no longer needed as the s3utils package doesn't
	// require these specific configurations, but we keep it for compatibility
}

// ValidateImageExists wraps the shared implementation for backward compatibility
func (u *S3UtilsWrapper) ValidateImageExists(ctx context.Context, s3Url string) error {
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

// ValidateImageExistsWithSize adapter method for validation that matches shared package signature
func (u *S3UtilsWrapper) ValidateImageExistsWithSize(ctx context.Context, s3Url string, maxSize int64) (bool, error) {
	return u.S3Utils.ValidateImageExists(ctx, s3Url, maxSize)
}