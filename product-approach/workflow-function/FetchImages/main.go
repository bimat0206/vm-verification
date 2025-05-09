package main

import (
    "context"
    "encoding/json"
    "fmt"
    //"os"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

func Handler(ctx context.Context, event interface{}) (FetchImagesResponse, error) {
    // Parse input based on event type
    var req FetchImagesRequest
    var err error

    // Handle different invocation types
    switch e := event.(type) {
    case events.LambdaFunctionURLRequest:
        // Function URL invocation
        if err = json.Unmarshal([]byte(e.Body), &req); err != nil {
            Error("Failed to parse Function URL input", err)
            return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", err)
        }
    case map[string]interface{}:
        // Direct invocation from Step Function
        data, _ := json.Marshal(e)
        if err = json.Unmarshal(data, &req); err != nil {
            Error("Failed to parse Step Function input", map[string]interface{}{
                "error": err.Error(),
                "input": fmt.Sprintf("%+v", e),
            })
            return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", err)
        }
    case FetchImagesRequest:
        // Direct struct invocation
        req = e
    default:
        // Try raw JSON unmarshal as fallback
        data, _ := json.Marshal(event)
        if err = json.Unmarshal(data, &req); err != nil {
            Error("Failed to parse unknown input type", map[string]interface{}{
                "error": err.Error(),
                "input": fmt.Sprintf("%+v", event),
            })
            return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", err)
        }
    }

    // Validate input
    if err := req.Validate(); err != nil {
        return FetchImagesResponse{}, NewBadRequestError("Input validation failed", err)
    }

    // Parse S3 URIs
    referenceS3, err := ParseS3URI(req.ReferenceImageUrl)
    if err != nil {
        return FetchImagesResponse{}, NewBadRequestError("Invalid referenceImageUrl", err)
    }
    checkingS3, err := ParseS3URI(req.CheckingImageUrl)
    if err != nil {
        return FetchImagesResponse{}, NewBadRequestError("Invalid checkingImageUrl", err)
    }

    // Fetch S3 metadata in parallel (simplified, not using goroutines here for brevity)
    referenceMeta, err := GetS3ImageMetadata(ctx, referenceS3)
    if err != nil {
        return FetchImagesResponse{}, NewNotFoundError("Reference image not found or inaccessible", err)
    }
    checkingMeta, err := GetS3ImageMetadata(ctx, checkingS3)
    if err != nil {
        return FetchImagesResponse{}, NewNotFoundError("Checking image not found or inaccessible", err)
    }

    // TODO: Fetch DynamoDB metadata as needed (layout/historical) in next steps

    // Construct response
    resp := FetchImagesResponse{
        VerificationId:      req.VerificationId,
        ReferenceImageUrl:   req.ReferenceImageUrl,
        ReferenceImageMeta:  referenceMeta,
        CheckingImageUrl:    req.CheckingImageUrl,
        CheckingImageMeta:   checkingMeta,
        // LayoutMetadata, HistoricalContext to be filled in later
    }
    return resp, nil
}

func main() {
    lambda.Start(Handler)
}
