package internal

import (
   "context"
   "encoding/base64"
   "fmt"
   "io"
   "strings"
   "time"

   "github.com/aws/aws-sdk-go-v2/config"
   "github.com/aws/aws-sdk-go-v2/service/s3"

   "workflow-function/shared/errors"
   "workflow-function/shared/schema"
)

// ProcessImageForBedrock retrieves and prepares Base64 data for Bedrock
func ProcessImageForBedrock(imageInfo *schema.ImageInfo) error {
   if imageInfo == nil {
   	return errors.NewValidationError("Image info is required", nil)
   }

   // If Base64 data already exists, validate and return
   if imageInfo.Base64Data != "" {
   	return validateExistingBase64Data(imageInfo)
   }

   // Determine how to retrieve Base64 data based on storage method
   switch imageInfo.StorageMethod {
   case schema.StorageMethodInline:
   	return errors.NewValidationError("Inline storage specified but no Base64 data found", 
   		map[string]interface{}{"storageMethod": imageInfo.StorageMethod})
   
   case schema.StorageMethodS3Temporary:
   	return retrieveBase64FromS3Temp(imageInfo)
   
   default:
   	// Default: download from regular S3 bucket and encode
   	return downloadAndEncodeFromS3(imageInfo)
   }
}

// validateExistingBase64Data validates existing Base64 data
func validateExistingBase64Data(imageInfo *schema.ImageInfo) error {
   // Basic Base64 validation
   if !isValidBase64(imageInfo.Base64Data) {
   	return errors.NewValidationError("Invalid Base64 data format", nil)
   }

   // Decode to check if it's a valid image (basic check)
   decoded, err := base64.StdEncoding.DecodeString(imageInfo.Base64Data)
   if err != nil {
   	return errors.NewValidationError("Failed to decode Base64 data", 
   		map[string]interface{}{"error": err.Error()})
   }

   // Validate image size (max 10MB for Bedrock)
   if len(decoded) > 10*1024*1024 {
   	return errors.NewValidationError("Image size exceeds Bedrock limit", 
   		map[string]interface{}{"sizeMB": len(decoded)/(1024*1024)})
   }

   // Set format if not already set
   if imageInfo.Format == "" {
   	imageInfo.Format = detectImageFormatFromHeader(decoded)
   }

   return nil
}

// retrieveBase64FromS3Temp retrieves Base64 data from S3 temporary storage
func retrieveBase64FromS3Temp(imageInfo *schema.ImageInfo) error {
   if imageInfo.Base64S3Bucket == "" || imageInfo.Base64S3Key == "" {
   	return errors.NewValidationError("S3 temporary storage info missing", 
   		map[string]interface{}{
   			"bucket": imageInfo.Base64S3Bucket,
   			"key": imageInfo.Base64S3Key,
   		})
   }

   // Create S3 client
   s3Client, err := createS3Client()
   if err != nil {
   	return errors.NewInternalError("s3-client-creation", err)
   }

   // Download Base64 string from S3
   base64Data, err := downloadS3ObjectAsString(s3Client, imageInfo.Base64S3Bucket, imageInfo.Base64S3Key)
   if err != nil {
   	return errors.NewInternalError("s3-temp-download", err)
   }

   // Validate the retrieved Base64 data
   imageInfo.Base64Data = base64Data
   return validateExistingBase64Data(imageInfo)
}

// downloadAndEncodeFromS3 downloads image from S3 and encodes to Base64
func downloadAndEncodeFromS3(imageInfo *schema.ImageInfo) error {
   if imageInfo.S3Bucket == "" || imageInfo.S3Key == "" {
   	return errors.NewValidationError("S3 storage info missing", 
   		map[string]interface{}{
   			"bucket": imageInfo.S3Bucket,
   			"key": imageInfo.S3Key,
   		})
   }

   // Create S3 client
   s3Client, err := createS3Client()
   if err != nil {
   	return errors.NewInternalError("s3-client-creation", err)
   }

   // Download image data
   imageData, err := downloadS3Object(s3Client, imageInfo.S3Bucket, imageInfo.S3Key)
   if err != nil {
   	return errors.NewInternalError("s3-image-download", err)
   }

   // Validate image size before encoding
   if len(imageData) > 10*1024*1024 {
   	return errors.NewValidationError("Image size exceeds Bedrock limit", 
   		map[string]interface{}{"sizeMB": len(imageData)/(1024*1024)})
   }

   // Detect and validate image format
   format := detectImageFormatFromHeader(imageData)
   if !isValidBedrockImageFormat(format) {
   	return errors.NewValidationError("Unsupported image format for Bedrock", 
   		map[string]interface{}{
   			"format": format,
   			"supported": []string{"jpeg", "png"},
   		})
   }

   // Encode to Base64
   imageInfo.Base64Data = base64.StdEncoding.EncodeToString(imageData)
   imageInfo.Format = format
   imageInfo.Base64Generated = true

   return nil
}

// BuildTemplateData creates the data object for template rendering
func BuildTemplateData(input *schema.WorkflowState) (TemplateData, error) {
   vCtx := input.VerificationContext
   data := TemplateData{
   	VerificationType: vCtx.VerificationType,
   	VerificationID:   vCtx.VerificationId,
   	VerificationAt:   vCtx.VerificationAt,
   	VendingMachineID: vCtx.VendingMachineId,
   	TurnNumber:       input.CurrentPrompt.TurnNumber,
   }
   
   // Process images for Bedrock if available
   if input.Images != nil {
   	// Process reference image
   	if refImage := input.Images.GetReference(); refImage != nil {
   		if err := ProcessImageForBedrock(refImage); err != nil {
   			return data, fmt.Errorf("failed to process reference image: %w", err)
   		}
   	}
   	
   	// Process checking image if available (though not needed for Turn 1)
   	if checkImage := input.Images.GetChecking(); checkImage != nil {
   		if err := ProcessImageForBedrock(checkImage); err != nil {
   			return data, fmt.Errorf("failed to process checking image: %w", err)
   		}
   	}
   }
   
   // Set machine structure based on verification type
   if vCtx.VerificationType == schema.VerificationTypeLayoutVsChecking && input.LayoutMetadata != nil {
   	// Convert from map to MachineStructure
   	var ms *MachineStructure
   	if machineStructure, ok := input.LayoutMetadata["machineStructure"].(map[string]interface{}); ok {
   		rowCount, _ := machineStructure["rowCount"].(int)
   		columnsPerRow, _ := machineStructure["columnsPerRow"].(int)
   		
   		var rowOrder []string
   		if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok {
   			for _, row := range rowOrderInterface {
   				if rowStr, ok := row.(string); ok {
   					rowOrder = append(rowOrder, rowStr)
   				}
   			}
   		}
   		
   		var columnOrder []string
   		if columnOrderInterface, ok := machineStructure["columnOrder"].([]interface{}); ok {
   			for _, col := range columnOrderInterface {
   				if colStr, ok := col.(string); ok {
   					columnOrder = append(columnOrder, colStr)
   				}
   			}
   		}
   		
   		ms = &MachineStructure{
   			RowCount:      rowCount,
   			ColumnsPerRow: columnsPerRow,
   			RowOrder:      rowOrder,
   			ColumnOrder:   columnOrder,
   		}
   	}
   	
   	if ms != nil {
   		data.MachineStructure = ms
   		data.RowCount = ms.RowCount
   		data.ColumnCount = ms.ColumnsPerRow
   		data.RowLabels = FormatArrayToString(ms.RowOrder)
   		data.ColumnLabels = FormatArrayToString(ms.ColumnOrder)
   		data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
   	}
   	
   	// Extract product mappings if available
   	if productPositionMap, ok := input.LayoutMetadata["productPositionMap"].(map[string]interface{}); ok {
   		data.ProductMappings = FormatProductMappingsFromMap(productPositionMap)
   	}
   	
   	if location, ok := input.LayoutMetadata["location"].(string); ok {
   		data.Location = location
   	}
   } else if vCtx.VerificationType == schema.VerificationTypePreviousVsCurrent && input.HistoricalContext != nil {
   	// Extract historical context data
   	if previousVerificationId, ok := input.HistoricalContext["previousVerificationId"].(string); ok {
   		data.PreviousVerificationID = previousVerificationId
   	}
   	
   	if previousVerificationAt, ok := input.HistoricalContext["previousVerificationAt"].(string); ok {
   		data.PreviousVerificationAt = previousVerificationAt
   	}
   	
   	if previousVerificationStatus, ok := input.HistoricalContext["previousVerificationStatus"].(string); ok {
   		data.PreviousVerificationStatus = previousVerificationStatus
   	}
   	
   	if hoursSinceLastVerification, ok := input.HistoricalContext["hoursSinceLastVerification"].(float64); ok {
   		data.HoursSinceLastVerification = hoursSinceLastVerification
   	}
   	
   	// Set machine structure from historical context if available
   	if machineStructure, ok := input.HistoricalContext["machineStructure"].(map[string]interface{}); ok {
   		rowCount, _ := machineStructure["rowCount"].(int)
   		columnsPerRow, _ := machineStructure["columnsPerRow"].(int)
   		
   		var rowOrder []string
   		if rowOrderInterface, ok := machineStructure["rowOrder"].([]interface{}); ok {
   			for _, row := range rowOrderInterface {
   				if rowStr, ok := row.(string); ok {
   					rowOrder = append(rowOrder, rowStr)
   				}
   			}
   		}
   		
   		var columnOrder []string
   		if columnOrderInterface, ok := machineStructure["columnOrder"].([]interface{}); ok {
   			for _, col := range columnOrderInterface {
   				if colStr, ok := col.(string); ok {
   					columnOrder = append(columnOrder, colStr)
   				}
   			}
   		}
   		
   		ms := &MachineStructure{
   			RowCount:      rowCount,
   			ColumnsPerRow: columnsPerRow,
   			RowOrder:      rowOrder,
   			ColumnOrder:   columnOrder,
   		}
   		
   		data.MachineStructure = ms
   		data.RowCount = ms.RowCount
   		data.ColumnCount = ms.ColumnsPerRow
   		data.RowLabels = FormatArrayToString(ms.RowOrder)
   		data.ColumnLabels = FormatArrayToString(ms.ColumnOrder)
   		data.TotalPositions = ms.RowCount * ms.ColumnsPerRow
   	}
   	
   	// Include previous verification summary if available
   	if verificationSummary, ok := input.HistoricalContext["verificationSummary"].(map[string]interface{}); ok {
   		summary := &VerificationSummary{}
   		
   		if totalPositionsChecked, ok := verificationSummary["totalPositionsChecked"].(int); ok {
   			summary.TotalPositionsChecked = totalPositionsChecked
   		}
   		
   		if correctPositions, ok := verificationSummary["correctPositions"].(int); ok {
   			summary.CorrectPositions = correctPositions
   		}
   		
   		if discrepantPositions, ok := verificationSummary["discrepantPositions"].(int); ok {
   			summary.DiscrepantPositions = discrepantPositions
   		}
   		
   		if missingProducts, ok := verificationSummary["missingProducts"].(int); ok {
   			summary.MissingProducts = missingProducts
   		}
   		
   		if incorrectProductTypes, ok := verificationSummary["incorrectProductTypes"].(int); ok {
   			summary.IncorrectProductTypes = incorrectProductTypes
   		}
   		
   		if unexpectedProducts, ok := verificationSummary["unexpectedProducts"].(int); ok {
   			summary.UnexpectedProducts = unexpectedProducts
   		}
   		
   		if emptyPositionsCount, ok := verificationSummary["emptyPositionsCount"].(int); ok {
   			summary.EmptyPositionsCount = emptyPositionsCount
   		}
   		
   		if overallAccuracy, ok := verificationSummary["overallAccuracy"].(float64); ok {
   			summary.OverallAccuracy = overallAccuracy
   		}
   		
   		if overallConfidence, ok := verificationSummary["overallConfidence"].(float64); ok {
   			summary.OverallConfidence = overallConfidence
   		}
   		
   		if verificationStatus, ok := verificationSummary["verificationStatus"].(string); ok {
   			summary.VerificationStatus = verificationStatus
   		}
   		
   		if verificationOutcome, ok := verificationSummary["verificationOutcome"].(string); ok {
   			summary.VerificationOutcome = verificationOutcome
   		}
   		
   		data.VerificationSummary = summary
   	}
   }
   
   return data, nil
}

// Helper functions

// createS3Client creates a new S3 client with default configuration
func createS3Client() (*s3.Client, error) {
   cfg, err := config.LoadDefaultConfig(context.TODO())
   if err != nil {
   	return nil, fmt.Errorf("failed to load AWS config: %w", err)
   }
   return s3.NewFromConfig(cfg), nil
}

// downloadS3Object downloads an object from S3 and returns the data
func downloadS3Object(client *s3.Client, bucket, key string) ([]byte, error) {
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()

   result, err := client.GetObject(ctx, &s3.GetObjectInput{
   	Bucket: &bucket,
   	Key:    &key,
   })
   if err != nil {
   	return nil, fmt.Errorf("failed to download object s3://%s/%s: %w", bucket, key, err)
   }
   defer result.Body.Close()

   data, err := io.ReadAll(result.Body)
   if err != nil {
   	return nil, fmt.Errorf("failed to read object body: %w", err)
   }

   return data, nil
}

// downloadS3ObjectAsString downloads an object from S3 and returns it as a string
func downloadS3ObjectAsString(client *s3.Client, bucket, key string) (string, error) {
   data, err := downloadS3Object(client, bucket, key)
   if err != nil {
   	return "", err
   }
   return string(data), nil
}

// isValidBase64 checks if a string is valid Base64
func isValidBase64(s string) bool {
   _, err := base64.StdEncoding.DecodeString(s)
   return err == nil
}

// detectImageFormatFromHeader detects image format from file header
func detectImageFormatFromHeader(data []byte) string {
   if len(data) < 4 {
   	return ""
   }

   // Check JPEG
   if data[0] == 0xFF && data[1] == 0xD8 {
   	return "jpeg"
   }

   // Check PNG
   if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
   	return "png"
   }

   return ""
}

// isValidBedrockImageFormat checks if the format is supported by Bedrock
func isValidBedrockImageFormat(format string) bool {
   format = strings.ToLower(format)
   return format == "jpeg" || format == "jpg" || format == "png"
}

// FormatArrayToString formats a string array to a comma-separated string
func FormatArrayToString(arr []string) string {
   if len(arr) == 0 {
   	return ""
   }
   return strings.Join(arr, ", ")
}

// FormatProductMappingsFromMap converts a product position map to a formatted array
func FormatProductMappingsFromMap(positionMap map[string]interface{}) []ProductMapping {
   if positionMap == nil {
   	return []ProductMapping{}
   }
   
   mappings := make([]ProductMapping, 0, len(positionMap))
   for position, infoInterface := range positionMap {
   	if info, ok := infoInterface.(map[string]interface{}); ok {
   		productID, _ := info["productId"].(int)
   		productName, _ := info["productName"].(string)
   		
   		mappings = append(mappings, ProductMapping{
   			Position:    position,
   			ProductID:   productID,
   			ProductName: productName,
   		})
   	}
   }
   
   return mappings
}

// FormatRowHighlights creates a map of row labels to descriptions based on historical context
func FormatRowHighlights(checkingStatus map[string]string) map[string]string {
   if checkingStatus == nil {
   	return nil
   }
   
   // Create a cleaned-up version of the checking status
   highlights := make(map[string]string)
   for row, status := range checkingStatus {
   	// Extract the key information and remove verbose parts
   	cleaned := strings.ReplaceAll(status, "Current: ", "")
   	cleaned = strings.ReplaceAll(cleaned, "Status: ", "")
   	
   	// Add to highlights
   	highlights[row] = cleaned
   }
   
   return highlights
}

// CreateTurn1MessageContent generates the text content for the Turn 1 message
func CreateTurn1MessageContent(verificationType string, promptText string, rowCount int, topRow, bottomRow string) string {
   if promptText != "" {
   	return promptText
   }
   
   // Fallback to basic message if template didn't provide content
   base := "Please analyze the FIRST image (Reference Image)\n\n"
   
   if verificationType == schema.VerificationTypeLayoutVsChecking {
   	base += "This image shows the approved product arrangement according to the planogram.\n\n"
   } else if verificationType == schema.VerificationTypePreviousVsCurrent {
   	base += "This image shows the previous state of the vending machine.\n\n"
   }
   
   base += fmt.Sprintf("Focus exclusively on analyzing this Reference Image in detail. Your goal is to identify the exact contents of all %d rows (%s-%s).\n\n", 
   	rowCount, topRow, bottomRow)
   
   base += "Important reminders:\n"
   base += fmt.Sprintf("1. Row identification is CRITICAL - Row %s is ALWAYS the topmost physical shelf, Row %s is ALWAYS the bottommost physical shelf.\n", 
   	topRow, bottomRow)
   base += "2. Be thorough and descriptive in your analysis of each row status (Full/Partial/Empty).\n"
   base += "3. DO NOT compare with any other image at this stage - just analyze this Reference Image.\n"
   
   return base
}