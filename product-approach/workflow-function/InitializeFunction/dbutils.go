// Package main provides the entry point for the initialization Lambda function
// This file contains adapters for the shared dbutils package
package main

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"workflow-function/shared/dbutils"
)

// ConfigVars holds environment configuration for legacy compatibility
type ConfigVars struct {
	LayoutTable        string
	VerificationTable  string
	VerificationPrefix string
	ReferenceBucket    string
	CheckingBucket     string
}

// Convert our ConfigVars to dbutils.Config
func convertConfig(config ConfigVars) dbutils.Config {
	return dbutils.Config{
		VerificationTable: config.VerificationTable,
		LayoutTable:       config.LayoutTable,
		DefaultTTLDays:    30, // Default TTL is 30 days
	}
}

// DynamoDBVerificationItem alias for backward compatibility
// This legacy type will be removed once other components are refactored
type DynamoDBVerificationItem struct {
	VerificationId         string `dynamodbav:"verificationId"`
	VerificationAt         string `dynamodbav:"verificationAt"`
	Status                 string `dynamodbav:"status"`
	VerificationType       string `dynamodbav:"verificationType"`
	VendingMachineId       string `dynamodbav:"vendingMachineId,omitempty"`
	LayoutId               int    `dynamodbav:"layoutId,omitempty"`
	LayoutPrefix           string `dynamodbav:"layoutPrefix,omitempty"`
	PreviousVerificationId string `dynamodbav:"previousVerificationId,omitempty"`
	ReferenceImageUrl      string `dynamodbav:"referenceImageUrl"`
	CheckingImageUrl       string `dynamodbav:"checkingImageUrl"`
	RequestId              string `dynamodbav:"requestId,omitempty"`
	NotificationEnabled    bool   `dynamodbav:"notificationEnabled"`
	TTL                    int64  `dynamodbav:"ttl,omitempty"`
	SchemaVersion          string `dynamodbav:"schemaVersion,omitempty"`
}

// DynamoDBUtilsWrapper wraps the shared package for local extension methods
type DynamoDBUtilsWrapper struct {
	*dbutils.DynamoDBUtils
}

// NewDynamoDBUtils creates a new DynamoDBUtilsWrapper instance
func NewDynamoDBUtils(client *dynamodb.Client, logger Logger) *dbutils.DynamoDBUtils {
	// Create a temporary config - it will be replaced when SetConfig is called
	config := dbutils.Config{
		VerificationTable: "",
		LayoutTable:       "",
		DefaultTTLDays:    30,
	}
	
	// Return the dbutils instance directly, we don't need to wrap it in most cases
	return dbutils.New(client, logger, config)
}