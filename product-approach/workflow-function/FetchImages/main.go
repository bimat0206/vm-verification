package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"workflow-function/shared/schema"
)

// Handler is the Lambda handler function
func Handler(ctx context.Context, event interface{}) (FetchImagesResponse, error) {
	// Initialize dependencies
	deps, err := initDependencies(ctx)
	if err != nil {
		return FetchImagesResponse{}, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Load configuration early for image processing
	cfg := LoadConfig()
	
	// Configure dependencies with environment variables
	deps.ConfigureDbUtils(cfg)

	logger := deps.GetLogger()

	// Parse input based on event type
	var req FetchImagesRequest
	var parseErr error

// Replace the section around line 35 with:

	// Log the incoming event for debugging
	eventBytes, _ := json.Marshal(event)
	logger.Info("Received event for Base64 image processing", map[string]interface{}{
		"eventType":    fmt.Sprintf("%T", event),
		"maxImageSize": cfg.MaxImageSize,
		"eventBody":    string(eventBytes), // Use the eventBytes variable
	})

	// Handle different invocation types
	switch e := event.(type) {
	case events.LambdaFunctionURLRequest:
		// Function URL invocation
		if parseErr = json.Unmarshal([]byte(e.Body), &req); parseErr != nil {
			logger.Error("Failed to parse Function URL input", map[string]interface{}{
				"error": parseErr.Error(),
			})
			return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", parseErr)
		}
	case map[string]interface{}:
		// Direct invocation from Step Function
		data, _ := json.Marshal(e)
		if parseErr = json.Unmarshal(data, &req); parseErr != nil {
			logger.Error("Failed to parse Step Function input", map[string]interface{}{
				"error": parseErr.Error(),
				"input": fmt.Sprintf("%+v", e),
			})
			return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", parseErr)
		}
	case FetchImagesRequest:
		// Direct struct invocation
		req = e
	default:
		// Try raw JSON unmarshal as fallback
		data, _ := json.Marshal(event)
		if parseErr = json.Unmarshal(data, &req); parseErr != nil {
			logger.Error("Failed to parse unknown input type", map[string]interface{}{
				"error": parseErr.Error(),
				"input": fmt.Sprintf("%+v", event),
			})
			return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", parseErr)
		}
	}

	// Validate input
	if err := req.Validate(); err != nil {
		return FetchImagesResponse{}, NewBadRequestError("Input validation failed", err)
	}

	// Initialize verificationContext if it doesn't exist or normalize from direct fields
	var verificationContext schema.VerificationContext
	if req.VerificationContext != nil {
		verificationContext = *req.VerificationContext
	} else {
		// Create from direct fields
		verificationContext = schema.VerificationContext{
			VerificationId:         req.VerificationId,
			VerificationType:       req.VerificationType,
			ReferenceImageUrl:      req.ReferenceImageUrl,
			CheckingImageUrl:       req.CheckingImageUrl,
			LayoutId:               req.LayoutId,
			LayoutPrefix:           req.LayoutPrefix,
			VendingMachineId:       req.VendingMachineId,
			PreviousVerificationId: req.PreviousVerificationId,
		}
	}

	// Update status in verification context
	verificationContext.Status = schema.StatusImagesFetched

	// Determine previousVerificationId for PREVIOUS_VS_CURRENT verification type
	var prevVerificationId string
	if verificationContext.VerificationType == schema.VerificationTypePreviousVsCurrent {
		// Make sure previousVerificationId exists for PREVIOUS_VS_CURRENT
		if verificationContext.PreviousVerificationId == "" {
			return FetchImagesResponse{}, NewBadRequestError(
				"PreviousVerificationId is required for PREVIOUS_VS_CURRENT verification type",
				fmt.Errorf("missing previousVerificationId"),
			)
		}
		prevVerificationId = verificationContext.PreviousVerificationId
	}
	// For LAYOUT_VS_CHECKING, prevVerificationId remains empty

	logger.Info("Starting parallel fetch with Base64 encoding", map[string]interface{}{
		"verificationId":   verificationContext.VerificationId,
		"verificationType": verificationContext.VerificationType,
		"maxImageSize":     cfg.MaxImageSize,
	})

	// Fetch all data in parallel (images with Base64 encoding, DynamoDB context)
	results := ParallelFetch(
		ctx,
		deps,
		cfg, // Pass configuration for image size limits
		verificationContext.ReferenceImageUrl,
		verificationContext.CheckingImageUrl,
		verificationContext.LayoutId,
		verificationContext.LayoutPrefix,
		prevVerificationId,
	)

	// Check for errors from parallel processing
	if len(results.Errors) > 0 {
		// Log all errors but return the first one
		for _, fetchErr := range results.Errors {
			logger.Error("Error during parallel fetch", map[string]interface{}{
				"error": fetchErr.Error(),
			})
		}
		return FetchImagesResponse{}, NewNotFoundError("Failed to fetch required resources", results.Errors[0])
	}

	// Validate that we have Base64 data for both images
	if !results.ReferenceMeta.HasBase64Data() {
		return FetchImagesResponse{}, fmt.Errorf("reference image missing Base64 data")
	}
	if !results.CheckingMeta.HasBase64Data() {
		return FetchImagesResponse{}, fmt.Errorf("checking image missing Base64 data")
	}

	// Construct response with complete verification context and Base64 images
	resp := FetchImagesResponse{
		VerificationContext: verificationContext,
		Images: ImagesData{
			ReferenceImageMeta: results.ReferenceMeta,
			CheckingImageMeta:  results.CheckingMeta,
		},
		LayoutMetadata:    results.LayoutMeta,
		HistoricalContext: results.HistoricalContext,
	}
	
	// Ensure HistoricalContext is never nil to satisfy Step Functions JSONPath
	if resp.HistoricalContext == nil {
		// Create an empty historical context to ensure field exists in JSON
		resp.HistoricalContext = map[string]interface{}{
			"previousVerificationId":      "",
			"previousVerificationAt":      "",
			"previousVerificationStatus":  "",
			"hoursSinceLastVerification":  0,
		}
	}

	// Optionally update status in DynamoDB (commented out as per original)
	// dbWrapper := NewDBUtils(deps.GetDbUtils())
	// dbWrapper.UpdateVerificationStatus(ctx, verificationContext.VerificationId, string(schema.StatusImagesFetched))

	logger.Info("Successfully processed images with Base64 encoding", map[string]interface{}{
		"verificationId":       verificationContext.VerificationId,
		"verificationType":     verificationContext.VerificationType,
		"referenceImageSize":   results.ReferenceMeta.Size,
		"checkingImageSize":    results.CheckingMeta.Size,
		"referenceBase64Size":  len(results.ReferenceMeta.Base64Data),
		"checkingBase64Size":   len(results.CheckingMeta.Base64Data),
		"referenceFormat":      results.ReferenceMeta.ImageFormat,
		"checkingFormat":       results.CheckingMeta.ImageFormat,
		"layoutMetadataFound":  len(results.LayoutMeta) > 0,
		"historicalDataFound":  len(results.HistoricalContext) > 0,
	})

	return resp, nil
}

// initDependencies initializes all required dependencies
func initDependencies(ctx context.Context) (*Dependencies, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return NewDependencies(awsCfg), nil
}

func main() {
	lambda.Start(Handler)
}