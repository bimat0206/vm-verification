package prefix

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateLayoutPrefix generates a unique prefix for layout processing
// Format: YYYYMMDD-HHMMSS-XXXXX where XXXXX is a random 5-character alphanumeric string
func (g *Generator) GenerateLayoutPrefix() (string, error) {
	now := time.Now()
	
	// Format: YYYYMMDD-HHMMSS
	dateTime := now.Format("20060102-150405")
	
	// Generate random 5-character alphanumeric suffix
	randomSuffix, err := g.generateRandomString(5)
	if err != nil {
		return "", fmt.Errorf("failed to generate random suffix: %v", err)
	}
	
	return fmt.Sprintf("%s-%s", dateTime, randomSuffix), nil
}

// generateRandomString generates a random alphanumeric string of specified length
func (g *Generator) generateRandomString(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	
	return string(result), nil
}

// IsValidPrefix checks if a prefix follows the expected format
func (g *Generator) IsValidPrefix(prefix string) bool {
	parts := strings.Split(prefix, "-")
	if len(parts) != 3 {
		return false
	}
	
	// Check date part (YYYYMMDD)
	if len(parts[0]) != 8 {
		return false
	}
	
	// Check time part (HHMMSS)
	if len(parts[1]) != 6 {
		return false
	}
	
	// Check random suffix (5 characters)
	if len(parts[2]) != 5 {
		return false
	}
	
	return true
}
