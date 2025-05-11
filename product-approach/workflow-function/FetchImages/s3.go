package main

import (
    "context"
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/sts"
)

// ParseS3URI parses an S3 URI (s3://bucket/key) into a struct.
func ParseS3URI(uri string) (S3URI, error) {
    if !strings.HasPrefix(uri, "s3://") {
        return S3URI{}, fmt.Errorf("invalid S3 URI: %s", uri)
    }
    parts := strings.SplitN(uri[5:], "/", 2)
    if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
        return S3URI{}, fmt.Errorf("invalid S3 URI format: %s", uri)
    }
    return S3URI{
        Bucket: parts[0],
        Key:    parts[1],
    }, nil
}

// GetS3ImageMetadata fetches S3 object metadata (HEAD request).
func GetS3ImageMetadata(ctx context.Context, s3uri S3URI) (ImageMetadata, error) {
    // Set default bucket owner to empty string
    bucketOwner := ""
    
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return ImageMetadata{}, fmt.Errorf("failed to load AWS config: %w", err)
    }
    
    // Get object metadata first
    client := s3.NewFromConfig(cfg)
    out, err := client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(s3uri.Bucket),
        Key:    aws.String(s3uri.Key),
    })
    if err != nil {
        return ImageMetadata{}, fmt.Errorf("failed to get S3 object metadata: %w", err)
    }
    
    // Try to get the AWS account ID using STS
    if accountID := os.Getenv("AWS_ACCOUNT_ID"); accountID != "" {
        // Use environment variable if available (most reliable)
        bucketOwner = accountID
        Info("Using AWS_ACCOUNT_ID for bucket owner", map[string]interface{}{
            "accountId": bucketOwner,
        })
    } else {
        // Use STS GetCallerIdentity as a fallback
        stsClient := sts.NewFromConfig(cfg)
        identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
        
        if err == nil && identity.Account != nil {
            bucketOwner = *identity.Account
            Info("Retrieved account ID from STS", map[string]interface{}{
                "accountId": bucketOwner,
            })
        } else {
            // If STS also fails, leave bucket owner as empty string
            Warn("Could not determine bucket owner", map[string]interface{}{
                "bucket": s3uri.Bucket,
                "error": err,
            })
        }
    }
    
    meta := ImageMetadata{
        ContentType:  aws.ToString(out.ContentType),
        Size:         aws.ToInt64(out.ContentLength),
        LastModified: out.LastModified.Format(time.RFC3339),
        ETag:         aws.ToString(out.ETag),
        BucketOwner:  bucketOwner,
        Bucket:       s3uri.Bucket,
        Key:          s3uri.Key,
    }
    return meta, nil
}