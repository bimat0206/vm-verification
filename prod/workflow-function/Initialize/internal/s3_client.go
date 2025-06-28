package internal

import (
	"context"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"workflow-function/shared/logger"
)

// S3Client wraps S3 operations with consistent error handling
type S3Client struct {
	client *s3.Client
	logger logger.Logger
	config Config
}

// NewS3Client creates a new S3Client with the provided AWS config
func NewS3Client(awsConfig aws.Config, cfg Config, log logger.Logger) *S3Client {
	client := s3.NewFromConfig(awsConfig)
	
	return &S3Client{
		client: client,
		logger: log.WithFields(map[string]interface{}{"component": "S3Client"}),
		config: cfg,
	}
}

// HeadObject checks if an object exists in a bucket without fetching the content
func (c *S3Client) HeadObject(ctx context.Context, bucket, key string) (*s3.HeadObjectOutput, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	
	c.logger.Debug("Checking S3 object existence", map[string]interface{}{
		"bucket": bucket,
		"key":    key,
	})
	
	result, err := c.client.HeadObject(ctx, input)
	if err != nil {
		c.logger.Error("Failed to check S3 object existence", map[string]interface{}{
			"bucket": bucket,
			"key":    key,
			"error":  err.Error(),
		})
		return nil, err
	}
	
	return result, nil
}

// Client returns the underlying S3 client
func (c *S3Client) Client() *s3.Client {
	return c.client
}