package main

import (
    "errors"
    "fmt"
    "regexp"
    "strings"
)

// VerificationContext represents the context for the verification workflow
type VerificationContext struct {
    VerificationId     string  `json:"verificationId"`
    VerificationAt     string  `json:"verificationAt,omitempty"`
    Status             string  `json:"status,omitempty"`
    VerificationType   string  `json:"verificationType"`
    VendingMachineId   string  `json:"vendingMachineId,omitempty"`
    LayoutId           int64   `json:"layoutId,omitempty"`
    LayoutPrefix       string  `json:"layoutPrefix,omitempty"`
    ReferenceImageUrl  string  `json:"referenceImageUrl"`
    CheckingImageUrl   string  `json:"checkingImageUrl"`
    PreviousVerificationId string `json:"previousVerificationId,omitempty"`
    NotificationEnabled bool   `json:"notificationEnabled,omitempty"`
}

// MachineStructure represents the physical structure of a vending machine
type MachineStructure struct {
    RowCount        int      `json:"rowCount"`
    ColumnsPerRow   int      `json:"columnsPerRow"`
    RowOrder        []string `json:"rowOrder"`
    ColumnOrder     []string `json:"columnOrder"`
}

// LayoutMetadata holds layout information from DynamoDB.
type LayoutMetadata struct {
    LayoutId          int64                  `json:"layoutId"`
    LayoutPrefix      string                 `json:"layoutPrefix"`
    VendingMachineId  string                 `json:"vendingMachineId,omitempty"`
    Location          string                 `json:"location,omitempty"`
    CreatedAt         string                 `json:"createdAt,omitempty"`
    UpdatedAt         string                 `json:"updatedAt,omitempty"`
    ReferenceImageUrl string                 `json:"referenceImageUrl,omitempty"`
    SourceJsonUrl     string                 `json:"sourceJsonUrl,omitempty"`
    MachineStructure  *MachineStructure      `json:"machineStructure,omitempty"`
    ProductPositionMap map[string]interface{} `json:"productPositionMap,omitempty"`
    RowProductMapping map[string]interface{} `json:"rowProductMapping,omitempty"`
}

// HistoricalContext holds previous verification summary.
type HistoricalContext struct {
    PreviousVerificationId      string                 `json:"previousVerificationId"`
    PreviousVerificationAt      string                 `json:"previousVerificationAt,omitempty"`
    PreviousVerificationStatus  string                 `json:"previousVerificationStatus,omitempty"`
    HoursSinceLastVerification  float64                `json:"hoursSinceLastVerification,omitempty"`
    MachineStructure            *MachineStructure      `json:"machineStructure,omitempty"`
    CheckingStatus              map[string]string      `json:"checkingStatus,omitempty"`
    Summary                     map[string]interface{} `json:"summary,omitempty"`
    VerificationSummary         map[string]interface{} `json:"verificationSummary,omitempty"`
}

// FetchImagesRequest represents the expected input to the Lambda function.
type FetchImagesRequest struct {
    VerificationContext *VerificationContext `json:"verificationContext,omitempty"`
    // Direct fields for backward compatibility or direct invocation
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
    // If we have a verificationContext object, validate it
    if r.VerificationContext != nil {
        if r.VerificationContext.VerificationId == "" {
            return errors.New("verificationContext.verificationId is required")
        }
        if r.VerificationContext.VerificationType == "" {
            return errors.New("verificationContext.verificationType is required")
        }
        if r.VerificationContext.ReferenceImageUrl == "" {
            return errors.New("verificationContext.referenceImageUrl is required")
        }
        if r.VerificationContext.CheckingImageUrl == "" {
            return errors.New("verificationContext.checkingImageUrl is required")
        }

        // Validate verification type
        switch r.VerificationContext.VerificationType {
        case "LAYOUT_VS_CHECKING":
            // For LAYOUT_VS_CHECKING, we need layoutId and layoutPrefix
            if r.VerificationContext.LayoutId == 0 {
                return errors.New("verificationContext.layoutId is required for LAYOUT_VS_CHECKING verification type")
            }
            if r.VerificationContext.LayoutPrefix == "" {
                return errors.New("verificationContext.layoutPrefix is required for LAYOUT_VS_CHECKING verification type")
            }
            // previousVerificationId is not required for LAYOUT_VS_CHECKING
        case "PREVIOUS_VS_CURRENT":
            // For PREVIOUS_VS_CURRENT, we need previousVerificationId
            if r.VerificationContext.PreviousVerificationId == "" {
                return errors.New("verificationContext.previousVerificationId is required for PREVIOUS_VS_CURRENT verification type")
            }
            // Validate that referenceImageUrl is from the checking bucket
            if !strings.Contains(r.VerificationContext.ReferenceImageUrl, "checking") {
                return errors.New("for PREVIOUS_VS_CURRENT verification, referenceImageUrl should point to a previous checking image")
            }
        default:
            return fmt.Errorf("unsupported verificationType: %s (must be LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT)",
                r.VerificationContext.VerificationType)
        }

        // Validation passed
        return nil
    }

    // Otherwise validate direct fields
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

    // Validate verification type
    switch r.VerificationType {
    case "LAYOUT_VS_CHECKING":
        // For LAYOUT_VS_CHECKING, we need layoutId and layoutPrefix
        if r.LayoutId == 0 {
            return errors.New("layoutId is required for LAYOUT_VS_CHECKING verification type")
        }
        if r.LayoutPrefix == "" {
            return errors.New("layoutPrefix is required for LAYOUT_VS_CHECKING verification type")
        }
        // previousVerificationId is not required for LAYOUT_VS_CHECKING
    case "PREVIOUS_VS_CURRENT":
        // For PREVIOUS_VS_CURRENT, we need previousVerificationId
        if r.PreviousVerificationId == "" {
            return errors.New("previousVerificationId is required for PREVIOUS_VS_CURRENT verification type")
        }
        // Validate that referenceImageUrl is from the checking bucket
        if !strings.Contains(r.ReferenceImageUrl, "checking") {
            return errors.New("for PREVIOUS_VS_CURRENT verification, referenceImageUrl should point to a previous checking image")
        }
    default:
        return fmt.Errorf("unsupported verificationType: %s (must be LAYOUT_VS_CHECKING or PREVIOUS_VS_CURRENT)",
            r.VerificationType)
    }
    
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
    BucketOwner   string `json:"bucketOwner"`
    Bucket        string `json:"bucket"`
    Key           string `json:"key"`
}

// ImagesData contains metadata for both reference and checking images
type ImagesData struct {
    ReferenceImageMeta  ImageMetadata `json:"referenceImageMeta"`
    CheckingImageMeta   ImageMetadata `json:"checkingImageMeta"`
}

// FetchImagesResponse represents the Lambda output.
type FetchImagesResponse struct {
    VerificationContext VerificationContext `json:"verificationContext"`
    Images              ImagesData          `json:"images"`
    LayoutMetadata      *LayoutMetadata     `json:"layoutMetadata,omitempty"`
    HistoricalContext   *HistoricalContext  `json:"historicalContext,omitempty"`
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