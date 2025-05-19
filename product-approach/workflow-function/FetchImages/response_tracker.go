package main

import (
	"sync"
)

// ResponseSizeTracker tracks the total Base64 size across all images
type ResponseSizeTracker struct {
	mu                    sync.Mutex
	totalBase64Size       int64
	referenceBase64Size   int64
	checkingBase64Size    int64
	estimatedTotalSize    int64
}

// NewResponseSizeTracker creates a new response size tracker
func NewResponseSizeTracker() *ResponseSizeTracker {
	return &ResponseSizeTracker{}
}

// UpdateReferenceSize updates the reference image Base64 size
func (rst *ResponseSizeTracker) UpdateReferenceSize(size int64) {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	rst.referenceBase64Size = size
	rst.updateTotalSize()
}

// UpdateCheckingSize updates the checking image Base64 size
func (rst *ResponseSizeTracker) UpdateCheckingSize(size int64) {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	rst.checkingBase64Size = size
	rst.updateTotalSize()
}

// updateTotalSize recalculates the total size (must be called with lock held)
func (rst *ResponseSizeTracker) updateTotalSize() {
	rst.totalBase64Size = rst.referenceBase64Size + rst.checkingBase64Size
}

// GetTotalSize returns the current total Base64 size
func (rst *ResponseSizeTracker) GetTotalSize() int64 {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	return rst.totalBase64Size
}

// GetEstimatedTotalSize returns the estimated total size including pending images
func (rst *ResponseSizeTracker) GetEstimatedTotalSize() int64 {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	return rst.estimatedTotalSize
}

// SetEstimatedTotal sets the estimated total size for both images
func (rst *ResponseSizeTracker) SetEstimatedTotal(size int64) {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	rst.estimatedTotalSize = size
}

// WouldExceedLimit checks if adding a new Base64 would exceed the response limit
func (rst *ResponseSizeTracker) WouldExceedLimit(additionalSize int64) bool {
	rst.mu.Lock()
	defer rst.mu.Unlock()
	return rst.totalBase64Size+additionalSize > MaxUsableResponseSize
}
