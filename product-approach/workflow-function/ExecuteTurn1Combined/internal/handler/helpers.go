package handler

import (
	"strings"
	"time"
)

// extractCheckingImageUrl extracts the checking image URL from the S3 key
// Expected format: verifications/{verificationId}/images/{checkingImageUrl}
func extractCheckingImageUrl(s3Key string) string {
	parts := strings.Split(s3Key, "/")
	if len(parts) >= 4 && parts[2] == "images" {
		// The checking image URL is typically the last part
		return parts[len(parts)-1]
	}
	return ""
}

// calculateHoursSince calculates hours elapsed since the given timestamp
func calculateHoursSince(timestamp string) float64 {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return 0
	}
	return time.Since(t).Hours()
}