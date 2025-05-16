# Lambda Function Migration Examples

This document provides concrete examples of how to migrate existing Lambda functions to support Base64-encoded images for Bedrock API calls.

## FetchImages Function Migration

### Before (S3 Only)
```go
package main

import (
    "context"
    "encoding/json"
    
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/service/s3"
    "workflow-function/shared/schema"
)

func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
    // Parse input
    var workflowState schema.WorkflowState
    json.Unmarshal(event, &workflowState)
    
    // Get image metadata from S3
    referenceInfo := &schema.ImageInfo{
        URL:      workflowState.VerificationContext.ReferenceImageUrl,
        S3Key:    extractS3Key(workflowState.VerificationContext.ReferenceImageUrl),
        S3Bucket: "reference-bucket",
    }
    
    checkingInfo := &schema.ImageInfo{
        URL:      workflowState.VerificationContext.CheckingImageUrl,
        S3Key:    extractS3Key(workflowState.VerificationContext.CheckingImageUrl),
        S3Bucket: "checking-bucket",
    }
    
    workflowState.Images = &schema.ImageData{
        Reference: referenceInfo,
        Checking:  checkingInfo,
    }
    
    return workflowState, nil
}
```

### After (S3 + Base64)
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/service/s3"
    "workflow-function/shared/schema"
)

func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
    // Parse input
    var workflowState schema.WorkflowState
    json.Unmarshal(event, &workflowState)
    
    // Initialize S3 client
    s3Client := s3.New(session.Must(session.NewSession()))
    
    // Fetch and process reference image
    referenceInfo, err := fetchAndProcessImage(
        s3Client,
        workflowState.VerificationContext.ReferenceImageUrl,
        "reference-bucket",
    )
    if err != nil {
        return nil, fmt.Errorf("failed to process reference image: %w", err)
    }
    
    // Fetch and process checking image  
    checkingInfo, err := fetchAndProcessImage(
        s3Client,
        workflowState.VerificationContext.CheckingImageUrl,
        "checking-bucket",
    )
    if err != nil {
        return nil, fmt.Errorf("failed to process checking image: %w", err)
    }
    
    // Create ImageData with Base64 support
    imageData := &schema.ImageData{
        Reference:       referenceInfo,
        Checking:        checkingInfo,
        Base64Generated: true,
        ProcessedAt:     schema.FormatISO8601(),
    }
    
    // Validate for Bedrock API
    if err := schema.ImageProcessor.ValidateForBedrock(imageData); err != nil {
        return nil, fmt.Errorf("image validation failed: %w", err)
    }
    
    workflowState.Images = imageData
    workflowState.VerificationContext.Status = schema.StatusImagesFetched
    
    return workflowState, nil
}

func fetchAndProcessImage(s3Client *s3.S3, imageUrl, bucket string) (*schema.ImageInfo, error) {
    // Extract S3 key from URL
    s3Key := extractS3Key(imageUrl)
    
    // Get object metadata
    headOutput, err := s3Client.HeadObject(&s3.HeadObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(s3Key),
    })
    if err != nil {
        return nil, err
    }
    
    // Get object data
    getOutput, err := s3Client.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(s3Key),
    })
    if err != nil {
        return nil, err
    }
    defer getOutput.Body.Close()
    
    // Read image bytes
    imageBytes, err := ioutil.ReadAll(getOutput.Body)
    if err != nil {
        return nil, err
    }
    
    // Build ImageInfo with Base64 data
    imageInfo, err := schema.NewImageInfoBuilder().
        WithS3Info(imageUrl, s3Key, bucket).
        WithImageData(imageBytes, aws.StringValue(headOutput.ContentType), s3Key).
        WithS3Metadata(
            headOutput.LastModified.Format(time.RFC3339),
            aws.StringValue(headOutput.ETag),
        ).
        Build()
    
    if err != nil {
        return nil, fmt.Errorf("failed to build image info: %w", err)
    }
    
    return imageInfo, nil
}
```

## PrepareTurn1Prompt Function Migration

### Before (Text Only)
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/aws/aws-lambda-go/lambda"
    "workflow-function/shared/schema"
)

func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
    var workflowState schema.WorkflowState
    json.Unmarshal(event, &workflowState)
    
    // Create text prompt
    promptText := fmt.Sprintf(`
    Please analyze the FIRST image (Reference Image)
    Focus exclusively on analyzing this Reference Image in detail
    according to the system prompt instructions.
    `)
    
    currentPrompt := &schema.CurrentPrompt{
        Text:         promptText,
        TurnNumber:   1,
        IncludeImage: "reference",
        PromptId:     fmt.Sprintf("prompt-%s-turn1", workflowState.VerificationContext.VerificationId),
        CreatedAt:    schema.FormatISO8601(),
    }
    
    workflowState.CurrentPrompt = currentPrompt
    workflowState.VerificationContext.Status = schema.StatusTurn1PromptReady
    
    return workflowState, nil
}
```

### After (Bedrock Messages)
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/aws/aws-lambda-go/lambda"
    "workflow-function/shared/schema"
)

func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
    var workflowState schema.WorkflowState
    json.Unmarshal(event, &workflowState)
    
    // Create prompt text
    promptText := buildTurn1PromptText(workflowState.VerificationContext)
    
    // Build CurrentPrompt with Bedrock messages
    currentPrompt := schema.NewCurrentPromptBuilder(1).
        WithIncludeImage("reference").
        WithBedrockMessages(promptText, workflowState.Images).
        WithText(promptText). // Keep for backward compatibility
        WithMetadata("verificationType", workflowState.VerificationContext.VerificationType).
        Build()
    
    // Validate Bedrock messages
    if errors := schema.ValidateCurrentPrompt(currentPrompt, true); len(errors) > 0 {
        return nil, fmt.Errorf("prompt validation failed: %v", errors)
    }
    
    workflowState.CurrentPrompt = currentPrompt
    workflowState.VerificationContext.Status = schema.StatusTurn1PromptReady
    
    return workflowState, nil
}

func buildTurn1PromptText(ctx *schema.VerificationContext) string {
    switch ctx.VerificationType {
    case schema.VerificationTypeLayoutVsChecking:
        return `Please analyze the FIRST image (Reference Image)
        
        Focus exclusively on analyzing this Reference Image in detail
        according to the system prompt instructions. Your goal is to identify
        the exact contents of all rows and slots.
        
        Important reminders:
        1. Row identification is CRITICAL
        2. Be thorough and descriptive in your analysis
        3. DO NOT compare with any other image at this stage
        4. Follow the ROW STATUS ANALYSIS format from the system prompt
        `
    case schema.VerificationTypePreviousVsCurrent:
        return `Please analyze the FIRST image (Previous State)
        
        The FIRST image shows the previous state of the vending machine.
        Analyze it to determine the baseline state for comparison.
        
        Focus on:
        1. Machine structure (rows, columns, layout)
        2. Product placements and quantities
        3. Empty positions
        4. Overall machine state
        `
    default:
        return "Please analyze the reference image according to the system prompt."
    }
}
```

## ExecuteTurn1 Function Migration

### Before (Generic API Call)
```go
package main

import (
    "context"
    "encoding/json"
    
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go/service/bedrock"
    "workflow-function/shared/schema"
)

func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
    var workflowState schema.WorkflowState
    json.Unmarshal(event, &workflowState)
    
    // Call Bedrock with text prompt
    bedrockClient := bedrock.New(session.Must(session.NewSession()))
    
    // Simple text-based API call (pseudo-code)
    response, err := bedrockClient.InvokeModel(&bedrock.InvokeModelInput{
        ModelId: aws.String("anthropic.claude-3-7-sonnet-20250219-v1:0"),
        Body:    []byte(workflowState.CurrentPrompt.Text),
    })
    
    if err != nil {
        return nil, err
    }
    
    // Process response
    workflowState.Turn1Response = map[string]interface{}{
        "content": string(response.Body),
        "timestamp": schema.FormatISO8601(),
    }
    
    return workflowState, nil
}
```

### After (Bedrock Converse API with Base64)
```go
package main

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
    "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
    "workflow-function/shared/schema"
)

func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
    var workflowState schema.WorkflowState
    json.Unmarshal(event, &workflowState)
    
    // Initialize Bedrock client
    cfg, _ := config.LoadDefaultConfig(ctx)
    bedrockClient := bedrockruntime.NewFromConfig(cfg)
    
    // Convert schema messages to Bedrock API format
    converseInput, err := buildConverseInput(workflowState)
    if err != nil {
        return nil, fmt.Errorf("failed to build converse input: %w", err)
    }
    
    // Call Bedrock Converse API
    startTime := time.Now()
    response, err := bedrockClient.Converse(ctx, converseInput)
    latency := time.Since(startTime).Milliseconds()
    
    if err != nil {
        return nil, fmt.Errorf("bedrock API call failed: %w", err)
    }
    
    // Process response
    turn1Response := &schema.TurnResponse{
        TurnId:    1,
        Timestamp: schema.FormatISO8601(),
        Prompt:    workflowState.CurrentPrompt.Text,
        ImageUrls: map[string]string{
            "reference": workflowState.Images.Reference.URL,
        },
        Response: schema.BedrockApiResponse{
            Content:    extractTextContent(response.Output),
            Thinking:   extractThinkingContent(response.Output), // if enabled
            StopReason: string(response.StopReason),
            ModelId:    *converseInput.ModelId,
        },
        LatencyMs:  latency,
        TokenUsage: extractTokenUsage(response.Usage),
        Stage:      "REFERENCE_ANALYSIS",
    }
    
    workflowState.Turn1Response = map[string]interface{}{
        "turnResponse": turn1Response,
    }
    workflowState.VerificationContext.Status = schema.StatusTurn1Completed
    
    return workflowState, nil
}

func buildConverseInput(workflowState schema.WorkflowState) (*bedrockruntime.ConverseInput, error) {
    // Validate that we have Bedrock messages
    if len(workflowState.CurrentPrompt.Messages) == 0 {
        return nil, fmt.Errorf("no Bedrock messages found in current prompt")
    }
    
    // Convert schema messages to Bedrock API format
    var messages []types.Message
    for _, schemaMsg := range workflowState.CurrentPrompt.Messages {
        var contentBlocks []types.ContentBlock
        
        for _, content := range schemaMsg.Content {
            switch content.Type {
            case "text":
                contentBlocks = append(contentBlocks, &types.ContentBlockMemberText{
                    Value: content.Text,
                })
            case "image":
                if content.Image != nil {
                    // Decode Base64 to bytes
                    imageBytes, err := base64.StdEncoding.DecodeString(content.Image.Source.Bytes)
                    if err != nil {
                        return nil, fmt.Errorf("failed to decode Base64 image: %w", err)
                    }
                    
                    contentBlocks = append(contentBlocks, &types.ContentBlockMemberImage{
                        Value: types.ImageBlock{
                            Format: types.ImageFormat(content.Image.Format),
                            Source: &types.ImageSourceMemberBytes{
                                Value: imageBytes,
                            },
                        },
                    })
                }
            }
        }
        
        messages = append(messages, types.Message{
            Role:    types.ConversationRole(schemaMsg.Role),
            Content: contentBlocks,
        })
    }
    
    // Build inference configuration
    inferenceConfig := &types.InferenceConfiguration{
        MaxTokens: aws.Int32(int32(workflowState.BedrockConfig.MaxTokens)),
    }
    
    // Add thinking configuration if enabled
    if workflowState.BedrockConfig.Thinking != nil {
        inferenceConfig.AdditionalModelRequestFields = map[string]interface{}{
            "thinking": map[string]interface{}{
                "type":          workflowState.BedrockConfig.Thinking.Type,
                "budget_tokens": workflowState.BedrockConfig.Thinking.BudgetTokens,
            },
        }
    }
    
    return &bedrockruntime.ConverseInput{
        ModelId:            aws.String("anthropic.claude-3-7-sonnet-20250219-v1:0"),
        Messages:           messages,
        InferenceConfig:    inferenceConfig,
        AdditionalModelRequestFields: map[string]interface{}{
            "anthropic_version": workflowState.BedrockConfig.AnthropicVersion,
        },
    }, nil
}

func extractTextContent(output types.ConverseOutput) string {
    if output.Message != nil && len(output.Message.Content) > 0 {
        for _, content := range output.Message.Content {
            if textContent, ok := content.(*types.ContentBlockMemberText); ok {
                return textContent.Value
            }
        }
    }
    return ""
}

func extractThinkingContent(output types.ConverseOutput) string {
    // Extract thinking content if available
    // This depends on the specific response format from Bedrock
    // Implementation would vary based on Bedrock's thinking output format
    return ""
}

func extractTokenUsage(usage *types.TokenUsage) *schema.TokenUsage {
    if usage == nil {
        return nil
    }
    
    tokenUsage := &schema.TokenUsage{
        InputTokens:  int(usage.InputTokens),
        OutputTokens: int(usage.OutputTokens),
        TotalTokens:  int(usage.TotalTokens),
    }
    
    // Add thinking tokens if available
    // This would depend on Bedrock's response format
    
    return tokenUsage
}
```

## ProcessTurn1Response Function Migration

### Before (Simple Text Processing)
```go
package main

import (
    "context"
    "encoding/json"
    "strings"
    
    "github.com/aws/aws-lambda-go/lambda"
    "workflow-function/shared/schema"
)

func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
    var workflowState schema.WorkflowState
    json.Unmarshal(event, &workflowState)
    
    // Simple text processing
    responseContent := workflowState.Turn1Response["content"].(string)
    
    // Basic parsing
    referenceAnalysis := map[string]interface{}{
        "status": "EXTRACTION_COMPLETE",
        "content": responseContent,
        "processedAt": schema.FormatISO8601(),
    }
    
    workflowState.ReferenceAnalysis = referenceAnalysis
    workflowState.VerificationContext.Status = schema.StatusTurn1Processed
    
    return workflowState, nil
}
```

### After (Structured Processing with Context)
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    
    "github.com/aws/aws-lambda-go/lambda"
    "workflow-function/shared/schema"
)

func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
    var workflowState schema.WorkflowState
    json.Unmarshal(event, &workflowState)
    
    // Extract TurnResponse from Turn1Response
    var turnResponse schema.TurnResponse
    if turnData, exists := workflowState.Turn1Response["turnResponse"]; exists {
        turnResponseBytes, _ := json.Marshal(turnData)
        json.Unmarshal(turnResponseBytes, &turnResponse)
    }
    
    // Process based on verification type
    var referenceAnalysis map[string]interface{}
    var err error
    
    switch workflowState.VerificationContext.VerificationType {
    case schema.VerificationTypeLayoutVsChecking:
        referenceAnalysis, err = processLayoutValidation(turnResponse)
    case schema.VerificationTypePreviousVsCurrent:
        if workflowState.HistoricalContext != nil && len(workflowState.HistoricalContext) > 0 {
            referenceAnalysis, err = processHistoricalEnhancement(turnResponse, workflowState.HistoricalContext)
        } else {
            referenceAnalysis, err = processFreshExtraction(turnResponse)
        }
    default:
        return nil, fmt.Errorf("unsupported verification type: %s", workflowState.VerificationContext.VerificationType)
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to process turn 1 response: %w", err)
    }
    
    // Add processing metadata
    referenceAnalysis["processedAt"] = schema.FormatISO8601()
    referenceAnalysis["turnId"] = turnResponse.TurnId
    referenceAnalysis["tokenUsage"] = turnResponse.TokenUsage
    
    workflowState.ReferenceAnalysis = referenceAnalysis
    workflowState.VerificationContext.Status = schema.StatusTurn1Processed
    
    return workflowState, nil
}

func processLayoutValidation(response schema.TurnResponse) (map[string]interface{}, error) {
    analysis := map[string]interface{}{
        "status": "VALIDATION_COMPLETE",
        "sourceType": "LAYOUT_REFERENCE",
        "validationResults": extractValidationResults(response.Response.Content),
        "basicObservations": extractBasicObservations(response.Response.Content),
        "contextForTurn2": map[string]interface{}{
            "referenceValidated": true,
            "useSystemPromptReference": true,
            "validationPassed": true,
            "readyForTurn2": true,
        },
    }
    
    return analysis, nil
}

func processHistoricalEnhancement(response schema.TurnResponse, historicalContext map[string]interface{}) (map[string]interface{}, error) {
    analysis := map[string]interface{}{
        "status": "EXTRACTION_COMPLETE",
        "sourceType": "HISTORICAL_WITH_VISUAL_CONFIRMATION",
        "historicalBaseline": historicalContext,
        "visualConfirmation": extractVisualConfirmation(response.Response.Content),
        "contextForTurn2": map[string]interface{}{
            "baselineEstablished": true,
            "useHistoricalAsReference": true,
            "visuallyConfirmed": true,
        },
    }
    
    return analysis, nil
}

func processFreshExtraction(response schema.TurnResponse) (map[string]interface{}, error) {
    analysis := map[string]interface{}{
        "status": "EXTRACTION_COMPLETE",
        "sourceType": "FRESH_VISUAL_ANALYSIS",
        "extractedStructure": extractMachineStructure(response.Response.Content),
        "extractedState": extractMachineState(response.Response.Content),
        "contextForTurn2": map[string]interface{}{
            "baselineSource": "EXTRACTED_STATE",
            "useHistoricalData": false,
            "extractedDataAvailable": true,
            "readyForTurn2": true,
        },
    }
    
    return analysis, nil
}

func extractValidationResults(content string) map[string]interface{} {
    // Extract validation information from the response
    // Look for confirmation keywords and structure validation
    return map[string]interface{}{
        "structureConfirmed": strings.Contains(strings.ToLower(content), "confirmed"),
        "layoutMatches": strings.Contains(strings.ToLower(content), "layout"),
        "productTypesConfirmed": extractProductTypeCount(content),
        "fullyStocked": strings.Contains(strings.ToLower(content), "fully"),
    }
}

func extractBasicObservations(content string) map[string]interface{} {
    // Extract basic observations for Turn 2 context
    return map[string]interface{}{
        "rowConfirmation": extractRowInformation(content),
        "columnConfirmation": extractColumnInformation(content),
        "productDistribution": extractProductDistribution(content),
        "overallState": extractOverallState(content),
    }
}

// Additional helper functions for content extraction
func extractProductTypeCount(content string) int {
    // Implementation to count product types mentioned
    return 0
}

func extractRowInformation(content string) string {
    // Implementation to extract row information
    return ""
}

func extractColumnInformation(content string) string {
    // Implementation to extract column information  
    return ""
}

func extractProductDistribution(content string) string {
    // Implementation to extract product distribution info
    return ""
}

func extractOverallState(content string) string {
    // Implementation to extract overall state description
    return ""
}

func extractVisualConfirmation(content string) map[string]interface{} {
    // Implementation to extract visual confirmation details
    return map[string]interface{}{}
}

func extractMachineStructure(content string) map[string]interface{} {
    // Implementation to extract machine structure from response
    return map[string]interface{}{}
}

func extractMachineState(content string) map[string]interface{} {
    // Implementation to extract machine state from response
    return map[string]interface{}{}
}
```

## Key Migration Points

### 1. Function Signature Consistency
All functions maintain the same input/output signature for Step Functions compatibility.

### 2. Backward Compatibility
- Keep existing fields alongside new Base64 fields
- Support both text and Bedrock message formats
- Handle old schema versions gracefully

### 3. Error Handling
- Validate Base64 data before Bedrock API calls
- Provide detailed error messages for debugging
- Handle API limits and constraints

### 4. Performance Optimization
- Generate Base64 once in FetchImages function
- Reuse Base64 data throughout workflow
- Monitor memory usage with large images

### 5. Logging and Traceability
- Always log S3 URLs for audit trail
- Track Base64 generation and validation
- Monitor API call success rates

This migration ensures that Lambda functions can take advantage of the new Base64 support while maintaining compatibility with existing implementations.