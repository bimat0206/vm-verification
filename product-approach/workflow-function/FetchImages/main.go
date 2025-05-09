package main

import (
    "context"
    "encoding/json"
    //"fmt"
    //"os"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

func Handler(ctx context.Context, event events.LambdaFunctionURLRequest) (FetchImagesResponse, error) {
    // Parse input
    var req FetchImagesRequest
    if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
        return FetchImagesResponse{}, NewBadRequestError("Invalid JSON input", err)
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
