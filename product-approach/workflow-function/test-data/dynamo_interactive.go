package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	// Import our populate_data package
	"test/populate_data"
)

// Configuration stores the mappings and settings
type Configuration struct {
	LayoutMetadataTable      string `json:"layoutMetadataTable"`
	VerificationResultsTable string `json:"verificationResultsTable"`
	ConversationHistoryTable string `json:"conversationHistoryTable"`
	Region                   string `json:"region"`
	NumLayouts               int    `json:"numLayouts"`
	NumVerifications         int    `json:"numVerifications"`
	NumConversations         int    `json:"numConversations"`
}

// TableSchemaInfo stores simplified schema information 
type TableSchemaInfo struct {
	TableName             string   `json:"tableName"`
	PrimaryKey           string   `json:"primaryKey"`
	SortKey              string   `json:"sortKey,omitempty"`
	SecondaryIndexes     []string `json:"secondaryIndexes,omitempty"`
	ApproximateItemCount int64    `json:"approximateItemCount"`
}

const configFile = "dynamo_config.json"

func main() {
	fmt.Println("=== DynamoDB Interactive Tool ===")
	
	// Load configuration if exists
	config := loadConfiguration()
	
	// Initialize AWS SDK
	ctx := context.Background()
	awsCfg, err := setupAWS(ctx, config.Region)
	if err != nil {
		log.Fatalf("Failed to set up AWS: %v", err)
	}
	
	// Create DynamoDB client
	client := dynamodb.NewFromConfig(awsCfg)
	
	// Main menu loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\nMain Menu:")
		fmt.Println("1. List and select DynamoDB tables")
		fmt.Println("2. View current configuration")
		fmt.Println("3. View table schema")
		fmt.Println("4. Populate test data")
		fmt.Println("5. Change AWS region")
		fmt.Println("6. Configure data generation counts")
		fmt.Println("7. Save configuration")
		fmt.Println("8. Exit")
		fmt.Print("Select an option: ")
		
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		
		switch choice {
		case "1":
			listAndSelectTables(ctx, client, &config, reader)
		case "2":
			viewConfiguration(config)
		case "3":
			viewTableSchema(ctx, client, config, reader)
		case "4":
			populateTestData(ctx, client, config, reader)
		case "5":
			changeRegion(&config, reader)
			// Reload AWS config with new region
			awsCfg, err = setupAWS(ctx, config.Region)
			if err != nil {
				log.Printf("Failed to update AWS region: %v", err)
				continue
			}
			client = dynamodb.NewFromConfig(awsCfg)
		case "6":
			configureDataCounts(&config, reader)
		case "7":
			saveConfiguration(config)
		case "8":
			fmt.Println("Exiting program.")
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func setupAWS(ctx context.Context, region string) (aws.Config, error) {
	if region == "" {
		region = "us-east-1" // Default region
	}
	
	return config.LoadDefaultConfig(ctx, 
		config.WithRegion(region),
	)
}

func loadConfiguration() Configuration {
	config := Configuration{
		Region:          "us-east-1", // Default region
		NumLayouts:      5,           // Default values
		NumVerifications: 10,
		NumConversations: 10,
	}
	
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("No configuration file found. Using defaults.")
		return config
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("Error reading configuration file: %v\n", err)
		return config
	}
	
	fmt.Println("Configuration loaded successfully.")
	return config
}

func saveConfiguration(config Configuration) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling configuration: %v\n", err)
		return
	}
	
	if err := ioutil.WriteFile(configFile, data, 0644); err != nil {
		fmt.Printf("Error saving configuration: %v\n", err)
		return
	}
	
	fmt.Println("Configuration saved successfully.")
}

func listAndSelectTables(ctx context.Context, client *dynamodb.Client, config *Configuration, reader *bufio.Reader) {
	fmt.Println("\nFetching DynamoDB tables...")
	
	// List all tables
	resp, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		fmt.Printf("Error listing tables: %v\n", err)
		return
	}
	
	if len(resp.TableNames) == 0 {
		fmt.Println("No tables found in this region.")
		return
	}
	
	// Display tables
	fmt.Println("\nAvailable DynamoDB Tables:")
	for i, table := range resp.TableNames {
		fmt.Printf("%d. %s\n", i+1, table)
	}
	
	// Map tables to our logical tables
	tableTypes := []struct {
		name        string
		description string
		setter      func(string)
	}{
		{"LayoutMetadata", "Stores layout information and product mapping", func(t string) { config.LayoutMetadataTable = t }},
		{"VerificationResults", "Stores verification results and discrepancies", func(t string) { config.VerificationResultsTable = t }},
		{"ConversationHistory", "Stores conversation history with Bedrock", func(t string) { config.ConversationHistoryTable = t }},
	}
	
	// For each logical table, ask user to select the actual DynamoDB table
	for _, tableType := range tableTypes {
		for {
			fmt.Printf("\nSelect table for %s (%s): \n", tableType.name, tableType.description)
			fmt.Println("0. Skip/None")
			for i, table := range resp.TableNames {
				fmt.Printf("%d. %s\n", i+1, table)
			}
			fmt.Print("Enter number: ")
			
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			
			idx, err := strconv.Atoi(input)
			if err != nil || idx < 0 || idx > len(resp.TableNames) {
				fmt.Println("Invalid selection. Please try again.")
				continue
			}
			
			if idx == 0 {
				tableType.setter("") // Skip this table
			} else {
				tableType.setter(resp.TableNames[idx-1])
			}
			break
		}
	}
	
	fmt.Println("\nTable selection complete.")
}

func viewConfiguration(config Configuration) {
	fmt.Println("\nCurrent Configuration:")
	fmt.Printf("AWS Region: %s\n", config.Region)
	fmt.Printf("Layout Metadata Table: %s\n", config.LayoutMetadataTable)
	fmt.Printf("Verification Results Table: %s\n", config.VerificationResultsTable)
	fmt.Printf("Conversation History Table: %s\n", config.ConversationHistoryTable)
	fmt.Printf("Number of Layouts to Generate: %d\n", config.NumLayouts)
	fmt.Printf("Number of Verifications to Generate: %d\n", config.NumVerifications)
	fmt.Printf("Number of Conversations to Generate: %d\n", config.NumConversations)
}

func viewTableSchema(ctx context.Context, client *dynamodb.Client, config Configuration, reader *bufio.Reader) {
	tables := map[string]string{
		"1. Layout Metadata":      config.LayoutMetadataTable,
		"2. Verification Results": config.VerificationResultsTable,
		"3. Conversation History": config.ConversationHistoryTable,
	}
	
	fmt.Println("\nSelect table to view schema:")
	for option, table := range tables {
		if table != "" {
			fmt.Printf("%s (%s)\n", option, table)
		} else {
			fmt.Printf("%s (not configured)\n", option)
		}
	}
	fmt.Println("4. Return to main menu")
	fmt.Print("Enter choice: ")
	
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	var tableName string
	switch choice {
	case "1":
		tableName = config.LayoutMetadataTable
	case "2":
		tableName = config.VerificationResultsTable
	case "3":
		tableName = config.ConversationHistoryTable
	case "4":
		return
	default:
		fmt.Println("Invalid choice.")
		return
	}
	
	if tableName == "" {
		fmt.Println("Table not configured.")
		return
	}
	
	// Get table description
	result, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		fmt.Printf("Error describing table: %v\n", err)
		return
	}
	
	tableDesc := result.Table
	
	// Extract schema info
	schemaInfo := TableSchemaInfo{
		TableName: *tableDesc.TableName,
		ApproximateItemCount: *tableDesc.ItemCount, // Dereference the pointer
	}
	
	// Extract key schema
	for _, key := range tableDesc.KeySchema {
		if key.KeyType == types.KeyTypeHash {
			schemaInfo.PrimaryKey = *key.AttributeName
		} else if key.KeyType == types.KeyTypeRange {
			schemaInfo.SortKey = *key.AttributeName
		}
	}
	
	// Extract secondary indexes
	for _, index := range tableDesc.GlobalSecondaryIndexes {
		schemaInfo.SecondaryIndexes = append(schemaInfo.SecondaryIndexes, *index.IndexName)
	}
	
	// Print schema info
	fmt.Println("\nTable Schema Information:")
	fmt.Printf("Table Name: %s\n", schemaInfo.TableName)
	fmt.Printf("Partition Key: %s\n", schemaInfo.PrimaryKey)
	if schemaInfo.SortKey != "" {
		fmt.Printf("Sort Key: %s\n", schemaInfo.SortKey)
	}
	fmt.Printf("Approximate Item Count: %d\n", schemaInfo.ApproximateItemCount)
	
	if len(schemaInfo.SecondaryIndexes) > 0 {
		fmt.Println("\nGlobal Secondary Indexes:")
		for _, idx := range schemaInfo.SecondaryIndexes {
			fmt.Printf("- %s\n", idx)
		}
	}
	
	// Print attribute definitions
	fmt.Println("\nAttribute Definitions:")
	for _, attr := range tableDesc.AttributeDefinitions {
		fmt.Printf("- %s: %s\n", *attr.AttributeName, attr.AttributeType)
	}
	
	fmt.Println("\nPress Enter to continue...")
	reader.ReadString('\n')
}

func populateTestData(ctx context.Context, client *dynamodb.Client, config Configuration, reader *bufio.Reader) {
	if config.LayoutMetadataTable == "" && config.VerificationResultsTable == "" && config.ConversationHistoryTable == "" {
		fmt.Println("No tables configured. Please configure tables first.")
		return
	}
	
	fmt.Println("\nSelect tables to populate with test data:")
	fmt.Println("1. All configured tables")
	
	options := []string{}
	if config.LayoutMetadataTable != "" {
		options = append(options, fmt.Sprintf("2. Layout Metadata only (%s)", config.LayoutMetadataTable))
	}
	if config.VerificationResultsTable != "" {
		options = append(options, fmt.Sprintf("3. Verification Results only (%s)", config.VerificationResultsTable))
	}
	if config.ConversationHistoryTable != "" {
		options = append(options, fmt.Sprintf("4. Conversation History only (%s)", config.ConversationHistoryTable))
	}
	
	for _, option := range options {
		fmt.Println(option)
	}
	
	fmt.Println("5. Return to main menu")
	fmt.Print("Enter choice: ")
	
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	switch choice {
	case "1":
		if config.LayoutMetadataTable != "" {
			fmt.Println("Calling populate_test_data with layout table...")
			callPopulateTestData(config, "layout")
		}
		if config.VerificationResultsTable != "" {
			fmt.Println("Calling populate_test_data with verification table...")
			callPopulateTestData(config, "verification")
		}
		if config.ConversationHistoryTable != "" {
			fmt.Println("Calling populate_test_data with conversation table...")
			callPopulateTestData(config, "conversation")
		}
	case "2":
		if config.LayoutMetadataTable != "" {
			fmt.Println("Calling populate_test_data with layout table...")
			callPopulateTestData(config, "layout")
		} else {
			fmt.Println("Layout Metadata table not configured.")
		}
	case "3":
		if config.VerificationResultsTable != "" {
			fmt.Println("Calling populate_test_data with verification table...")
			callPopulateTestData(config, "verification")
		} else {
			fmt.Println("Verification Results table not configured.")
		}
	case "4":
		if config.ConversationHistoryTable != "" {
			fmt.Println("Calling populate_test_data with conversation table...")
			callPopulateTestData(config, "conversation")
		} else {
			fmt.Println("Conversation History table not configured.")
		}
	case "5":
		return
	default:
		fmt.Println("Invalid choice.")
		return
	}
}

func callPopulateTestData(config Configuration, tableType string) {
	// Call the populate_test_data function directly
	fmt.Printf("Executing populate_test_data with region %s and table-type %s\n", config.Region, tableType)
	
	// Import the populate_data package and call its function
	populate_data.RunPopulateTestData(
		configFile,           // configFilePath
		config.Region,        // regionName
		tableType,            // tableTypeStr
		config.NumLayouts,    // layoutCount
		config.NumVerifications, // verificationCount
		config.NumConversations, // conversationCount
		true,                 // verboseLogging
		config.LayoutMetadataTable,      // layoutMetadataTable
		config.VerificationResultsTable, // verificationResultsTable
		config.ConversationHistoryTable, // conversationHistoryTable
	)
}

func changeRegion(config *Configuration, reader *bufio.Reader) {
	fmt.Printf("\nCurrent region: %s\n", config.Region)
	fmt.Print("Enter new region (e.g., us-east-1, us-west-2): ")
	
	region, _ := reader.ReadString('\n')
	region = strings.TrimSpace(region)
	
	if region != "" {
		config.Region = region
		fmt.Printf("Region updated to: %s\n", config.Region)
	}
}

func configureDataCounts(config *Configuration, reader *bufio.Reader) {
	fmt.Println("\nConfigure Data Generation Counts")
	
	fmt.Printf("Number of layouts to generate (current: %d): ", config.NumLayouts)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		if num, err := strconv.Atoi(input); err == nil && num > 0 {
			config.NumLayouts = num
		} else {
			fmt.Println("Invalid number. Keeping current value.")
		}
	}
	
	fmt.Printf("Number of verifications to generate (current: %d): ", config.NumVerifications)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		if num, err := strconv.Atoi(input); err == nil && num > 0 {
			config.NumVerifications = num
		} else {
			fmt.Println("Invalid number. Keeping current value.")
		}
	}
	
	fmt.Printf("Number of conversations to generate (current: %d): ", config.NumConversations)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		if num, err := strconv.Atoi(input); err == nil && num > 0 {
			config.NumConversations = num
		} else {
			fmt.Println("Invalid number. Keeping current value.")
		}
	}
	
	fmt.Println("Configuration updated.")
}