package main

import (
	"context"
	"encoding/json"
	"fmt"
	"workflow-function/shared/logger"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"workflow-function/shared/schema"
)

// Handler is the Lambda handler function with hybrid Base64 storage support
func Handler(ctx context.Context, event interface{}) (FetchImagesResponse, error) {
	// Initialize dependencies
	deps, err := initDependencies(ctx)
	if err != nil {
		return FetchImagesResponse{}, fmt.Errorf("failed to initialize dependencies: %w", err)
	}

	// Load configuration early for hybrid storage
	cfg := deps.GetConfig()
	
	// Configure dependencies with environment variables
	deps.ConfigureDbUtils(cfg)

	logger := deps.GetLogger()

	// Validate storage configuration for hybrid Base64 storage
	if err := cfg.ValidateConfig(); err != nil {
		logger.Warn("Storage configuration validation issues", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Log storage configuration for debugging
	logger.Info("Hybrid storage configuration", cfg.GetStorageConfig())

	// Parse input based on event type
	var req FetchImagesRequest
	var parseErr error

	// Log the incoming event for debugging with storage info
	eventBytes, _ := json.Marshal(event)
	logger.Info("Received event for hybrid Base64 image processing", map[string]interface{}{
		"eventType":            fmt.Sprintf("%T", event),
		"maxImageSize":         cfg.MaxImageSize,
		"maxInlineBase64Size":  cfg.MaxInlineBase64Size,
		"tempBase64Bucket":     cfg.TempBase64Bucket,
		"storageConfig":        cfg.GetStorageConfig(),
		"eventBody":            string(eventBytes),
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

	logger.Info("Starting parallel fetch with hybrid Base64 encoding", map[string]interface{}{
		"verificationId":       verificationContext.VerificationId,
		"verificationType":     verificationContext.VerificationType,
		"maxImageSize":         cfg.MaxImageSize,
		"maxInlineBase64Size":  cfg.MaxInlineBase64Size,
		"tempBase64Bucket":     cfg.TempBase64Bucket,
		"hybridStorageEnabled": cfg.TempBase64Bucket != "",
	})

	// Fetch all data in parallel with hybrid Base64 encoding
	results := ParallelFetch(
		ctx,
		deps,
		cfg, // Pass configuration for hybrid storage
		verificationContext.ReferenceImageUrl,
		verificationContext.CheckingImageUrl,
		verificationContext.LayoutId,
		verificationContext.LayoutPrefix,
		prevVerificationId,
	)

	// Check for errors from parallel processing
	if len(results.Errors) > 0 {
		// Log all errors but return the first one
		for i, fetchErr := range results.Errors {
			logger.Error("Error during parallel fetch", map[string]interface{}{
				"errorIndex": i,
				"error":      fetchErr.Error(),
			})
		}
		return FetchImagesResponse{}, NewNotFoundError("Failed to fetch required resources", results.Errors[0])
	}

	// Validate hybrid storage integrity for both images
	if validationErr := validateHybridStorageResults(results, logger); validationErr != nil {
		return FetchImagesResponse{}, validationErr
	}

	// Log storage method summary
	logStorageMethodSummary(results, logger)

	// Construct response with complete verification context and hybrid Base64 storage
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
	// dbClient := deps.GetDBClient()
	// verificationTable := deps.GetVerificationTable()
	// UpdateVerificationStatus(ctx, dbClient, verificationTable, verificationContext.VerificationId, string(schema.StatusImagesFetched))

	// Log successful completion with hybrid storage details
	logger.Info("Successfully processed images with hybrid Base64 encoding", map[string]interface{}{
		"verificationId":            verificationContext.VerificationId,
		"verificationType":          verificationContext.VerificationType,
		"referenceImageSize":        results.ReferenceMeta.Size,
		"checkingImageSize":         results.CheckingMeta.Size,
		"referenceStorageMethod":    results.ReferenceMeta.StorageMethod,
		"checkingStorageMethod":     results.CheckingMeta.StorageMethod,
		"referenceFormat":           results.ReferenceMeta.ImageFormat,
		"checkingFormat":            results.CheckingMeta.ImageFormat,
		"referenceBase64Available":  results.ReferenceMeta.HasBase64Data(),
		"checkingBase64Available":   results.CheckingMeta.HasBase64Data(),
		"layoutMetadataFound":       len(results.LayoutMeta) > 0,
		"historicalDataFound":       len(results.HistoricalContext) > 0,
		"referenceStorageInfo":      results.ReferenceMeta.GetStorageInfo(),
		"checkingStorageInfo":       results.CheckingMeta.GetStorageInfo(),
		"hybridStorageUsed":         isHybridStorageUsed(results),
		"bothImagesInline":          areBothImagesInline(results),
		"anyImageInS3Temp":          isAnyImageInS3Temp(results),
	})

	return resp, nil
}

// validateHybridStorageResults validates that both images have proper Base64 data with hybrid storage
func validateHybridStorageResults(results ParallelFetchResults, logger logger.Logger) error {
	// Validate reference image Base64 availability
	if !results.ReferenceMeta.HasBase64Data() {
		err := fmt.Errorf("reference image missing Base64 data (storage method: %s)", 
			results.ReferenceMeta.StorageMethod)
		logger.Error("Reference image Base64 validation failed", map[string]interface{}{
			"storageMethod":  results.ReferenceMeta.StorageMethod,
			"hasInlineData":  results.ReferenceMeta.HasInlineData(),
			"hasS3Storage":   results.ReferenceMeta.HasS3Storage(),
			"storageInfo":    results.ReferenceMeta.GetStorageInfo(),
		})
		return err
	}

	// Validate checking image Base64 availability
	if !results.CheckingMeta.HasBase64Data() {
		err := fmt.Errorf("checking image missing Base64 data (storage method: %s)", 
			results.CheckingMeta.StorageMethod)
		logger.Error("Checking image Base64 validation failed", map[string]interface{}{
			"storageMethod":  results.CheckingMeta.StorageMethod,
			"hasInlineData":  results.CheckingMeta.HasInlineData(),
			"hasS3Storage":   results.CheckingMeta.HasS3Storage(),
			"storageInfo":    results.CheckingMeta.GetStorageInfo(),
		})
		return err
	}

	// Validate storage method consistency
	if err := results.ReferenceMeta.ValidateStorageMethod(); err != nil {
		logger.Error("Reference image storage method validation failed", map[string]interface{}{
			"error": err.Error(),
			"storageMethod": results.ReferenceMeta.StorageMethod,
		})
		return fmt.Errorf("reference image storage validation failed: %w", err)
	}

	if err := results.CheckingMeta.ValidateStorageMethod(); err != nil {
		logger.Error("Checking image storage method validation failed", map[string]interface{}{
			"error": err.Error(),
			"storageMethod": results.CheckingMeta.StorageMethod,
		})
		return fmt.Errorf("checking image storage validation failed: %w", err)
	}

	logger.Info("Hybrid storage validation passed", map[string]interface{}{
		"referenceStorageMethod": results.ReferenceMeta.StorageMethod,
		"checkingStorageMethod":  results.CheckingMeta.StorageMethod,
		"bothImagesHaveBase64":   true,
	})

	return nil
}

// logStorageMethodSummary logs a summary of storage methods used
func logStorageMethodSummary(results ParallelFetchResults, logger logger.Logger) {
	summary := map[string]interface{}{
		"totalImages": 2,
		"storageMethodDistribution": map[string]int{
			StorageMethodInline:      0,
			StorageMethodS3Temporary: 0,
		},
		"images": []map[string]interface{}{
			{
				"type":          "reference",
				"storageMethod": results.ReferenceMeta.StorageMethod,
				"size":          results.ReferenceMeta.Size,
				"format":        results.ReferenceMeta.ImageFormat,
			},
			{
				"type":          "checking",
				"storageMethod": results.CheckingMeta.StorageMethod,
				"size":          results.CheckingMeta.Size,
				"format":        results.CheckingMeta.ImageFormat,
			},
		},
	}

	// Count storage methods
	if results.ReferenceMeta.StorageMethod == StorageMethodInline {
		summary["storageMethodDistribution"].(map[string]int)[StorageMethodInline]++
	} else if results.ReferenceMeta.StorageMethod == StorageMethodS3Temporary {
		summary["storageMethodDistribution"].(map[string]int)[StorageMethodS3Temporary]++
	}

	if results.CheckingMeta.StorageMethod == StorageMethodInline {
		summary["storageMethodDistribution"].(map[string]int)[StorageMethodInline]++
	} else if results.CheckingMeta.StorageMethod == StorageMethodS3Temporary {
		summary["storageMethodDistribution"].(map[string]int)[StorageMethodS3Temporary]++
	}

	// Add efficiency metrics
	summary["hybridStorageEfficiency"] = map[string]interface{}{
		"hybridStorageUsed":     isHybridStorageUsed(results),
		"bothImagesInline":      areBothImagesInline(results),
		"anyImageInS3Temp":      isAnyImageInS3Temp(results),
		"s3TemporaryPercentage": float64(summary["storageMethodDistribution"].(map[string]int)[StorageMethodS3Temporary]) / 2.0 * 100,
	}

	logger.Info("Storage method summary", summary)
}

// isHybridStorageUsed checks if different storage methods were used for the images
func isHybridStorageUsed(results ParallelFetchResults) bool {
	return results.ReferenceMeta.StorageMethod != results.CheckingMeta.StorageMethod
}

// areBothImagesInline checks if both images are stored inline
func areBothImagesInline(results ParallelFetchResults) bool {
	return results.ReferenceMeta.StorageMethod == StorageMethodInline &&
		   results.CheckingMeta.StorageMethod == StorageMethodInline
}

// isAnyImageInS3Temp checks if any image is stored in S3 temporary storage
func isAnyImageInS3Temp(results ParallelFetchResults) bool {
	return results.ReferenceMeta.StorageMethod == StorageMethodS3Temporary ||
		   results.CheckingMeta.StorageMethod == StorageMethodS3Temporary
}

// validateStorageConfiguration validates the storage configuration at runtime
func validateStorageConfiguration(cfg ConfigVars, logger logger.Logger) error {
	errors := []string{}

	// Check if temp bucket is configured for production workloads
	if cfg.TempBase64Bucket == "" {
		logger.Warn("No temporary Base64 bucket configured", map[string]interface{}{
			"impact": "Large images (>2MB Base64) will fail to process",
			"recommendation": "Set TEMP_BASE64_BUCKET environment variable",
		})
	}

	// Validate size thresholds
	if cfg.MaxInlineBase64Size > cfg.MaxImageSize {
		errors = append(errors, "MaxInlineBase64Size is larger than MaxImageSize")
	}

	// Validate bucket accessibility (if configured)
	if cfg.TempBase64Bucket != "" {
		// Note: Actual bucket validation would require S3 operations
		// For now, we just log the configuration
		logger.Info("Temporary Base64 bucket configured", map[string]interface{}{
			"bucket": cfg.TempBase64Bucket,
			"note":   "Bucket accessibility will be validated during actual operations",
		})
	}

	if len(errors) > 0 {
		errMsg := fmt.Sprintf("Storage configuration issues: %v", errors)
		logger.Error("Storage configuration validation failed", map[string]interface{}{
			"errors": errors,
		})
		return fmt.Errorf(errMsg)
	}

	return nil
}

// initDependencies initializes all required dependencies
func initDependencies(ctx context.Context) (*Dependencies, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	
	deps := NewDependencies(awsCfg)
	
	// Validate storage configuration
	if err := validateStorageConfiguration(deps.GetConfig(), deps.GetLogger()); err != nil {
		deps.GetLogger().Warn("Storage configuration validation issues", map[string]interface{}{
			"error": err.Error(),
		})
		// Don't fail initialization for configuration issues, just warn
	}
	
	return deps, nil
}

// logEnvironmentInfo logs environment information for debugging
func logEnvironmentInfo(logger logger.Logger) {
	envInfo := GetEnvironmentInfo()
	logger.Info("Environment information", envInfo)
}

func main() {
	lambda.Start(Handler)
}
