// Package schema provides image data structures and methods
package schema

// ImageData represents the standardized structure for image references with S3 Base64 storage
// Supports both new format (Reference/Checking) and legacy format (ReferenceImage/CheckingImage)
type ImageData struct {
	// New format (primary)
	Reference *ImageInfo `json:"reference,omitempty"`
	Checking  *ImageInfo `json:"checking,omitempty"`
	
	// Legacy format support for backward compatibility
	ReferenceImage *ImageInfo `json:"referenceImage,omitempty"`
	CheckingImage  *ImageInfo `json:"checkingImage,omitempty"`
	
	// Processing metadata
	ProcessedAt           string                `json:"processedAt,omitempty"`
	Base64Generated       bool                  `json:"base64Generated"`
	BucketOwner           string                `json:"bucketOwner,omitempty"`
	
	// S3 storage metadata
	StorageDecisionAt     string                `json:"storageDecisionAt,omitempty"`
	TotalS3References     int                   `json:"totalS3References"`      // Count of S3-stored Base64 references
	StorageConfig         *S3StorageConfig      `json:"storageConfig,omitempty"`
}

// GetReference returns the reference image, supporting both formats
func (images *ImageData) GetReference() *ImageInfo {
	if images.Reference != nil {
		return images.Reference
	}
	return images.ReferenceImage
}

// GetChecking returns the checking image, supporting both formats
func (images *ImageData) GetChecking() *ImageInfo {
	if images.Checking != nil {
		return images.Checking
	}
	return images.CheckingImage
}

// SetReference sets the reference image in both formats for compatibility
func (images *ImageData) SetReference(img *ImageInfo) {
	images.Reference = img
	images.ReferenceImage = img
}

// SetChecking sets the checking image in both formats for compatibility
func (images *ImageData) SetChecking(img *ImageInfo) {
	images.Checking = img
	images.CheckingImage = img
}

// GetTotalBase64Size returns the total size of all Base64 data
func (images *ImageData) GetTotalBase64Size() int64 {
	var total int64
	
	ref := images.GetReference()
	if ref != nil {
		total += ref.GetBase64SizeEstimate()
	}
	
	checking := images.GetChecking()
	if checking != nil {
		total += checking.GetBase64SizeEstimate()
	}
	
	return total
}

// ValidateForPayloadLimits checks if the total size is within limits
func (images *ImageData) ValidateForPayloadLimits() error {
	// With S3-only storage, we don't need to check payload limits
	// as Base64 data is always stored in S3
	return nil
}

// GetStorageSummary returns a summary of storage methods used
func (images *ImageData) GetStorageSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"base64Generated":      images.Base64Generated,
		"totalS3References":    images.TotalS3References,
		"processedAt":          images.ProcessedAt,
	}
	
	ref := images.GetReference()
	if ref != nil {
		summary["referenceStorage"] = ref.GetStorageInfo()
	}
	
	checking := images.GetChecking()
	if checking != nil {
		summary["checkingStorage"] = checking.GetStorageInfo()
	}
	
	return summary
}

// UpdateStorageSummary updates the storage summary in ImageData
func (images *ImageData) UpdateStorageSummary() {
	var totalS3References int
	
	// Handle reference image (supporting both formats)
	ref := images.GetReference()
	if ref != nil && ref.IsS3TemporaryStorage() {
		totalS3References++
	}
	
	// Handle checking image (supporting both formats)
	checking := images.GetChecking()
	if checking != nil && checking.IsS3TemporaryStorage() {
		totalS3References++
	}
	
	images.TotalS3References = totalS3References
}
