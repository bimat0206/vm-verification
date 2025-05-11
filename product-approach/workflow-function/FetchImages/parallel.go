package main

import (
    "context"
    "sync"
)

// ParallelFetchResults holds the results of parallel fetches.
type ParallelFetchResults struct {
    ReferenceMeta     ImageMetadata
    CheckingMeta      ImageMetadata
    LayoutMeta        *LayoutMetadata
    HistoricalContext *HistoricalContext
    Errors            []error
}

// ParallelFetch executes the S3 and DynamoDB fetches concurrently.
func ParallelFetch(
    ctx context.Context,
    referenceS3 S3URI,
    checkingS3 S3URI,
    layoutId int64,
    layoutPrefix string,
    prevVerificationId string,
) ParallelFetchResults {
    var wg sync.WaitGroup
    results := ParallelFetchResults{}
    var mu sync.Mutex

    wg.Add(2)

    // S3: Reference image metadata
    go func() {
        defer wg.Done()
        meta, err := GetS3ImageMetadata(ctx, referenceS3)
        mu.Lock()
        defer mu.Unlock()
        if err != nil {
            results.Errors = append(results.Errors, err)
            Error("Failed to fetch reference image metadata", map[string]interface{}{
                "bucket": referenceS3.Bucket,
                "key":    referenceS3.Key,
                "error":  err.Error(),
            })
        } else {
            results.ReferenceMeta = meta
            Info("Successfully fetched reference image metadata", map[string]interface{}{
                "contentType": meta.ContentType,
                "size":       meta.Size,
            })
        }
    }()

    // S3: Checking image metadata
    go func() {
        defer wg.Done()
        meta, err := GetS3ImageMetadata(ctx, checkingS3)
        mu.Lock()
        defer mu.Unlock()
        if err != nil {
            results.Errors = append(results.Errors, err)
            Error("Failed to fetch checking image metadata", map[string]interface{}{
                "bucket": checkingS3.Bucket,
                "key":    checkingS3.Key,
                "error":  err.Error(),
            })
        } else {
            results.CheckingMeta = meta
            Info("Successfully fetched checking image metadata", map[string]interface{}{
                "contentType": meta.ContentType,
                "size":       meta.Size,
            })
        }
    }()

    // DynamoDB: Layout metadata (optional)
    if layoutId != 0 && layoutPrefix != "" {
        wg.Add(1)
        go func() {
            defer wg.Done()
            meta, err := FetchLayoutMetadata(ctx, layoutId, layoutPrefix)
            mu.Lock()
            defer mu.Unlock()
            if err != nil {
                // Only add as error if verificationType requires it
                // (would need to pass verificationType to this function)
                results.Errors = append(results.Errors, err)
                Error("Failed to fetch layout metadata", map[string]interface{}{
                    "layoutId":     layoutId,
                    "layoutPrefix": layoutPrefix,
                    "error":        err.Error(),
                })
            } else {
                results.LayoutMeta = meta
                Info("Successfully fetched layout metadata", map[string]interface{}{
                    "layoutId":     layoutId,
                    "layoutPrefix": layoutPrefix,
                })
            }
        }()
    }

    // DynamoDB: Historical context (optional)
    if prevVerificationId != "" {
        wg.Add(1)
        go func() {
            defer wg.Done()
            ctxObj, err := FetchHistoricalContext(ctx, prevVerificationId)
            mu.Lock()
            defer mu.Unlock()
            if err != nil {
                // Only add as error if verificationType requires it
                results.Errors = append(results.Errors, err)
                Error("Failed to fetch historical verification", map[string]interface{}{
                    "previousVerificationId": prevVerificationId,
                    "error":                  err.Error(),
                })
            } else {
                results.HistoricalContext = ctxObj
                Info("Successfully fetched historical verification", map[string]interface{}{
                    "previousVerificationId": prevVerificationId,
                })
            }
        }()
    }

    wg.Wait()
    return results
}