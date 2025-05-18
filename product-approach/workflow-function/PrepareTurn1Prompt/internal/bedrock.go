package internal

import (
	"path/filepath"
	"strings"
)

// GetImageFormatFromKey extracts the image format from an S3 key
func getImageFormatFromKey(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	if ext == "" {
		return ""
	}
	
	// Remove leading dot
	ext = ext[1:]
	
	// Check if format is supported
	supportedFormats := map[string]bool{
		"jpeg": true,
		"jpg":  true,
		"png":  true,
	}
	
	if supportedFormats[ext] {
		// Normalize jpg to jpeg for Bedrock
		if ext == "jpg" {
			return "jpeg"
		}
		return ext
	}
	
	return ""
}
