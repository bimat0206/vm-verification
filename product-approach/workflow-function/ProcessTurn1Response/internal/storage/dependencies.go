package storage

import (
	"workflow-function/shared/logger"
)

// Dependencies holds all AWS service clients and utilities
type Dependencies struct {
	Logger    logger.Logger
	S3Manager *S3Manager
	DBManager *DBManager
}

// NewDependencies creates a new Dependencies instance
func NewDependencies(log logger.Logger, s3Manager *S3Manager, dbManager *DBManager) *Dependencies {
	return &Dependencies{
		Logger:    log,
		S3Manager: s3Manager,
		DBManager: dbManager,
	}
}

// GlobalDependencies is a placeholder for the global dependencies instance
var GlobalDependencies *Dependencies
