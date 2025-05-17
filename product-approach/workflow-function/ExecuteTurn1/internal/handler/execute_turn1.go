package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	//"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	//"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	//"github.com/google/uuid"

	"workflow-function/ExecuteTurn1/internal/config"
	"workflow-function/ExecuteTurn1/internal/dependencies"
	"workflow-function/ExecuteTurn1/internal/models"
)

// Handler is the main handler for ExecuteTurn1
type Handler struct {
	bedrockClient    *bedrockruntime.Client
	s3Client         *s3.Client
	config           *config.Config
	responseProcessor *ResponseProcessor
}

// NewHandler creates a new ExecuteTurn1 handler
func NewHandler(clients *dependencies.Clients) *Handler {
	return &Handler{
		bedrockClient:    clients.BedrockClient,
		s3Client:         clients.S3Client,
		config:           clients.Config,
		responseProcessor: NewResponseProcessor(
			clients.Config.ThinkingType != "",
			clients.Config.ThinkingBudgetTokens,
		),
	}
}

// HandleRequest handles the Lambda function request
func (h *Handler) HandleRequest(ctx context.Context, request models.ExecuteTurn1Request) (*models.ExecuteTurn1Response, error) {
	// Validate input
	if err := validateInput(request); err != nil {
		return createErrorResponse(request.WorkflowState, err, false), nil
	}

	// Create a hybrid Base64 retriever if enabled
	var retriever hybridBase64Retriever
	if h.config.EnableHybridStorage {
		retriever = &s3HybridBase64Retriever{
			client: h.s3Client,
			bucket: h.config.TempBase64Bucket,
			timeout: h.config.Base64RetrievalTimeout,
		}
	} else {
		retriever = &inlineBase64Retriever{}
	}

	// Call Bedrock and process response
	turnResponse, err := h.callBedrock(ctx, request.WorkflowState, retriever)
	if err != nil {
		structErr := ExtractBedrockError(err)
		return createErrorResponse(request.WorkflowState, structErr, structErr.Retryable), nil
	}

	// Update workflow state
	updatedState := h.responseProcessor.UpdateWorkflowState(
		&request.WorkflowState,
		turnResponse,
	)

	// Return successful response
	return &models.ExecuteTurn1Response{
		WorkflowState: *updatedState,
	}, nil
}

// callBedrock calls the Bedrock API with the prepared prompt
func (h *Handler) callBedrock(
	ctx context.Context,
	state models.WorkflowState,
	retriever hybridBase64Retriever,
) (*models.TurnResponse, error) {
	startTime := time.Now()

	// Prepare the Bedrock request
	bedrockRequest, err := h.prepareBedrockRequest(state, retriever)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare Bedrock request: %w", err)
	}

	// Marshal the request to JSON
	requestBytes, err := json.Marshal(bedrockRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Bedrock request: %w", err)
	}

	// Prepare the invoke model input
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(h.config.BedrockModelID),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
		Body:        requestBytes,
	}

	// Set a context timeout for the Bedrock API call
	bedrockCtx, cancel := context.WithTimeout(ctx, h.config.BedrockTimeout)
	defer cancel()

	// Call the Bedrock API
	output, err := h.bedrockClient.InvokeModel(bedrockCtx, input)
	if err != nil {
		// Handle error conditions
		return nil, err
	}

	// Calculate latency
	latencyMs := time.Since(startTime).Milliseconds()

	// Parse the response
	var bedrockResponse models.BedrockResponse
	if err := json.Unmarshal(output.Body, &bedrockResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Bedrock response: %w", err)
	}

	// Process the response
	turnResponse, err := h.responseProcessor.ProcessResponse(
		bedrockResponse,
		latencyMs,
		state.Stage,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to process Bedrock response: %w", err)
	}

	return turnResponse, nil
}

// prepareBedrockRequest prepares a request to the Bedrock API
func (h *Handler) prepareBedrockRequest(
	state models.WorkflowState,
	retriever hybridBase64Retriever,
) (*models.BedrockRequest, error) {
	// Get current prompt
	prompt := state.CurrentPrompt

	// Handle different types of prompts (text vs messages)
	if len(prompt.Messages) > 0 {
		return h.prepareBedrockMessagesRequest(prompt, state.ImageData, retriever)
	} else if prompt.PromptText != "" {
		return h.prepareBedrockTextRequest(prompt, state.ImageData, retriever)
	}

	return nil, fmt.Errorf("no prompt text or messages found in CurrentPrompt")
}

// prepareBedrockMessagesRequest prepares a Bedrock request using the messages format
func (h *Handler) prepareBedrockMessagesRequest(
	prompt models.PromptDetails,
	imageData models.ImageDetails,
	retriever hybridBase64Retriever,
) (*models.BedrockRequest, error) {
	// Clone the messages to avoid modifying the original
	messages := make([]models.BedrockMessage, len(prompt.Messages))
	copy(messages, prompt.Messages)

	// Process any image message content
	for i := range messages {
		for j := range messages[i].Content {
			content := &messages[i].Content[j]
			if content.Type == "image" && content.Source != nil && content.Source.Type == "base64" {
				// If image data is inline in the message, no need to retrieve
				if content.Source.Data != "" {
					continue
				}

				// Get the Base64 data from the image details
				base64Data, err := retriever.retrieveBase64(imageData)
				if err != nil {
					return nil, fmt.Errorf("failed to retrieve base64 data: %w", err)
				}

				// Add the data to the message content
				content.Source.Data = base64Data
				if content.Source.MediaType == "" {
					content.Source.MediaType = determineMediaType(imageData.ContentType)
				}
			}
		}
	}

	// Create the Bedrock request
	return &models.BedrockRequest{
		AnthropicVersion: h.config.AnthropicVersion,
		Messages:         messages,
		MaxTokens:        h.config.MaxTokens,
		Temperature:      h.config.Temperature,
		// No system message for messages format
	}, nil
}

// prepareBedrockTextRequest prepares a Bedrock request using the legacy text format
func (h *Handler) prepareBedrockTextRequest(
	prompt models.PromptDetails,
	imageData models.ImageDetails,
	retriever hybridBase64Retriever,
) (*models.BedrockRequest, error) {
	// Get the Base64 data
	base64Data, err := retriever.retrieveBase64(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve base64 data: %w", err)
	}

	// Create a user message with text and image
	userMessage := models.BedrockMessage{
		Role: "user",
		Content: []models.MessageContent{
			{
				Type: "image",
				Source: &models.ImageSource{
					Type:      "base64",
					MediaType: determineMediaType(imageData.ContentType),
					Data:      base64Data,
				},
			},
			{
				Type: "text",
				Text: prompt.PromptText,
			},
		},
	}

	// Create the Bedrock request
	return &models.BedrockRequest{
		AnthropicVersion: h.config.AnthropicVersion,
		Messages:         []models.BedrockMessage{userMessage},
		MaxTokens:        h.config.MaxTokens,
		Temperature:      h.config.Temperature,
		// No system message for text format
	}, nil
}

// hybridBase64Retriever is an interface for retrieving Base64 data
type hybridBase64Retriever interface {
	retrieveBase64(imageData models.ImageDetails) (string, error)
}

// inlineBase64Retriever retrieves Base64 data directly from the image details
type inlineBase64Retriever struct{}

// retrieveBase64 retrieves Base64 data from the image details
func (r *inlineBase64Retriever) retrieveBase64(imageData models.ImageDetails) (string, error) {
	if imageData.ImageBase64 != "" {
		return imageData.ImageBase64, nil
	}
	return "", fmt.Errorf("no base64 data found in image details")
}

// s3HybridBase64Retriever retrieves Base64 data from S3
type s3HybridBase64Retriever struct {
	client  *s3.Client
	bucket  string
	timeout time.Duration
}

// retrieveBase64 retrieves Base64 data from S3 if available, otherwise from the image details
func (r *s3HybridBase64Retriever) retrieveBase64(imageData models.ImageDetails) (string, error) {
	// If inline Base64 is available, use it
	if imageData.ImageBase64 != "" {
		return imageData.ImageBase64, nil
	}

	// If S3 key is available, retrieve from S3
	if imageData.ImageS3Key != "" {
		s3Bucket := imageData.ImageS3Bucket
		if s3Bucket == "" {
			s3Bucket = r.bucket
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		defer cancel()

		// Get object from S3
		output, err := r.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(imageData.ImageS3Key),
		})
		if err != nil {
			return "", fmt.Errorf("failed to get object from S3: %w", err)
		}
		defer output.Body.Close()

		// Read the object body
		data, err := io.ReadAll(output.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read object body: %w", err)
		}

		// Encode as Base64
		return base64.StdEncoding.EncodeToString(data), nil
	}

	return "", fmt.Errorf("no base64 data or S3 key found in image details")
}

// validateInput validates the incoming request
func validateInput(request models.ExecuteTurn1Request) error {
	state := request.WorkflowState

	if state.VerificationID == "" {
		return fmt.Errorf("verificationId is required")
	}

	if state.CurrentPrompt.PromptID == "" {
		return fmt.Errorf("currentPrompt.promptId is required")
	}

	if state.CurrentPrompt.PromptText == "" && len(state.CurrentPrompt.Messages) == 0 {
		return fmt.Errorf("either promptText or messages must be provided")
	}

	if state.ImageData.ImageID == "" {
		return fmt.Errorf("imageData.imageId is required")
	}

	return nil
}

// createErrorResponse creates an error response
func createErrorResponse(
	state models.WorkflowState,
	err error,
	retryable bool,
) *models.ExecuteTurn1Response {
	var structErr *models.Error
	if e, ok := err.(*models.Error); ok {
		structErr = e
	} else {
		structErr = &models.Error{
			Code:      "EXECUTION_ERROR",
			Message:   err.Error(),
			Retryable: retryable,
			Context:   make(map[string]string),
		}
	}

	// Update workflow state
	state.Status = "ERROR"
	state.Timestamp = models.FormatISO8601()

	return &models.ExecuteTurn1Response{
		WorkflowState: state,
		Error:         structErr,
	}
}

// determineMediaType determines the media type from a content type
func determineMediaType(contentType string) string {
	if contentType == "" {
		return "image/jpeg" // Default to JPEG
	}

	// Clean up content type
	contentType = strings.TrimSpace(contentType)
	contentType = strings.ToLower(contentType)

	// Handle common content types
	switch {
	case strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg"):
		return "image/jpeg"
	case strings.Contains(contentType, "png"):
		return "image/png"
	case strings.Contains(contentType, "gif"):
		return "image/gif"
	case strings.Contains(contentType, "webp"):
		return "image/webp"
	default:
		return contentType
	}
}