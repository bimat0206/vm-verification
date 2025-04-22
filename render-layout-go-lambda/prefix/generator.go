package prefix

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Generator handles the creation of unique layout prefixes
type Generator struct {
	// You could add configuration here if needed
}

// NewGenerator creates a new prefix generator
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateLayoutPrefix creates a unique identifier for a layout version
// The prefix is composed of:
// - Timestamp elements (year, month, day as single characters)
// - Random string (5 characters)
// This ensures uniqueness while maintaining readability
func (g *Generator) GenerateLayoutPrefix() (string, error) {
	// Get current time
	now := time.Now()
	
	// Extract single character for year, month, day
	// Convert year to single char (e.g., 2025 -> 5)
	year := now.Year() % 10
	// Convert month to single char (1-9, a-c for 10-12)
	var monthChar string
	month := now.Month()
	if month < 10 {
		monthChar = fmt.Sprintf("%d", month)
	} else {
		// Use a-c for months 10-12
		monthChar = string(rune('a' + int(month) - 10))
	}
	// Convert day to single char (1-9, a-v for 10-31)
	var dayChar string
	day := now.Day()
	if day < 10 {
		dayChar = fmt.Sprintf("%d", day)
	} else {
		// Use a-v for days 10-31
		dayChar = string(rune('a' + day - 10))
	}
	
	// Generate a random string (5 characters)
	randomBytes := make([]byte, 3) // 3 bytes will give us 6 hex chars
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}
	randomStr := hex.EncodeToString(randomBytes)[:5] // Take first 5 chars
	
	// Combine all elements into the prefix
	prefix := fmt.Sprintf("%d%s%s%s", year, monthChar, dayChar, randomStr)
	
	return prefix, nil
}