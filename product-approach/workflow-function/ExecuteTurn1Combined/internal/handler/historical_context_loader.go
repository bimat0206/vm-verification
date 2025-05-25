package handler

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"workflow-function/ExecuteTurn1Combined/internal/models"
	"workflow-function/ExecuteTurn1Combined/internal/services"
	"workflow-function/shared/logger"
	"workflow-function/shared/schema"
)

// HistoricalContextLoader handles loading historical verification context
type HistoricalContextLoader struct {
	dynamo services.DynamoDBService
	log    logger.Logger
}

// NewHistoricalContextLoader creates a new instance of HistoricalContextLoader
func NewHistoricalContextLoader(dynamo services.DynamoDBService, log logger.Logger) *HistoricalContextLoader {
	return &HistoricalContextLoader{
		dynamo: dynamo,
		log:    log,
	}
}

// LoadHistoricalContext loads historical context for PREVIOUS_VS_CURRENT verification type
func (h *HistoricalContextLoader) LoadHistoricalContext(ctx context.Context, req *models.Turn1Request, contextLogger logger.Logger) (time.Duration, error) {
	if req.VerificationContext.VerificationType != schema.VerificationTypePreviousVsCurrent {
		return 0, nil
	}

	historicalStart := time.Now()

	// Extract checking image URL from the reference image S3 key
	checkingImageUrl := extractCheckingImageUrl(req.S3Refs.Images.ReferenceBase64.Key)

	if checkingImageUrl == "" {
		return time.Since(historicalStart), nil
	}

	contextLogger.Info("Loading historical verification context", map[string]interface{}{
		"checking_image_url": checkingImageUrl,
		"verification_type":  req.VerificationContext.VerificationType,
	})

	// Query for previous verification using the checking image URL
	previousVerification, err := h.dynamo.QueryPreviousVerification(ctx, checkingImageUrl)
	if err != nil {
		// Log warning but continue - historical context is optional enhancement
		contextLogger.Warn("Failed to load historical verification context", map[string]interface{}{
			"error":              err.Error(),
			"checking_image_url": checkingImageUrl,
		})
		return time.Since(historicalStart), nil
	}

	if previousVerification == nil {
		return time.Since(historicalStart), nil
	}

	// Populate historical context with previous verification data
	req.VerificationContext.HistoricalContext = map[string]interface{}{
		"PreviousVerificationAt":     previousVerification.VerificationAt,
		"PreviousVerificationStatus": previousVerification.CurrentStatus,
		"PreviousVerificationId":     previousVerification.VerificationId,
		"HoursSinceLastVerification": calculateHoursSince(previousVerification.VerificationAt),
	}

	// Add layout information from the previous verification if present
	if previousVerification.LayoutId > 0 {
		req.VerificationContext.HistoricalContext["LayoutId"] = previousVerification.LayoutId
	}
	if previousVerification.LayoutPrefix != "" {
		req.VerificationContext.HistoricalContext["LayoutPrefix"] = previousVerification.LayoutPrefix
	}

	var machineStructureFound bool

	// Attempt to extract machine structure directly from previous verification via reflection
	pvValue := reflect.ValueOf(previousVerification).Elem()
	msField := pvValue.FieldByName("MachineStructure")
	if msField.IsValid() {
		if msMap, ok := msField.Interface().(map[string]interface{}); ok && msMap != nil {
			if rc, ok := msMap["rowCount"].(float64); ok {
				req.VerificationContext.HistoricalContext["RowCount"] = int(rc)
				machineStructureFound = true
			}
			if cc, ok := msMap["columnCount"].(float64); ok {
				req.VerificationContext.HistoricalContext["ColumnCount"] = int(cc)
			}
			if rl, ok := msMap["rowOrder"].([]interface{}); ok {
				rowLabels := make([]string, len(rl))
				for i, v := range rl {
					rowLabels[i] = fmt.Sprintf("%v", v)
				}
				req.VerificationContext.HistoricalContext["RowLabels"] = rowLabels
			}
			if machineStructureFound {
				contextLogger.Info("Populated historical machine structure from previous verification", nil)
			}
		}
	}

	// If not found directly, use layout metadata lookup
	if !machineStructureFound && previousVerification.LayoutId > 0 && previousVerification.LayoutPrefix != "" {
		contextLogger.Info("Attempting to load LayoutMetadata for historical context", map[string]interface{}{
			"layoutId":     previousVerification.LayoutId,
			"layoutPrefix": previousVerification.LayoutPrefix,
		})
		layoutMeta, err := h.dynamo.GetLayoutMetadata(ctx, previousVerification.LayoutId, previousVerification.LayoutPrefix)
		if err == nil && layoutMeta != nil && layoutMeta.MachineStructure != nil {
			if rc, ok := layoutMeta.MachineStructure["rowCount"].(float64); ok {
				req.VerificationContext.HistoricalContext["RowCount"] = int(rc)
				machineStructureFound = true
			}
			if cc, ok := layoutMeta.MachineStructure["columnCount"].(float64); ok {
				req.VerificationContext.HistoricalContext["ColumnCount"] = int(cc)
			}
			if rl, ok := layoutMeta.MachineStructure["rowOrder"].([]interface{}); ok {
				rowLabels := make([]string, len(rl))
				for i, v := range rl {
					rowLabels[i] = fmt.Sprintf("%v", v)
				}
				req.VerificationContext.HistoricalContext["RowLabels"] = rowLabels
			}
			if machineStructureFound {
				contextLogger.Info("Successfully populated historical machine structure from LayoutMetadata", map[string]interface{}{"layoutId": layoutMeta.LayoutId})
			}
		} else if err != nil {
			contextLogger.Warn("Failed to load LayoutMetadata for historical context", map[string]interface{}{
				"error":        err.Error(),
				"layoutId":     previousVerification.LayoutId,
				"layoutPrefix": previousVerification.LayoutPrefix,
			})
		}
	}

	if !machineStructureFound {
		contextLogger.Warn("Machine structure could not be determined for historical context", nil)
	}

	contextLogger.Info("Successfully loaded historical verification context", map[string]interface{}{
		"previous_verification_id": previousVerification.VerificationId,
		"previous_verification_at": previousVerification.VerificationAt,
		"hours_since_last":         req.VerificationContext.HistoricalContext["HoursSinceLastVerification"],
	})

	return time.Since(historicalStart), nil
}
