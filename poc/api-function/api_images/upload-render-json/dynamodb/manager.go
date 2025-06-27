package dynamodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"api_images_upload_render/renderer"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Manager handles DynamoDB operations for layout metadata
type Manager struct {
	client    *dynamodb.Client
	tableName string
}

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

// ProductInfo represents product information for a specific position
type ProductInfo struct {
	ProductID           int    `json:"productId" dynamodbav:"productId"`
	ProductTemplateID   int    `json:"productTemplateId" dynamodbav:"productTemplateId"`
	ProductTemplateName string `json:"productTemplateName" dynamodbav:"productTemplateName"`
	ProductTemplateImage string `json:"productTemplateImage" dynamodbav:"productTemplateImage"`
	MaxQuantity         int    `json:"maxQuantity" dynamodbav:"maxQuantity"`
	Status              int    `json:"status" dynamodbav:"status"`
}

// NewManager creates a new DynamoDB manager
func NewManager(ctx context.Context, region, tableName string) (*Manager, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	
	return &Manager{
		client:    client,
		tableName: tableName,
	}, nil
}

// StoreLayoutMetadata stores layout metadata in DynamoDB
func (m *Manager) StoreLayoutMetadata(ctx context.Context, layout *renderer.Layout, layoutPrefix, s3Bucket, processedKey, sourceKey string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	
	// Build machine structure from layout data
	machineStructure := m.buildMachineStructure(layout)
	
	// Build row-product mapping
	rowProductMapping := m.buildRowProductMapping(layout)
	
	// Build product position map
	productPositionMap := m.buildProductPositionMap(layout)
	
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

	// Convert to DynamoDB attribute values
	item, err := attributevalue.MarshalMap(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal layout metadata: %v", err)
	}

	// Put item with condition to prevent overwriting existing items
	_, err = m.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(m.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(layoutId)"),
	})

	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			return fmt.Errorf("layout with ID %d already exists", layout.LayoutID)
		}
		return fmt.Errorf("failed to store layout metadata: %v", err)
	}

	return nil
}

// buildMachineStructure extracts machine structure from layout
func (m *Manager) buildMachineStructure(layout *renderer.Layout) MachineStructure {
	if len(layout.SubLayoutList) == 0 || len(layout.SubLayoutList[0].TrayList) == 0 {
		return MachineStructure{}
	}

	trays := layout.SubLayoutList[0].TrayList
	rowCount := len(trays)
	
	// Determine columns per row (assuming all rows have the same number of columns)
	columnsPerRow := 0
	if len(trays) > 0 && len(trays[0].SlotList) > 0 {
		// Find the maximum slot number to determine column count
		for _, slot := range trays[0].SlotList {
			if slot.SlotNo > columnsPerRow {
				columnsPerRow = slot.SlotNo
			}
		}
	}

	// Build row order
	rowOrder := make([]string, rowCount)
	for i, tray := range trays {
		rowOrder[i] = tray.TrayCode
	}

	// Build column order (1, 2, 3, ...)
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

// buildRowProductMapping creates a mapping of row -> column -> product name
func (m *Manager) buildRowProductMapping(layout *renderer.Layout) map[string]map[string]string {
	mapping := make(map[string]map[string]string)
	
	if len(layout.SubLayoutList) == 0 {
		return mapping
	}

	for _, tray := range layout.SubLayoutList[0].TrayList {
		rowMapping := make(map[string]string)
		
		for _, slot := range tray.SlotList {
			columnKey := fmt.Sprintf("%d", slot.SlotNo)
			rowMapping[columnKey] = slot.ProductTemplateName
		}
		
		mapping[tray.TrayCode] = rowMapping
	}

	return mapping
}

// buildProductPositionMap creates a mapping of position -> product info
func (m *Manager) buildProductPositionMap(layout *renderer.Layout) map[string]ProductInfo {
	positionMap := make(map[string]ProductInfo)
	
	if len(layout.SubLayoutList) == 0 {
		return positionMap
	}

	for _, tray := range layout.SubLayoutList[0].TrayList {
		for _, slot := range tray.SlotList {
			position := fmt.Sprintf("%s%d", tray.TrayCode, slot.SlotNo)
			
			productInfo := ProductInfo{
				ProductID:           slot.ProductId,
				ProductTemplateID:   slot.ProductTemplateId,
				ProductTemplateName: slot.ProductTemplateName,
				ProductTemplateImage: slot.ProductTemplateImage,
				MaxQuantity:         slot.MaxQuantity,
				Status:              slot.Status,
			}
			
			positionMap[position] = productInfo
		}
	}

	return positionMap
}
