package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"vending-machine-layout-generator/renderer"
)

// LayoutMetadata represents the structure stored in DynamoDB
type LayoutMetadata struct {
	LayoutID           int64             `json:"layoutId" dynamodbav:"layoutId"`
	LayoutPrefix       string            `json:"layoutPrefix" dynamodbav:"layoutPrefix"`
	VendingMachineID   string            `json:"vendingMachineId" dynamodbav:"vendingMachineId"`
	Location           string            `json:"location" dynamodbav:"location"`
	CreatedAt          string            `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt          string            `json:"updatedAt" dynamodbav:"updatedAt"`
	ReferenceImageURL  string            `json:"referenceImageUrl" dynamodbav:"referenceImageUrl"`
	SourceJSONURL      string            `json:"sourceJsonUrl" dynamodbav:"sourceJsonUrl"`
	MachineStructure   MachineStructure  `json:"machineStructure" dynamodbav:"machineStructure"`
	RowProductMapping  map[string]map[string]string `json:"rowProductMapping" dynamodbav:"rowProductMapping"`
	ProductPositionMap map[string]ProductInfo `json:"productPositionMap" dynamodbav:"productPositionMap"`
}

// MachineStructure represents the physical structure of the vending machine
type MachineStructure struct {
	RowCount       int      `json:"rowCount" dynamodbav:"rowCount"`
	ColumnsPerRow  int      `json:"columnsPerRow" dynamodbav:"columnsPerRow"`
	ColumnOrder    []string `json:"columnOrder" dynamodbav:"columnOrder"`
	RowOrder       []string `json:"rowOrder" dynamodbav:"rowOrder"`
}

// ProductInfo represents information about a product in a specific position
type ProductInfo struct {
	ProductID    int    `json:"productId" dynamodbav:"productId"`
	ProductName  string `json:"productName" dynamodbav:"productName"`
	ProductImage string `json:"productImage" dynamodbav:"productImage"`
}

// Manager handles DynamoDB operations
type Manager struct {
	client    *dynamodb.Client
	tableName string
}

// NewManager creates a new DynamoDB manager
func NewManager(ctx context.Context, region, tableName string) (*Manager, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}
	
	client := dynamodb.NewFromConfig(cfg)
	
	return &Manager{
		client:    client,
		tableName: tableName,
	}, nil
}

// StoreLayoutMetadata saves layout metadata to DynamoDB
func (m *Manager) StoreLayoutMetadata(ctx context.Context, layout *renderer.Layout, layoutPrefix, s3Bucket, processedKey, sourceKey string) error {
	// Create machine structure from the layout
	machineStructure := extractMachineStructure(layout)
	
	// Generate row product mapping
	rowProductMapping := extractRowProductMapping(layout)
	
	// Generate product position map
	productPositionMap := extractProductPositionMap(layout)
	
	// Create the complete metadata object
	now := time.Now().Format(time.RFC3339)
	
	// Extract vendor machine ID and location (in a real implementation, this could come from the layout data)
	vendingMachineID := fmt.Sprintf("VM-%d", layout.LayoutID)
	location := "Default Location" // This would come from the layout in a real implementation
	
	metadata := LayoutMetadata{
		LayoutID:          layout.LayoutID,
		LayoutPrefix:      layoutPrefix,
		VendingMachineID:  vendingMachineID,
		Location:          location,
		CreatedAt:         now,
		UpdatedAt:         now,
		ReferenceImageURL: fmt.Sprintf("s3://%s/%s", s3Bucket, processedKey),
		SourceJSONURL:     fmt.Sprintf("s3://%s/%s", s3Bucket, sourceKey),
		MachineStructure:  machineStructure,
		RowProductMapping: rowProductMapping,
		ProductPositionMap: productPositionMap,
	}
	
	// Convert metadata to DynamoDB attribute values
	item, err := attributevalue.MarshalMap(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}
	
	// Create condition expression to prevent overwriting existing items with the same key
	conditionExpression := "attribute_not_exists(layoutId) AND attribute_not_exists(layoutPrefix)"
	
	// Save to DynamoDB with conditional write
	_, err = m.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(m.tableName),
		Item:                item,
		ConditionExpression: aws.String(conditionExpression),
	})
	
	if err != nil {
		var _, ok = err.(*types.ConditionalCheckFailedException)
		if ok {
			return fmt.Errorf("layout with ID %d and prefix %s already exists", layout.LayoutID, layoutPrefix)
		}
		return fmt.Errorf("failed to store metadata in DynamoDB: %v", err)
	}
	
	return nil
}

// extractMachineStructure creates a MachineStructure from layout data
func extractMachineStructure(layout *renderer.Layout) MachineStructure {
	// In a real implementation, this would extract row and column structure from the layout
	// For this example, we'll create a simple structure
	
	// Get the number of rows (trays)
	rowCount := 0
	if len(layout.SubLayoutList) > 0 {
		rowCount = len(layout.SubLayoutList[0].TrayList)
	}
	
	// Get the max columns per row
	columnsPerRow := 0
	columnMap := make(map[int]bool)
	rowOrder := make([]string, 0, rowCount)
	
	if len(layout.SubLayoutList) > 0 {
		for _, tray := range layout.SubLayoutList[0].TrayList {
			// Add row identifier to rowOrder
			rowOrder = append(rowOrder, tray.TrayCode)
			
			// Find max columns
			for _, slot := range tray.SlotList {
				columnMap[slot.SlotNo] = true
				if slot.SlotNo > columnsPerRow {
					columnsPerRow = slot.SlotNo
				}
			}
		}
	}
	
	// Generate column order
	columnOrder := make([]string, columnsPerRow)
	for i := 0; i < columnsPerRow; i++ {
		columnOrder[i] = fmt.Sprintf("%d", i+1)
	}
	
	return MachineStructure{
		RowCount:      rowCount,
		ColumnsPerRow: columnsPerRow,
		ColumnOrder:   columnOrder,
		RowOrder:      rowOrder,
	}
}

// extractRowProductMapping creates a mapping of row/column to product name
func extractRowProductMapping(layout *renderer.Layout) map[string]map[string]string {
	mapping := make(map[string]map[string]string)
	
	if len(layout.SubLayoutList) > 0 {
		for _, tray := range layout.SubLayoutList[0].TrayList {
			rowMapping := make(map[string]string)
			
			for _, slot := range tray.SlotList {
				columnKey := fmt.Sprintf("%d", slot.SlotNo)
				rowMapping[columnKey] = slot.ProductTemplateName
			}
			
			mapping[tray.TrayCode] = rowMapping
		}
	}
	
	return mapping
}

// extractProductPositionMap creates a mapping of position to product info
func extractProductPositionMap(layout *renderer.Layout) map[string]ProductInfo {
	mapping := make(map[string]ProductInfo)
	
	if len(layout.SubLayoutList) > 0 {
		for _, tray := range layout.SubLayoutList[0].TrayList {
			for _, slot := range tray.SlotList {
				position := fmt.Sprintf("%s%02d", tray.TrayCode, slot.SlotNo)
				
				mapping[position] = ProductInfo{
					ProductID:    slot.ProductId,
					ProductName:  slot.ProductTemplateName,
					ProductImage: slot.ProductTemplateImage,
				}
			}
		}
	}
	
	return mapping
}