package images

import (
	"encoding/base64"
	"strings"
)

// detectImageFormatFromHeader detects image format from file header
func detectImageFormatFromHeader(data []byte) string {
	if len(data) < 8 {
		return ""
	}

	// Check JPEG
	if data[0] == 0xFF && data[1] == 0xD8 {
		return "jpeg"
	}

	// Check PNG
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return "png"
	}

	// Check GIF
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "gif"
	}

	// Check WebP
	if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return "webp"
	}

	// If no match, return empty
	return ""
}

// isValidBedrockImageFormat checks if the format is supported by Bedrock
func isValidBedrockImageFormat(format string) bool {
	format = strings.ToLower(format)
	return format == "jpeg" || format == "jpg" || format == "png"
}

// isValidBase64 checks if a string is valid Base64
func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// getMimeTypeForFormat returns the MIME type for a given image format
func getMimeTypeForFormat(format string) string {
	format = strings.ToLower(format)
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}