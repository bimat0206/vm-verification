package s3helper

import (
	"context"
	"encoding/json"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// GetS3Object returns the bytes of an object from S3
func GetS3Object(ctx context.Context, client *s3.Client, bucket, key string) ([]byte, error) {
	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()
	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetS3ObjectAsJSON fetches an object and unmarshals JSON into target
func GetS3ObjectAsJSON(ctx context.Context, client *s3.Client, bucket, key string, target interface{}) error {
	bytes, err := GetS3Object(ctx, client, bucket, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}
