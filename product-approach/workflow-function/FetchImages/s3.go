package main

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    //"github.com/aws/aws-sdk-go-v2/service/s3/types"
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
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return ImageMetadata{}, fmt.Errorf("failed to load AWS config: %w", err)
    }
    client := s3.NewFromConfig(cfg)
    
    // Get object metadata
    out, err := client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(s3uri.Bucket),
        Key:    aws.String(s3uri.Key),
    })
    if err != nil {
        return ImageMetadata{}, fmt.Errorf("failed to get S3 object metadata: %w", err)
    }
    
    // Get bucket owner
    aclOut, err := client.GetBucketAcl(ctx, &s3.GetBucketAclInput{
        Bucket: aws.String(s3uri.Bucket),
    })
    
    // Default bucket owner
    bucketOwner := ""
    
    // Extract bucket owner if available
    if err == nil && aclOut.Owner != nil {
        bucketOwner = aws.ToString(aclOut.Owner.ID)
    } else {
        // Log the error but don't fail the operation
        Error("Failed to get bucket owner", map[string]interface{}{
            "bucket": s3uri.Bucket,
            "error":  err.Error(),
        })
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
