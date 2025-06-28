package internal

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	
	"workflow-function/shared/logger"
)

var (
	ErrInvalidS3URL        = errors.New("invalid S3 URL format")
	ErrMissingBucket       = errors.New("missing bucket name in S3 URL")
	ErrMissingKey          = errors.New("missing key in S3 URL")
	ErrInvalidImageFormat  = errors.New("invalid image format")
)

// S3URLParser provides methods to parse and validate S3 URLs
type S3URLParser struct {
	logger logger.Logger
	config Config
}

// NewS3URLParser creates a new S3URLParser
func NewS3URLParser(cfg Config, log logger.Logger) *S3URLParser {
	return &S3URLParser{
		logger: log.WithFields(map[string]interface{}{"component": "S3URLParser"}),
		config: cfg,
	}
}

// ParseS3URL extracts bucket and key from an S3 URL
// Supports both s3:// and https://bucket.s3.region.amazonaws.com formats
func (p *S3URLParser) ParseS3URL(s3URL string) (string, string, error) {
	p.logger.Debug("Parsing S3 URL", map[string]interface{}{
		"url": s3URL,
	})
	
	// Check for empty URL
	if s3URL == "" {
		return "", "", fmt.Errorf("%w: URL is empty", ErrInvalidS3URL)
	}
	
	// Parse the URL
	parsedURL, err := url.Parse(s3URL)
	if err != nil {
		return "", "", fmt.Errorf("%w: %s", ErrInvalidS3URL, err.Error())
	}
	
	var bucket, key string
	
	// Handle s3:// scheme
	if parsedURL.Scheme == "s3" {
		bucket = parsedURL.Host
		key = strings.TrimPrefix(parsedURL.Path, "/")
	} else if parsedURL.Scheme == "http" || parsedURL.Scheme == "https" {
		// Handle https://bucket.s3.region.amazonaws.com/key format
		host := parsedURL.Host
		
		// Virtual hosted style - bucket.s3.region.amazonaws.com
		if strings.Contains(host, ".s3.") && strings.HasSuffix(host, ".amazonaws.com") {
			parts := strings.Split(host, ".")
			bucket = parts[0]
			key = strings.TrimPrefix(parsedURL.Path, "/")
		} else if host == "s3.amazonaws.com" || strings.HasSuffix(host, ".s3.amazonaws.com") {
			// Path style - s3.amazonaws.com/bucket/key
			pathParts := strings.SplitN(strings.TrimPrefix(parsedURL.Path, "/"), "/", 2)
			if len(pathParts) >= 1 {
				bucket = pathParts[0]
			}
			if len(pathParts) >= 2 {
				key = pathParts[1]
			}
		} else {
			return "", "", fmt.Errorf("%w: unsupported S3 URL format: %s", ErrInvalidS3URL, s3URL)
		}
	} else {
		return "", "", fmt.Errorf("%w: unsupported scheme: %s", ErrInvalidS3URL, parsedURL.Scheme)
	}
	
	// Validate bucket and key
	if bucket == "" {
		return "", "", fmt.Errorf("%w: %s", ErrMissingBucket, s3URL)
	}
	
	if key == "" {
		return "", "", fmt.Errorf("%w: %s", ErrMissingKey, s3URL)
	}
	
	// Log successful parsing result
	p.logger.Debug("Parsed S3 URL", map[string]interface{}{
		"url":    s3URL,
		"bucket": bucket,
		"key":    key,
	})
	
	return bucket, key, nil
}

// IsValidImageExtension checks if a file has a valid image extension
func (p *S3URLParser) IsValidImageExtension(key string) bool {
	ext := strings.ToLower(filepath.Ext(key))
	
	// Valid image extensions
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	
	return validExtensions[ext]
}

// ParseS3URLs parses and validates multiple S3 URLs at once
func (p *S3URLParser) ParseS3URLs(urls ...string) ([]string, []string, error) {
	if len(urls) == 0 {
		return nil, nil, nil
	}
	
	buckets := make([]string, len(urls))
	keys := make([]string, len(urls))
	
	for i, url := range urls {
		bucket, key, err := p.ParseS3URL(url)
		if err != nil {
			return nil, nil, err
		}
		
		if !p.IsValidImageExtension(key) {
			return nil, nil, fmt.Errorf("%w: %s", ErrInvalidImageFormat, url)
		}
		
		buckets[i] = bucket
		keys[i] = key
	}
	
	return buckets, keys, nil
}