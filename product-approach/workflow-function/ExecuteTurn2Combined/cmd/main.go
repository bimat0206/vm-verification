package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	internalConfig "workflow-function/ExecuteTurn2Combined/internal/config"
	"workflow-function/ExecuteTurn2Combined/internal/handler"
	"workflow-function/ExecuteTurn2Combined/internal/services"
	"workflow-function/shared/errors"
	"workflow-function/shared/logger"
)

// Application container wires dependencies for the Lambda
type applicationContainer struct {
	cfg *internalConfig.Config
	log logger.Logger
	hnd *handler.Handler
}

var containerInstance *applicationContainer

func init() {
	cfg, err := internalConfig.LoadConfiguration()
	if err != nil {
		if wfErr, ok := err.(*errors.WorkflowError); ok {
			out, _ := json.Marshal(wfErr)
			fmt.Fprintf(os.Stderr, "%s\n", out)
		}
		log.Fatalf("configuration error: %v", err)
	}

	logr := logger.New("ExecuteTurn2Combined", "main")

	// Initialize services
	s3Mgr, err := services.NewS3StateManager(*cfg, logr)
	if err != nil {
		log.Fatalf("failed to init s3 manager: %v", err)
	}

	bedrockSvc, err := services.NewBedrockService(context.Background(), *cfg)
	if err != nil {
		log.Fatalf("failed to init bedrock service: %v", err)
	}
	dynamoSvc, err := services.NewDynamoDBService(cfg)
	if err != nil {
		log.Fatalf("failed to init dynamodb service: %v", err)
	}

	promptSvc, err := services.NewPromptService(cfg)
	if err != nil {
		log.Fatalf("failed to init prompt service: %v", err)
	}

	h, err := handler.NewHandler(s3Mgr, bedrockSvc, dynamoSvc, promptSvc, logr, cfg)
	if err != nil {
		log.Fatalf("failed to init handler: %v", err)
	}

	containerInstance = &applicationContainer{cfg: cfg, log: logr, hnd: h}
}

func main() {
	lambda.Start(func(ctx context.Context, event json.RawMessage) (*handler.StepFunctionResponse, error) {
		var req handler.Turn2Request
		if err := json.Unmarshal(event, &req); err != nil {
			return nil, errors.NewInternalError("json_parse", err)
		}
		return containerInstance.hnd.Handle(ctx, &req)
	})
}
