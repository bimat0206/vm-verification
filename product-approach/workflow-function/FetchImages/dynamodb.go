package main

import (
    "context"
    "fmt"
    "strconv"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)




// FetchLayoutMetadata retrieves layout metadata from DynamoDB.
func FetchLayoutMetadata(ctx context.Context, layoutId int64, layoutPrefix string) (*LayoutMetadata, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    client := dynamodb.NewFromConfig(cfg)

    // Load configuration for table name
    config := LoadConfig()
    tableName := config.LayoutTableName

    // Log the table being accessed
    Info("Accessing DynamoDB table for layout metadata", map[string]interface{}{
        "tableName": tableName,
        "layoutId": layoutId,
        "layoutPrefix": layoutPrefix,
    })

    out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "layoutId":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", layoutId)},
            "layoutPrefix": &types.AttributeValueMemberS{Value: layoutPrefix},
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to fetch layout metadata: %w", err)
    }
    if out.Item == nil {
        return nil, fmt.Errorf("layout metadata not found")
    }

    // Create layout metadata from attributes
    layoutMeta := &LayoutMetadata{
        LayoutId:     layoutId,
        LayoutPrefix: layoutPrefix,
    }
    
    // Extract string attributes
    if val, ok := out.Item["vendingMachineId"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.VendingMachineId = sv.Value
        }
    }
    
    if val, ok := out.Item["location"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.Location = sv.Value
        }
    }
    
    if val, ok := out.Item["createdAt"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.CreatedAt = sv.Value
        }
    }
    
    if val, ok := out.Item["updatedAt"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.UpdatedAt = sv.Value
        }
    }
    
    if val, ok := out.Item["referenceImageUrl"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.ReferenceImageUrl = sv.Value
        }
    }
    
    if val, ok := out.Item["sourceJsonUrl"]; ok {
        if sv, ok := val.(*types.AttributeValueMemberS); ok {
            layoutMeta.SourceJsonUrl = sv.Value
        }
    }
    
    // Extract machine structure
    if machineVal, ok := out.Item["machineStructure"]; ok {
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
    if productMapVal, ok := out.Item["productPositionMap"]; ok {
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
    if rowMapVal, ok := out.Item["rowProductMapping"]; ok {
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

// FetchHistoricalContext retrieves previous verification from DynamoDB.
func FetchHistoricalContext(ctx context.Context, verificationId string) (*HistoricalContext, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    client := dynamodb.NewFromConfig(cfg)

    // Load configuration for table name
    config := LoadConfig()
    tableName := config.VerificationTableName

    // Log the table being accessed
    Info("Accessing DynamoDB table for historical verification", map[string]interface{}{
        "tableName": tableName,
        "verificationId": verificationId,
    })

    out, err := client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "verificationId": &types.AttributeValueMemberS{Value: verificationId},
        },
    })
    if err != nil {
        return nil, fmt.Errorf("failed to fetch historical context: %w", err)
    }
    if out.Item == nil {
        return nil, fmt.Errorf("historical verification not found")
    }

    // Extract verification timestamp
    var verificationAt string
    if atVal, ok := out.Item["verificationAt"]; ok {
        if sv, ok := atVal.(*types.AttributeValueMemberS); ok {
            verificationAt = sv.Value
        }
    }
    
    // Extract status
    var status string
    if statusVal, ok := out.Item["status"]; ok {
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
    if machineStructureVal, ok := out.Item["machineStructure"]; ok {
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
    if summaryVal, ok := out.Item["verificationSummary"]; ok {
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

// Optionally, update verification status in DynamoDB.
func UpdateVerificationStatus(ctx context.Context, verificationId, status string) error {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return fmt.Errorf("failed to load AWS config: %w", err)
    }
    client := dynamodb.NewFromConfig(cfg)

    // Load configuration for table name
    config := LoadConfig()
    tableName := config.VerificationTableName

    // Log the table being accessed
    Info("Updating verification status in DynamoDB", map[string]interface{}{
        "tableName": tableName,
        "verificationId": verificationId,
        "newStatus": status,
    })

    _, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String(tableName),
        Key: map[string]types.AttributeValue{
            "verificationId": &types.AttributeValueMemberS{Value: verificationId},
        },
        UpdateExpression: aws.String("SET verificationStatus = :status, updatedAt = :now"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":status": &types.AttributeValueMemberS{Value: status},
            ":now":    &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
        },
    })
    if err != nil {
        return fmt.Errorf("failed to update verification status: %w", err)
    }
    return nil
}
