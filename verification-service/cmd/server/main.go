package main

import (
	"context"
	"fmt"
	//"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"verification-service/internal/api"
	"verification-service/internal/api/handlers"
	"verification-service/internal/app/services"
	"verification-service/internal/config"
	"verification-service/internal/domain/engines"
	"verification-service/internal/infrastructure/bedrock"
	"verification-service/internal/infrastructure/dynamodb"
	"verification-service/internal/infrastructure/logger"
	"verification-service/internal/infrastructure/s3"
)

func main() {
	// Initialize logger
	logger := logger.NewLogger()
	logger.Info("Starting verification service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", err)
	}

	// Initialize infrastructure
	bedrockProvider, err := bedrock.NewBedrockProvider(
		cfg.AWS.Region,
		cfg.Bedrock.ModelID,
		cfg.Bedrock.MaxRetries,
	)
	if err != nil {
		logger.Fatal("Failed to initialize Bedrock provider", err)
	}

	verificationRepo, err := dynamodb.NewVerificationRepository(
		cfg.AWS.Region,
		cfg.DynamoDB.VerificationResultsTable,
		cfg.DynamoDB.LayoutMetadataTable,
	)
	if err != nil {
		logger.Fatal("Failed to initialize DynamoDB repository", err)
	}

	imageService, err := s3.NewImageService(
		cfg.AWS.Region,
		cfg.S3.ReferenceBucket,
		cfg.S3.CheckingBucket,
		cfg.S3.ResultsBucket,
	)
	if err != nil {
		logger.Fatal("Failed to initialize S3 image service", err)
	}

	// Initialize domain engines
	verificationEngine := engines.NewVerificationEngine(
		bedrockProvider,
	)

	promptGenerator := engines.NewPromptGenerator()
	responseAnalyzer := engines.NewResponseAnalyzer()

	// Initialize services
	visualizationService := services.NewVisualizationService(imageService)
	notificationService := services.NewNotificationService(cfg.Notification)

	verificationService := services.NewVerificationService(
		verificationRepo,
		imageService,
		verificationEngine,
		promptGenerator,
		responseAnalyzer,
		bedrockProvider,
		visualizationService,
		notificationService,
	)

	// Initialize handlers
	verificationHandler := handlers.NewVerificationHandler(verificationService)
	healthHandler := handlers.NewHealthHandler()

	// Initialize router
	router := api.NewRouter(verificationHandler, healthHandler)

	// Configure HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSecs) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSecs) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSecs) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info(fmt.Sprintf("Server listening on port %d", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown gracefully
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", err)
	}

	logger.Info("Server exited gracefully")
}