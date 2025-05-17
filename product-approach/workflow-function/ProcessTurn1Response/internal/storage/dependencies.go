package storage

import (
	"workflow-function/shared/logger"
)

// Dependencies holds storage-related dependencies
type Dependencies struct {
	Logger    logger.Logger
	DBManager *DBManager // Only needed for minimal conversation history updates
}

// NewDependencies creates a new Dependencies instance
func NewDependencies(log logger.Logger, dbManager *DBManager) *Dependencies {
	return &Dependencies{
		Logger:    log,
		DBManager: dbManager,
	}
}

// GlobalDependencies is a placeholder for the global dependencies instance
var GlobalDependencies *Dependencies
