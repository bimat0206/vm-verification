package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/FinalizeAndStoreResults/internal/config"
	"workflow-function/FinalizeAndStoreResults/internal/dynamodbhelper"
	"workflow-function/FinalizeAndStoreResults/internal/models"
	"workflow-function/FinalizeAndStoreResults/internal/parser"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
	"workflow-function/shared/schema"
)

var (
	appConfig    *config.LambdaConfig
	awsClients   *config.AWSClients
	log          logger.Logger
	stateManager s3state.Manager
)

// extractNestedReference extracts a reference from nested JSON structure
// This handles cases where the input has nested s3References like:
// {"s3References": {"responses": {"turn2Processed": {...}}}}
func extractNestedReference(input interface{}, category, refName string) *s3state.Reference {
	// Convert input to map[string]interface{}
	var inputMap map[string]interface{}

	switch v := input.(type) {
	case map[string]interface{}:
		inputMap = v
	case []byte:
		if err := json.Unmarshal(v, &inputMap); err != nil {
			return nil
		}
	case string:
		if err := json.Unmarshal([]byte(v), &inputMap); err != nil {
			return nil
		}
	default:
		return nil
	}

	// Navigate to s3References
	s3Refs, ok := inputMap["s3References"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Navigate to category (e.g., "responses")
	categoryRefs, ok := s3Refs[category].(map[string]interface{})
	if !ok {
		return nil
	}

	// Get the specific reference (e.g., "turn2Processed")
	refData, ok := categoryRefs[refName].(map[string]interface{})
	if !ok {
		return nil
	}

	// Extract bucket, key, and size
	bucket, _ := refData["bucket"].(string)
	key, _ := refData["key"].(string)
	var size int64
	if sizeVal, ok := refData["size"]; ok {
		switch s := sizeVal.(type) {
		case float64:
			size = int64(s)
		case int64:
			size = s
		case int:
			size = int64(s)
		}
	}

	if bucket == "" || key == "" {
		return nil
	}

	return &s3state.Reference{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	}
}

// parseInputAndExtractReferences parses the input and extracts required references
func parseInputAndExtractReferences(input interface{}) (*s3state.Envelope, *s3state.Reference, *s3state.Reference, error) {
	// Convert input to map[string]interface{}
	var inputMap map[string]interface{}

	switch v := input.(type) {
	case map[string]interface{}:
		inputMap = v
	case []byte:
		if err := json.Unmarshal(v, &inputMap); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to unmarshal input bytes: %w", err)
		}
	case string:
		if err := json.Unmarshal([]byte(v), &inputMap); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to unmarshal input string: %w", err)
		}
	default:
		return nil, nil, nil, fmt.Errorf("unsupported input type: %T", input)
	}

	// Extract basic envelope fields
	verificationID, _ := inputMap["verificationId"].(string)
	status, _ := inputMap["status"].(string)
	if verificationID == "" {
		return nil, nil, nil, fmt.Errorf("verificationId is required")
	}

	// Navigate to s3References
	s3Refs, ok := inputMap["s3References"].(map[string]interface{})
	if !ok {
		return nil, nil, nil, fmt.Errorf("s3References not found or invalid")
	}

	// Extract processing_initialization reference (top-level)
	initRef := extractReferenceFromMap(s3Refs, "processing_initialization")
	if initRef == nil {
		return nil, nil, nil, fmt.Errorf("processing_initialization reference not found")
	}

	// Extract turn2Processed reference - try both nested and flat structures
	turn2Ref := extractNestedReference(input, "responses", "turn2Processed")
	if turn2Ref == nil {
		// Try flat structure for backward compatibility
		turn2Ref = extractReferenceFromMap(s3Refs, "turn2Processed")
		if turn2Ref == nil {
			return nil, nil, nil, fmt.Errorf("turn2Processed reference not found in both nested responses and flat structure")
		}
	}

	// Create envelope with flat references structure for compatibility
	envelope := &s3state.Envelope{
		VerificationID: verificationID,
		Status:         status,
		References:     make(map[string]*s3state.Reference),
		Summary:        make(map[string]interface{}),
	}

	// Add references to envelope for compatibility with existing code
	envelope.References["processing_initialization"] = initRef
	envelope.References["turn2Processed"] = turn2Ref

	return envelope, initRef, turn2Ref, nil
}

// extractReferenceFromMap extracts a reference from a map
func extractReferenceFromMap(refMap map[string]interface{}, refName string) *s3state.Reference {
	refData, ok := refMap[refName].(map[string]interface{})
	if !ok {
		return nil
	}

	bucket, _ := refData["bucket"].(string)
	key, _ := refData["key"].(string)
	var size int64
	if sizeVal, ok := refData["size"]; ok {
		switch s := sizeVal.(type) {
		case float64:
			size = int64(s)
		case int64:
			size = s
		case int:
			size = int64(s)
		}
	}

	if bucket == "" || key == "" {
		return nil
	}

	return &s3state.Reference{
		Bucket: bucket,
		Key:    key,
		Size:   size,
	}
}

func init() {
	var err error
	log = logger.New("FinalizeAndStoreResults", "main")

	appConfig, err = config.LoadEnvConfig()
	if err != nil {
		log.Error("failed_to_load_config", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	awsClients, err = config.NewAWSClients(context.Background())
	if err != nil {
		log.Error("failed_to_init_aws", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	stateManager, err = s3state.New(appConfig.StateBucket)
	if err != nil {
		log.Error("failed_to_init_s3state", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

func HandleRequest(ctx context.Context, input interface{}) (*s3state.Envelope, error) {
	// Parse input to extract references and create envelope
	envelope, initRef, turn2Ref, err := parseInputAndExtractReferences(input)
	if err != nil {
		wfErr := errors.NewValidationError("failed to parse input", map[string]interface{}{
			"error": err.Error(),
		})
		log.Error("input_parse_failed", map[string]interface{}{"error": wfErr.Error()})
		return nil, wfErr
	}

	log.LogReceivedEvent(envelope)

	log.Info("extracted_references", map[string]interface{}{
		"init_bucket":  initRef.Bucket,
		"init_key":     initRef.Key,
		"turn2_bucket": turn2Ref.Bucket,
		"turn2_key":    turn2Ref.Key,
		"turn2_size":   turn2Ref.Size,
	})

	// Load initialization data using s3state manager
	var initData models.InitializationData
	err = stateManager.RetrieveJSON(initRef, &initData)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeS3, "failed to load initialization data", false)
		wfErr.VerificationID = envelope.VerificationID
		log.Error("fetch_init_failed", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
			"bucket":         initRef.Bucket,
			"key":            initRef.Key,
		})
		return nil, wfErr
	}

	// Validate that we have the verification context
	if initData.VerificationContext == nil {
		wfErr := errors.NewValidationError("verificationContext is missing from initialization data", map[string]interface{}{
			"schemaVersion": initData.SchemaVersion,
		})
		wfErr.VerificationID = envelope.VerificationID
		log.Error("missing_verification_context", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
			"schemaVersion":  initData.SchemaVersion,
		})
		return nil, wfErr
	}

	// Log initialization data for debugging (sanitized)
	log.Info("initialization_data_loaded", map[string]interface{}{
		"schemaVersion":          initData.SchemaVersion,
		"verificationId":         initData.VerificationContext.VerificationID,
		"verificationType":       initData.VerificationContext.VerificationType,
		"layoutId":               initData.VerificationContext.LayoutID,
		"layoutPrefix":           initData.VerificationContext.LayoutPrefix,
		"vendingMachineId":       initData.VerificationContext.VendingMachineID,
		"hasReferenceImageUrl":   initData.VerificationContext.ReferenceImageUrl != "",
		"hasCheckingImageUrl":    initData.VerificationContext.CheckingImageUrl != "",
		"status":                 initData.VerificationContext.Status,
		"lastUpdatedAt":          initData.VerificationContext.LastUpdatedAt,
		"previousVerificationId": initData.VerificationContext.PreviousVerificationID,
	})

	// Load turn2 processed data using s3state manager
	turn2Bytes, err := stateManager.Retrieve(turn2Ref)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeS3, "failed to load turn2 processed data", false)
		wfErr.VerificationID = envelope.VerificationID
		log.Error("fetch_turn2_failed", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
			"bucket":         turn2Ref.Bucket,
			"key":            turn2Ref.Key,
		})
		return nil, wfErr
	}

	// Parse turn2 response data
	parsed, err := parser.ParseTurn2ResponseData(turn2Bytes)
	if err != nil {
		wfErr := errors.NewParsingError("turn2 response data", err)
		wfErr.VerificationID = envelope.VerificationID
		log.Error("parse_turn2_failed", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
		})
		return nil, wfErr
	}

	// Create DynamoDB item using shared schema constants
	now := schema.FormatISO8601()

	// Ensure verificationType is set - default to LAYOUT_VS_CHECKING if missing
	verificationType := initData.VerificationContext.VerificationType
	if verificationType == "" {
		verificationType = schema.VerificationTypeLayoutVsChecking
		log.Info("verification_type_defaulted", map[string]interface{}{
			"verificationId": envelope.VerificationID,
			"defaultType":    verificationType,
			"reason":         "verificationType was empty in verification context",
		})
	}

	// Ensure verificationStatus is set - this should contain the actual AI verification result (CORRECT/INCORRECT)
	// parsed from the Turn2 response. If parsing fails, default to SUCCESSED as a fallback.
	// This is required for the DynamoDB VerificationStatusIndex (cannot be empty string).
	verificationStatus := parsed.VerificationStatus
	if verificationStatus == "" {
		// SUCCESS indicates workflow completed but couldn't determine actual verification result
		// This is different from CORRECT (AI determined layout was correct) or INCORRECT (AI found issues)
		verificationStatus = "SUCCESS"
		log.Warn("verification_status_defaulted", map[string]interface{}{
			"verificationId": envelope.VerificationID,
			"defaultStatus":  verificationStatus,
			"reason":         "Could not parse actual verification status from Turn2 response - using SUCCESS as fallback since workflow completed",
		})
	}

	item := models.VerificationResultItem{
		VerificationID:         envelope.VerificationID,
		VerificationAt:         initData.VerificationContext.VerificationAt, // Use existing verificationAt to update the correct record
		VerificationType:       verificationType,
		LayoutID:               &initData.VerificationContext.LayoutID,
		LayoutPrefix:           initData.VerificationContext.LayoutPrefix,
		VendingMachineID:       initData.VerificationContext.VendingMachineID,
		ReferenceImageUrl:      initData.VerificationContext.ReferenceImageUrl,
		CheckingImageUrl:       initData.VerificationContext.CheckingImageUrl,
		VerificationStatus:     verificationStatus,
		CurrentStatus:          schema.StatusCompleted,
		LastUpdatedAt:          now,
		ProcessingStartedAt:    initData.VerificationContext.VerificationAt, // Use verificationAt as processing start
		InitialConfirmation:    parsed.InitialConfirmation,
		VerificationSummary:    parsed.VerificationSummary,
		PreviousVerificationID: initData.VerificationContext.PreviousVerificationID,
	}
	if initData.VerificationContext.LayoutID == 0 {
		item.LayoutID = nil
	}

	// Store verification result in DynamoDB with enhanced error handling
	err = dynamodbhelper.StoreVerificationResult(ctx, awsClients.DynamoDBClient, appConfig.VerificationResultsTable, item, log)
	if err != nil {
		// The enhanced helper already provides detailed error information and logging
		return nil, err
	}

	// Update conversation history in DynamoDB with enhanced error handling
	err = dynamodbhelper.UpdateConversationHistory(ctx, awsClients.DynamoDBClient, appConfig.ConversationHistoryTable, envelope.VerificationID, initData.VerificationContext.VerificationAt, log)
	if err != nil {
		// The enhanced helper already provides detailed error information and logging
		return nil, err
	}

	// Store verificationSummary as JSON file in S3
	summaryRef, err := storeVerificationSummaryJSON(envelope.VerificationID, parsed.VerificationSummary, log)
	if err != nil {
		wfErr := errors.WrapError(err, errors.ErrorTypeS3, "failed to store verification summary JSON", false)
		wfErr.VerificationID = envelope.VerificationID
		log.Error("store_summary_json_failed", map[string]interface{}{
			"error":          wfErr.Error(),
			"verificationId": envelope.VerificationID,
		})
		return nil, wfErr
	}

	// Update envelope status and add summary
	envelope.SetStatus(schema.StatusCompleted)
	if envelope.Summary == nil {
		envelope.Summary = make(map[string]interface{})
	}
	envelope.Summary["verificationStatus"] = verificationStatus // Use the corrected verificationStatus
	envelope.Summary["verificationAt"] = now
	envelope.Summary["message"] = "Verification finalized and stored"

	// Add results reference to envelope
	if envelope.References == nil {
		envelope.References = make(map[string]*s3state.Reference)
	}
	envelope.References["results"] = summaryRef

	log.LogOutputEvent(envelope)
	return envelope, nil
}

// storeVerificationSummaryJSON stores the verification summary as a JSON file in S3
func storeVerificationSummaryJSON(verificationID string, summary models.OutputVerificationSummary, log logger.Logger) (*s3state.Reference, error) {
	// Create the date path for the verification ID (e.g., "2025/06/05/verif-20250605074028-f5c4")
	datePath := extractDatePathFromVerificationID(verificationID)

	// Create the S3 key for the verification summary JSON
	key := "results/verificationSummary.json"

	log.Info("storing_verification_summary_json", map[string]interface{}{
		"verificationId": verificationID,
		"datePath":       datePath,
		"key":            key,
	})

	// Store the summary as JSON using the s3state manager
	stateRef, err := stateManager.StoreJSON(datePath, key, summary)
	if err != nil {
		return nil, err
	}

	log.Info("verification_summary_json_stored", map[string]interface{}{
		"verificationId": verificationID,
		"bucket":         stateRef.Bucket,
		"key":            stateRef.Key,
		"size":           stateRef.Size,
	})

	return stateRef, nil
}

// extractDatePathFromVerificationID extracts the date path from verification ID
// e.g., "verif-20250605074028-f5c4" -> "2025/06/05/verif-20250605074028-f5c4"
func extractDatePathFromVerificationID(verificationID string) string {
	// Extract date from verification ID format: verif-YYYYMMDDHHMMSS-xxxx
	if len(verificationID) >= 21 && verificationID[:6] == "verif-" {
		dateStr := verificationID[6:14] // Extract YYYYMMDD
		if len(dateStr) == 8 {
			year := dateStr[:4]
			month := dateStr[4:6]
			day := dateStr[6:8]
			return fmt.Sprintf("%s/%s/%s/%s", year, month, day, verificationID)
		}
	}

	// Fallback: use current date if parsing fails
	now := schema.FormatISO8601()
	dateStr := now[:10] // Extract YYYY-MM-DD
	dateParts := strings.Split(dateStr, "-")
	if len(dateParts) == 3 {
		return fmt.Sprintf("%s/%s/%s/%s", dateParts[0], dateParts[1], dateParts[2], verificationID)
	}

	// Final fallback
	return fmt.Sprintf("unknown/%s", verificationID)
}

func main() {
	lambda.Start(HandleRequest)
}
