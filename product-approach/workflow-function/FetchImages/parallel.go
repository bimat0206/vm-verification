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
        } else {
            results.ReferenceMeta = meta
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
        } else {
            results.CheckingMeta = meta
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
                results.Errors = append(results.Errors, err)
            } else {
                results.LayoutMeta = meta
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
                results.Errors = append(results.Errors, err)
            } else {
                results.HistoricalContext = ctxObj
            }
        }()
    }

    wg.Wait()
    return results
}
