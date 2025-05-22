package s3state

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Manager interface for S3 state operations
type Manager interface {
	Store(category, key string, data []byte) (*Reference, error)
	Retrieve(ref *Reference) ([]byte, error)
	StoreJSON(category, key string, data interface{}) (*Reference, error)
	RetrieveJSON(ref *Reference, target interface{}) error
	SaveToEnvelope(envelope *Envelope, category, filename string, data interface{}) error
	GetStateBucket() string
}

// manager implements the Manager interface
type manager struct {
	s3Client *s3.Client
	bucket   string
}

// New creates a new S3 state manager
func New(bucket string) (Manager, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &manager{
		s3Client: s3.NewFromConfig(cfg),
		bucket:   bucket,
	}, nil
}

// GetStateBucket returns the configured state bucket name
func (m *manager) GetStateBucket() string {
	return m.bucket
}

// Store saves raw bytes to S3 with category-based organization
func (m *manager) Store(category, key string, data []byte) (*Reference, error) {
	s3Key := fmt.Sprintf("%s/%s", category, key)
	
	err := m.putObject(s3Key, data, "application/octet-stream")
	if err != nil {
		return nil, fmt.Errorf("failed to store object: %w", err)
	}

	return &Reference{
		Bucket: m.bucket,
		Key:    s3Key,
		Size:   int64(len(data)),
	}, nil
}

// Retrieve gets raw bytes from S3
func (m *manager) Retrieve(ref *Reference) ([]byte, error) {
	if ref == nil {
		return nil, fmt.Errorf("reference is nil")
	}

	data, err := m.getObject(ref.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve object: %w", err)
	}

	return data, nil
}

// StoreJSON marshals data to JSON and stores it in S3
func (m *manager) StoreJSON(category, key string, data interface{}) (*Reference, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Build S3 key properly handling empty categories
	var s3Key string
	if category == "" {
		s3Key = key
	} else {
		s3Key = fmt.Sprintf("%s/%s", category, key)
	}
	
	err = m.putObject(s3Key, jsonData, "application/json")
	if err != nil {
		return nil, fmt.Errorf("failed to store JSON object: %w", err)
	}

	return &Reference{
		Bucket: m.bucket,
		Key:    s3Key,
		Size:   int64(len(jsonData)),
	}, nil
}

// RetrieveJSON gets data from S3 and unmarshals it into target
func (m *manager) RetrieveJSON(ref *Reference, target interface{}) error {
	if ref == nil {
		return fmt.Errorf("reference is nil")
	}

	data, err := m.getObject(ref.Key)
	if err != nil {
		return fmt.Errorf("failed to retrieve JSON object: %w", err)
	}

	err = json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// isDatePath checks if a path contains a date-based hierarchy (YYYY/MM/DD)
func isDatePath(path string) bool {
	// Match pattern like "2025/05/20" at the beginning of path
	re := regexp.MustCompile(`^(\d{4}/\d{2}/\d{2})`)
	return re.MatchString(path)
}

// extractVerificationID extracts verification ID from path component
func extractVerificationID(pathComponent string) (string, bool) {
	// Check if the component starts with "verif-" which is a common prefix
	if strings.HasPrefix(pathComponent, "verif-") {
		return pathComponent, true
	}
	return "", false
}

// SaveToEnvelope stores data and adds reference to envelope
func (m *manager) SaveToEnvelope(envelope *Envelope, category, filename string, data interface{}) error {
	if envelope == nil {
		return fmt.Errorf("envelope is nil")
	}

	// Create current date components for hierarchical path
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	
	// Generate the date-based path prefix
	datePath := fmt.Sprintf("%s/%s/%s", year, month, day)
	
	// Build the S3 key with date-based hierarchy
	var key string
	
	// Check if category already contains a date path or verification ID
	if isDatePath(category) {
		// Category already contains date path, don't add it again
		if strings.Contains(category, envelope.VerificationID) {
			// Both date path and verification ID are present
			key = fmt.Sprintf("%s/%s.json", category, filename)
		} else {
			// Date path present but no verification ID
			key = fmt.Sprintf("%s/%s/%s.json", category, envelope.VerificationID, filename)
		}
	} else if strings.Contains(category, envelope.VerificationID) {
		// Category contains verification ID but no date path
		key = fmt.Sprintf("%s/%s/%s.json", datePath, category, filename)
	} else {
		// Standard case: neither date path nor verification ID in category
		key = fmt.Sprintf("%s/%s/%s/%s.json", datePath, envelope.VerificationID, category, filename)
	}
	
	// Store the data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	err = m.putObject(key, jsonData, "application/json")
	if err != nil {
		return fmt.Errorf("failed to store JSON object: %w", err)
	}
	
	// Create reference
	ref := &Reference{
		Bucket: m.bucket,
		Key:    key,
		Size:   int64(len(jsonData)),
	}
	
	// Initialize references map if needed
	if envelope.References == nil {
		envelope.References = make(map[string]*Reference)
	}
	
	// Create clean reference key without path components
	// Simply use {category}_{filename} format
	refKey := fmt.Sprintf("%s_%s", category, strings.TrimSuffix(filename, ".json"))
	envelope.References[refKey] = ref
	
	return nil
}

// putObject uploads data to S3 with retry logic
func (m *manager) putObject(key string, data []byte, contentType string) error {
	const maxRetries = 3
	const baseDelay = 100 * time.Millisecond

	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := m.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:      aws.String(m.bucket),
			Key:         aws.String(key),
			Body:        bytes.NewReader(data),
			ContentType: aws.String(contentType),
		})
		
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Don't retry on the last attempt
		if attempt < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
			time.Sleep(delay)
		}
	}
	
	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// getObject downloads data from S3 with retry logic
func (m *manager) getObject(key string) ([]byte, error) {
	const maxRetries = 3
	const baseDelay = 100 * time.Millisecond

	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := m.s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(m.bucket),
			Key:    aws.String(key),
		})
		
		if err == nil {
			defer result.Body.Close()
			
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(result.Body)
			if err != nil {
				lastErr = err
				continue
			}
			
			return buf.Bytes(), nil
		}
		
		lastErr = err
		
		// Don't retry on certain error types
		if strings.Contains(err.Error(), "NoSuchKey") {
			return nil, fmt.Errorf("object not found: %s", key)
		}
		
		// Don't retry on the last attempt
		if attempt < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
			time.Sleep(delay)
		}
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// Helper function to create a manager from environment (for Lambda use)
func NewFromEnv() (Manager, error) {
	// This would typically read from environment variables
	// For now, it's a placeholder that would need to be implemented
	// based on your specific environment setup
	return nil, fmt.Errorf("NewFromEnv not implemented - use New() with bucket name")
}