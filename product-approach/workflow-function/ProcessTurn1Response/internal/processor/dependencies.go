package processor

import (
	"workflow-function/ProcessTurn1Response/internal/dependencies"
	"workflow-function/ProcessTurn1Response/internal/storage"
	"workflow-function/shared/logger"
)

// GetStorageDependencies creates storage dependencies from global dependencies
func GetStorageDependencies(log logger.Logger) *storage.Dependencies {
	// Get global dependencies
	deps := dependencies.GlobalDependencies
	
	if deps == nil {
		// If global dependencies are not initialized, create a new instance
		deps = dependencies.NewDependencies(log)
		dependencies.GlobalDependencies = deps
	}
	
	// Create simplified storage dependencies - only include DBManager for conversation history
	return &storage.Dependencies{
		Logger:    log,
		DBManager: deps.DBManager,
	}
}
