package s3state

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// GetS3ObjectAsJSON fetches an object from S3 and unmarshals its JSON content into target.
func GetS3ObjectAsJSON(ctx context.Context, client *s3.Client, bucket, key string, target interface{}) error {
	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("get object: %w", err)
	}
	defer out.Body.Close()
	decoder := json.NewDecoder(out.Body)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}
