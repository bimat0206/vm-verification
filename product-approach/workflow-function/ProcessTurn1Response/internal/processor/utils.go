package processor

import (
	"strings"
)

// Contains checks if a string contains a substring (case-insensitive)
func Contains(s string, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}

// MapSize returns the size of a map safely
func MapSize(m map[string]interface{}) int {
	if m == nil {
		return 0
	}
	return len(m)
}

// SliceSize returns the size of a slice safely
func SliceSize(s []interface{}) int {
	if s == nil {
		return 0
	}
	return len(s)
}