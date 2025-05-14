// Package main provides the entry point for the initialization Lambda function
// This file contains adapters for the shared dbutils package
package main

import (
	"context"
	"workflow-function/shared/dbutils"
	"workflow-function/shared/schema"
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

// DynamoDBUtils type alias for backward compatibility
type DynamoDBUtils = dbutils.DynamoDBUtils

// Wrapper function for backward compatibility
func NewDynamoDBUtils(client interface{}, logger Logger) *DynamoDBUtils {
	// Create a temporary config - it will be replaced when SetConfig is called
	config := dbutils.Config{
		VerificationTable: "",
		LayoutTable:       "",
		DefaultTTLDays:    30,
	}
	
	// Use the dynamic client from dependencies.go
	return dbutils.New(client, logger, config)
}

// SetConfig is a compatibility method that will update the dbutils config
func (d *DynamoDBUtils) SetConfig(config ConfigVars) {
	// This is just a stub for backward compatibility
	// The actual config setting is handled in dependencies.go
}