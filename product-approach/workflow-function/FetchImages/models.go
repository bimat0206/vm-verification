package main

import (
    "errors"
    "fmt"
    "regexp"
    //"strings"
)

// FetchImagesRequest represents the expected input to the Lambda function.
type FetchImagesRequest struct {
    VerificationId     string `json:"verificationId"`
    VerificationType   string `json:"verificationType"`
    ReferenceImageUrl  string `json:"referenceImageUrl"`
    CheckingImageUrl   string `json:"checkingImageUrl"`
    LayoutId           int64  `json:"layoutId,omitempty"`
    LayoutPrefix       string `json:"layoutPrefix,omitempty"`
    PreviousVerificationId string `json:"previousVerificationId,omitempty"`
    VendingMachineId   string `json:"vendingMachineId,omitempty"`
}

// Validate checks for required fields and basic format.
func (r *FetchImagesRequest) Validate() error {
    if r.VerificationId == "" {
        return errors.New("verificationId is required")
    }
    if r.VerificationType == "" {
        return errors.New("verificationType is required")
    }
    if r.ReferenceImageUrl == "" {
        return errors.New("referenceImageUrl is required")
    }
    if r.CheckingImageUrl == "" {
        return errors.New("checkingImageUrl is required")
    }
    // Add more validation as needed, e.g., allowed types, S3 URI pattern, etc.
    return nil
}

// S3URI represents a parsed S3 URI (bucket and key).
type S3URI struct {
    Bucket string
    Key    string
}

// ImageMetadata holds S3 object metadata.
type ImageMetadata struct {
    ContentType   string `json:"contentType"`
    Size          int64  `json:"size"`
    LastModified  string `json:"lastModified"`
    ETag          string `json:"etag"`
}

// FetchImagesResponse represents the Lambda output.
type FetchImagesResponse struct {
    VerificationId      string        `json:"verificationId"`
    ReferenceImageUrl   string        `json:"referenceImageUrl"`
    ReferenceImageMeta  ImageMetadata `json:"referenceImageMeta"`
    CheckingImageUrl    string        `json:"checkingImageUrl"`
    CheckingImageMeta   ImageMetadata `json:"checkingImageMeta"`
    // LayoutMetadata, HistoricalContext to be added later
}

// ErrorResponse is a standardized error response.
type ErrorResponse struct {
    ErrorType string `json:"error"`
    Message   string `json:"message"`
    Details   string `json:"details,omitempty"`
}

// Error helpers
func NewBadRequestError(msg string, err error) error {
    return fmt.Errorf("%s: %w", msg, err)
}
func NewNotFoundError(msg string, err error) error {
    return fmt.Errorf("%s: %w", msg, err)
}

// S3 URI validator (simple)
var s3uriPattern = regexp.MustCompile(`^s3://([^/]+)/(.+)$`)

func IsValidS3URI(uri string) bool {
    return s3uriPattern.MatchString(uri)
}
