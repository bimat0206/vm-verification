// Package populate_data provides functions for populating DynamoDB tables with test data
package populate_data

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	// Renamed import to avoid conflict with variable name
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
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

// RunPopulateTestData is the main entry point that can be called from other packages
func RunPopulateTestData(configFilePath string, regionName string, tableTypeStr string, 
                        layoutCount int, verificationCount int, conversationCount int, verboseLogging bool,
                        layoutTableName string, verificationTableName string, conversationTableName string) {
	// Set up configuration
	configFile := configFilePath
	region := regionName
	tableType := tableTypeStr
	numLayouts := layoutCount
	numVerifications := verificationCount
	numConversations := conversationCount
	verbose := verboseLogging
	layoutTable := layoutTableName
	verificationTable := verificationTableName
	conversationTable := conversationTableName

	// Call the main function with these parameters
	populateTestDataInternal(configFile, region, tableType, numLayouts, numVerifications, numConversations, verbose,
		layoutTable, verificationTable, conversationTable)
}

// Internal implementation that both entry points use
func populateTestDataInternal(configFile string, region string, tableType string,
                             numLayouts int, numVerifications int, numConversations int, verbose bool,
							 layoutTable string, verificationTable string, conversationTable string) {

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Configure logging
	if verbose {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	// Load configuration
	config := loadConfiguration(configFile)

	// Override with command-line args if provided
	if region != "" {
		config.Region = region
	}
	if numLayouts > 0 {
		config.NumLayouts = numLayouts
	}
	if numVerifications > 0 {
		config.NumVerifications = numVerifications
	}
	if numConversations > 0 {
		config.NumConversations = numConversations
	}
	if layoutTable != "" {
		config.LayoutMetadataTable = layoutTable
	}
	if verificationTable != "" {
		config.VerificationResultsTable = verificationTable
	}
	if conversationTable != "" {
		config.ConversationHistoryTable = conversationTable
	}

	// Validate configuration
	if err := validateConfiguration(config, tableType); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create AWS session
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(config.Region),
	)
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	// Process based on table type
	switch tableType {
	case "all":
		if config.LayoutMetadataTable != "" {
			createLayoutMetadataTestData(client, config)
		}
		if config.VerificationResultsTable != "" {
			createVerificationResultsTestData(client, config)
		}
		if config.ConversationHistoryTable != "" {
			createConversationHistoryTestData(client, config)
		}
	case "layout":
		if config.LayoutMetadataTable != "" {
			createLayoutMetadataTestData(client, config)
		} else {
			log.Fatal("Layout metadata table not configured")
		}
	case "verification":
		if config.VerificationResultsTable != "" {
			createVerificationResultsTestData(client, config)
		} else {
			log.Fatal("Verification results table not configured")
		}
	case "conversation":
		if config.ConversationHistoryTable != "" {
			createConversationHistoryTestData(client, config)
		} else {
			log.Fatal("Conversation history table not configured")
		}
	default:
		log.Fatalf("Unknown table type: %s", tableType)
	}

	log.Println("Test data population complete")
}

func loadConfiguration(configFile string) Configuration {
	// Default configuration
	config := Configuration{
		Region:           "us-east-1",
		NumLayouts:       5,
		NumVerifications: 10,
		NumConversations: 10,
	}

	// Try to read configuration file
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("Warning: Could not read configuration file: %v", err)
		log.Println("Using default configuration instead")
		return config
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("Warning: Error parsing configuration file: %v", err)
		log.Println("Using default configuration instead")
		return config
	}

	log.Printf("Loaded configuration from %s", configFile)
	return config
}

func validateConfiguration(config Configuration, tableType string) error {
	switch tableType {
	case "all":
		if config.LayoutMetadataTable == "" && config.VerificationResultsTable == "" && config.ConversationHistoryTable == "" {
			return fmt.Errorf("no tables configured")
		}
	case "layout":
		if config.LayoutMetadataTable == "" {
			return fmt.Errorf("layout metadata table not configured")
		}
	case "verification":
		if config.VerificationResultsTable == "" {
			return fmt.Errorf("verification results table not configured")
		}
	case "conversation":
		if config.ConversationHistoryTable == "" {
			return fmt.Errorf("conversation history table not configured")
		}
	default:
		return fmt.Errorf("unknown table type: %s", tableType)
	}

	// Validate AWS region
	awsRegionPattern := regexp.MustCompile(`^[a-z]{2}-[a-z]+-\d+$`)
	if !awsRegionPattern.MatchString(config.Region) {
		return fmt.Errorf("invalid AWS region format: %s", config.Region)
	}

	return nil
}

// Helper function to generate random strings
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Placeholder functions that would be implemented in the full version
func createLayoutMetadataTestData(client *dynamodb.Client, config Configuration) {
	log.Printf("Creating layout metadata test data in table: %s", config.LayoutMetadataTable)
	// Implementation would go here
}

func createVerificationResultsTestData(client *dynamodb.Client, config Configuration) {
	log.Printf("Creating verification results test data in table: %s", config.VerificationResultsTable)
	// Implementation would go here
}

func createConversationHistoryTestData(client *dynamodb.Client, config Configuration) {
	log.Printf("Creating conversation history test data in table: %s", config.ConversationHistoryTable)
	// Implementation would go here
}