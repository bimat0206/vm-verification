package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"

	"workflow-function/FinalizeWithErrorFunction/internal/config"
	"workflow-function/FinalizeWithErrorFunction/internal/handler"
	"workflow-function/FinalizeWithErrorFunction/internal/models"
	"workflow-function/shared/logger"
	"workflow-function/shared/s3state"
)

var (
	appHandler *handler.ErrorHandler
)

func init() {
	cfg := config.LoadConfiguration()
	log := logger.New("kootoro-verification", "FinalizeWithErrorFunction")
	mgr, err := s3state.New(cfg.AWS.S3Bucket)
	if err != nil {
		log.Error("failed to create s3 manager", map[string]interface{}{"error": err.Error()})
	}

	appHandler = handler.NewErrorHandler(log, mgr)
}

func HandleRequest(ctx context.Context, event json.RawMessage) (*models.FinalizeWithErrorOutput, error) {
	var input models.FinalizeWithErrorInput
	if err := json.Unmarshal(event, &input); err != nil {
		return nil, err
	}
	return appHandler.ProcessError(ctx, &input)
}

func main() {
	lambda.Start(HandleRequest)
}
