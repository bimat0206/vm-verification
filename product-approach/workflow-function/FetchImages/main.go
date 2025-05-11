package main

import (
    "context"
    "encoding/json"
    "fmt"
    "strconv"
    //"strings"
    "time"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go-v2/aws"
    //"github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)



// Handler is the Lambda handler function
func Handler(ctx context.Context, event interface{}) (FetchImagesResponse, error) {
    // Parse input based on event type
    var req FetchImagesRequest
    var err error

    // Log the incoming event for debugging
    eventBytes, _ := json.Marshal(event)
    Info("Received event", string(eventBytes))

    // Handle different invocation types
    switch e := event.(type) {
    case events.LambdaFunctionURLRequest:
        // Function URL invocation
        if err = json.Unmarshal([]byte(e.Body), &req); err != nil {
            Error("Failed to parse Function URL input", err)
            return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", err)
        }
    case map[string]interface{}:
        // Direct invocation from Step Function
        data, _ := json.Marshal(e)
        if err = json.Unmarshal(data, &req); err != nil {
            Error("Failed to parse Step Function input", map[string]interface{}{
                "error": err.Error(),
                "input": fmt.Sprintf("%+v", e),
            })
            return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", err)
        }
    case FetchImagesRequest:
        // Direct struct invocation
        req = e
    default:
        // Try raw JSON unmarshal as fallback
        data, _ := json.Marshal(event)
        if err = json.Unmarshal(data, &req); err != nil {
            Error("Failed to parse unknown input type", map[string]interface{}{
                "error": err.Error(),
                "input": fmt.Sprintf("%+v", event),
            })
            return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", err)
        }
    }

    // Validate input
    if err := req.Validate(); err != nil {
        return FetchImagesResponse{}, NewBadRequestError("Input validation failed", err)
    }

    // Initialize verificationContext if it doesn't exist or normalize from direct fields
    var verificationContext VerificationContext
    if req.VerificationContext != nil {
        verificationContext = *req.VerificationContext
    } else {
        // Create from direct fields
        verificationContext = VerificationContext{
            VerificationId:    req.VerificationId,
            VerificationType:  req.VerificationType,
            ReferenceImageUrl: req.ReferenceImageUrl,
            CheckingImageUrl:  req.CheckingImageUrl,
            LayoutId:          req.LayoutId,
            LayoutPrefix:      req.LayoutPrefix,
            VendingMachineId:  req.VendingMachineId,
        }
    }

    // Update status in verification context
    verificationContext.Status = "IMAGES_FETCHED"

    // Parse S3 URIs
    referenceS3, err := ParseS3URI(verificationContext.ReferenceImageUrl)
    if err != nil {
        return FetchImagesResponse{}, NewBadRequestError("Invalid referenceImageUrl", err)
    }
    checkingS3, err := ParseS3URI(verificationContext.CheckingImageUrl)
    if err != nil {
        return FetchImagesResponse{}, NewBadRequestError("Invalid checkingImageUrl", err)
    }

    // Fetch all data in parallel (metadata, DynamoDB context)
    results := ParallelFetch(
        ctx,
        referenceS3,
        checkingS3,
        verificationContext.LayoutId,
        verificationContext.LayoutPrefix,
        verificationContext.PreviousVerificationId,
    )

    // Check for errors from parallel processing
    if len(results.Errors) > 0 {
        // Log all errors but return the first one
        for _, fetchErr := range results.Errors {
            Error("Error during parallel fetch", fetchErr)
        }
        return FetchImagesResponse{}, NewNotFoundError("Failed to fetch required resources", results.Errors[0])
    }

    // Construct response with complete verification context
    resp := FetchImagesResponse{
        VerificationContext: verificationContext,
        Images: ImagesData{
            ReferenceImageMeta: results.ReferenceMeta,
            CheckingImageMeta:  results.CheckingMeta,
        },
        LayoutMetadata:    results.LayoutMeta,
        HistoricalContext: results.HistoricalContext,
    }

    // Optionally update status in DynamoDB (optional)
    // UpdateVerificationStatus(ctx, verificationContext.VerificationId, "IMAGES_FETCHED")

    Info("Successfully processed images", map[string]interface{}{
        "verificationId":     verificationContext.VerificationId,
        "verificationType":   verificationContext.VerificationType,
        "referenceImageSize": results.ReferenceMeta.Size,
        "checkingImageSize":  results.CheckingMeta.Size,
    })

    return resp, nil
}

// fetchLayoutMetadataFromDynamoDB retrieves the layout metadata from DynamoDB
// This version manually parses the DynamoDB attributes without using the attributevalue package
func fetchLayoutMetadataFromDynamoDB(ctx context.Context, client *dynamodb.Client, tableName string, layoutId int64, layoutPrefix string) (*LayoutMetadata, error) {
    // Create key for GetItem
    key := map[string]types.AttributeValue{
        "layoutId":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", layoutId)},
        "layoutPrefix": &types.AttributeValueMemberS{Value: layoutPrefix},
    }
    
    // Get item from DynamoDB
    result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(tableName),
        Key:       key,
    })
    if err != nil {
        return nil, fmt.Errorf("GetItem error: %w", err)
    }
    
    // Check if item exists
    if result.Item == nil || len(result.Item) == 0 {
        return nil, fmt.Errorf("layout not found: %d/%s", layoutId, layoutPrefix)
    }
    
    // Create layout metadata manually from attributes
    layoutMeta := &LayoutMetadata{
        LayoutId:     layoutId,
        LayoutPrefix: layoutPrefix,
    }
    
    // Extract string attributes
    if val, ok := result.Item["vendingMachineId"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.VendingMachineId = sv.Value
        }
    }
    
    if val, ok := result.Item["location"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.Location = sv.Value
        }
    }
    
    if val, ok := result.Item["createdAt"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.CreatedAt = sv.Value
        }
    }
    
    if val, ok := result.Item["updatedAt"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.UpdatedAt = sv.Value
        }
    }
    
    if val, ok := result.Item["referenceImageUrl"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.ReferenceImageUrl = sv.Value
        }
    }
    
    if val, ok := result.Item["sourceJsonUrl"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.SourceJsonUrl = sv.Value
        }
    }
    
    // Extract machine structure
    if machineVal, ok := result.Item["machineStructure"]; ok {
        if mv, ok := machineVal.(*types.AttributeValueMemberM); ok {
            machineStructure := &MachineStructure{}
            
            // Extract row count
            if rowCountVal, ok := mv.Value["rowCount"]; ok {
                if rv, ok := rowCountVal.(*types.AttributeValueMemberN); ok {
                    rowCount, _ := strconv.Atoi(rv.Value)
                    machineStructure.RowCount = rowCount
                }
            }
            
            // Extract columns per row
            if colsPerRowVal, ok := mv.Value["columnsPerRow"]; ok {
                if cv, ok := colsPerRowVal.(*types.AttributeValueMemberN); ok {
                    colsPerRow, _ := strconv.Atoi(cv.Value)
                    machineStructure.ColumnsPerRow = colsPerRow
                }
            }
            
            // Extract row order
            if rowOrderVal, ok := mv.Value["rowOrder"]; ok {
                if rv, ok := rowOrderVal.(*types.AttributeValueMemberL); ok {
                    rowOrder := make([]string, 0, len(rv.Value))
                    for _, rowVal := range rv.Value {
                        if sv, ok := rowVal.(*types.AttributeValueMemberS); ok {
                            rowOrder = append(rowOrder, sv.Value)
                        }
                    }
                    machineStructure.RowOrder = rowOrder
                }
            }
            
            // Extract column order
            if colOrderVal, ok := mv.Value["columnOrder"]; ok {
                if cv, ok := colOrderVal.(*types.AttributeValueMemberL); ok {
                    colOrder := make([]string, 0, len(cv.Value))
                    for _, colVal := range cv.Value {
                        if sv, ok := colVal.(*types.AttributeValueMemberS); ok {
                            colOrder = append(colOrder, sv.Value)
                        }
                    }
                    machineStructure.ColumnOrder = colOrder
                }
            }
            
            layoutMeta.MachineStructure = machineStructure
        }
    }
    
    // Extract product position map (keep as generic map)
    if productMapVal, ok := result.Item["productPositionMap"]; ok {
        if pv, ok := productMapVal.(*types.AttributeValueMemberM); ok {
            // Convert to a simple map[string]interface{} for JSON output
            productMap := make(map[string]interface{})
            
            // Loop through each position
            for position, posValue := range pv.Value {
                if posMap, ok := posValue.(*types.AttributeValueMemberM); ok {
                    // Parse product data
                    product := make(map[string]interface{})
                    for key, val := range posMap.Value {
                        switch v := val.(type) {
                        case *types.AttributeValueMemberS:
                            product[key] = v.Value
                        case *types.AttributeValueMemberN:
                            if num, err := strconv.ParseInt(v.Value, 10, 64); err == nil {
                                product[key] = num
                            } else {
                                product[key] = v.Value
                            }
                        }
                    }
                    productMap[position] = product
                }
            }
            
            layoutMeta.ProductPositionMap = productMap
        }
    }
    
    // Extract row product mapping (keep as generic map)
    if rowMapVal, ok := result.Item["rowProductMapping"]; ok {
        if rm, ok := rowMapVal.(*types.AttributeValueMemberM); ok {
            rowMap := make(map[string]interface{})
            
            // Loop through each row
            for row, rowValue := range rm.Value {
                if rowAttr, ok := rowValue.(*types.AttributeValueMemberM); ok {
                    rowProducts := make(map[string]string)
                    
                    // Loop through each column in the row
                    for col, colValue := range rowAttr.Value {
                        if colAttr, ok := colValue.(*types.AttributeValueMemberS); ok {
                            rowProducts[col] = colAttr.Value
                        }
                    }
                    
                    rowMap[row] = rowProducts
                }
            }
            
            layoutMeta.RowProductMapping = rowMap
        }
    }
    
    return layoutMeta, nil
}

// fetchHistoricalContextFromDynamoDB retrieves the historical context from DynamoDB
func fetchHistoricalContextFromDynamoDB(ctx context.Context, client *dynamodb.Client, tableName string, verificationId string) (*HistoricalContext, error) {
    // Create key for GetItem
    key := map[string]types.AttributeValue{
        "verificationId": &types.AttributeValueMemberS{Value: verificationId},
    }
    
    // Get item from DynamoDB
    result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(tableName),
        Key:       key,
    })
    if err != nil {
        return nil, fmt.Errorf("GetItem error: %w", err)
    }
    
    // Check if item exists
    if result.Item == nil || len(result.Item) == 0 {
        return nil, fmt.Errorf("verification not found: %s", verificationId)
    }
    
    // Extract verification timestamp
    var verificationAt string
    if atVal, ok := result.Item["verificationAt"]; ok {
        if sv, ok := atVal.(*types.AttributeValueMemberS); ok {
            verificationAt = sv.Value
        }
    }
    
    // Extract status
    var status string
    if statusVal, ok := result.Item["status"]; ok {
        if sv, ok := statusVal.(*types.AttributeValueMemberS); ok {
            status = sv.Value
        }
    }
    
    // Calculate hours since verification
    var hoursSince float64 = 0
    if verificationAt != "" {
        verTime, err := time.Parse(time.RFC3339, verificationAt)
        if err == nil {
            hoursSince = time.Since(verTime).Hours()
        }
    }
    
    // Create historical context
    historicalCtx := &HistoricalContext{
        PreviousVerificationId:     verificationId,
        PreviousVerificationAt:     verificationAt,
        PreviousVerificationStatus: status,
        HoursSinceLastVerification: hoursSince,
    }
    
    // Extract machine structure if available
    if machineStructureVal, ok := result.Item["machineStructure"]; ok {
        if mv, ok := machineStructureVal.(*types.AttributeValueMemberM); ok {
            machineStructure := &MachineStructure{}
            
            // Extract row count
            if rowCountVal, ok := mv.Value["rowCount"]; ok {
                if rv, ok := rowCountVal.(*types.AttributeValueMemberN); ok {
                    rowCount, _ := strconv.Atoi(rv.Value)
                    machineStructure.RowCount = rowCount
                }
            }
            
            // Extract columns per row
            if colsPerRowVal, ok := mv.Value["columnsPerRow"]; ok {
                if cv, ok := colsPerRowVal.(*types.AttributeValueMemberN); ok {
                    colsPerRow, _ := strconv.Atoi(cv.Value)
                    machineStructure.ColumnsPerRow = colsPerRow
                }
            }
            
            // Extract row order
            if rowOrderVal, ok := mv.Value["rowOrder"]; ok {
                if rv, ok := rowOrderVal.(*types.AttributeValueMemberL); ok {
                    rowOrder := make([]string, 0, len(rv.Value))
                    for _, rowVal := range rv.Value {
                        if sv, ok := rowVal.(*types.AttributeValueMemberS); ok {
                            rowOrder = append(rowOrder, sv.Value)
                        }
                    }
                    machineStructure.RowOrder = rowOrder
                }
            }
            
            // Extract column order
            if colOrderVal, ok := mv.Value["columnOrder"]; ok {
                if cv, ok := colOrderVal.(*types.AttributeValueMemberL); ok {
                    colOrder := make([]string, 0, len(cv.Value))
                    for _, colVal := range cv.Value {
                        if sv, ok := colVal.(*types.AttributeValueMemberS); ok {
                            colOrder = append(colOrder, sv.Value)
                        }
                    }
                    machineStructure.ColumnOrder = colOrder
                }
            }
            
            historicalCtx.MachineStructure = machineStructure
        }
    }
    
    // Extract verification summary if available 
    if summaryVal, ok := result.Item["verificationSummary"]; ok {
        if sv, ok := summaryVal.(*types.AttributeValueMemberM); ok {
            summary := make(map[string]interface{})
            
            // Process each summary field
            for key, val := range sv.Value {
                switch v := val.(type) {
                case *types.AttributeValueMemberS:
                    summary[key] = v.Value
                case *types.AttributeValueMemberN:
                    // Try to convert to number if possible
                    if n, err := strconv.ParseFloat(v.Value, 64); err == nil {
                        summary[key] = n
                    } else if i, err := strconv.ParseInt(v.Value, 10, 64); err == nil {
                        summary[key] = i
                    } else {
                        summary[key] = v.Value // Keep as string if parsing fails
                    }
                }
            }
            
            historicalCtx.Summary = summary
        }
    }
    
    return historicalCtx, nil
}

func main() {
    lambda.Start(Handler)
}
