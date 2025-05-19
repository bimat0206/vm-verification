package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// S3URL represents a parsed S3 URL
type S3URL struct {
	Bucket string
	Key    string
	Region string
}

// ParseS3URL parses an S3 URL into bucket and key components
func ParseS3URL(s3url string) (S3URL, error) {
	// Handle s3:// protocol
	if strings.HasPrefix(s3url, "s3://") {
		parts := strings.SplitN(s3url[5:], "/", 2)
		if len(parts) != 2 {
			return S3URL{}, fmt.Errorf("invalid S3 URL format: %s", s3url)
		}
		return S3URL{
			Bucket: parts[0],
			Key:    parts[1],
		}, nil
	}

	// Handle https://bucket.s3.region.amazonaws.com/key format
	s3RegionRegex := regexp.MustCompile(`https?://([^.]+)\.s3[.-]([^.]+)\.amazonaws\.com/(.+)`)
	matches := s3RegionRegex.FindStringSubmatch(s3url)
	if len(matches) == 4 {
		return S3URL{
			Bucket: matches[1],
			Key:    matches[3],
			Region: matches[2],
		}, nil
	}

	// Handle https://s3.region.amazonaws.com/bucket/key format
	s3BucketRegex := regexp.MustCompile(`https?://s3[.-]([^.]+)\.amazonaws\.com/([^/]+)/(.+)`)
	matches = s3BucketRegex.FindStringSubmatch(s3url)
	if len(matches) == 4 {
		return S3URL{
			Bucket: matches[2],
			Key:    matches[3],
			Region: matches[1],
		}, nil
	}

	// Try to parse as a URL
	parsedURL, err := url.Parse(s3url)
	if err != nil {
		return S3URL{}, fmt.Errorf("failed to parse S3 URL: %w", err)
	}

	// Extract bucket and key from path
	path := parsedURL.Path
	if path == "" {
		return S3URL{}, fmt.Errorf("invalid S3 URL format: %s", s3url)
	}

	// Remove leading slash
	if path[0] == '/' {
		path = path[1:]
	}

	// Split path into bucket and key
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return S3URL{}, fmt.Errorf("invalid S3 URL format: %s", s3url)
	}

	return S3URL{
		Bucket: parts[0],
		Key:    parts[1],
	}, nil
}

// IsImageContentType checks if a content type is an image
func IsImageContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/")
}
