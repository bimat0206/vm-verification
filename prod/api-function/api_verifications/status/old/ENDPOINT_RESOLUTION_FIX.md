# DynamoDB Endpoint Resolution Fix

## Problem
The "not found, ResolveEndpointV2" error occurs due to AWS SDK Go v2 module version incompatibilities. This is a known issue documented in:
- [GitHub Issue #2370](https://github.com/aws/aws-sdk-go-v2/issues/2370)
- [GitHub Issue #2397](https://github.com/aws/aws-sdk-go-v2/issues/2397)

## Root Cause
The error happens when:
1. Service modules from before 11/15/23 are used with newer root modules
2. There's a version mismatch between different AWS SDK v2 modules
3. The endpoint resolver cannot properly resolve service endpoints in Lambda environment

## Solutions Implemented

### 1. Custom Endpoint Resolver
Added a custom endpoint resolver that explicitly sets service endpoints:
```go
customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
    switch service {
    case dynamodb.ServiceID:
        return aws.Endpoint{
            URL:           fmt.Sprintf("https://dynamodb.%s.amazonaws.com", region),
            SigningRegion: region,
            PartitionID:   "aws",
            HostnameImmutable: true,
        }, nil
    // ... other services
    }
})
```

### 2. Enhanced Error Handling
- Added retry logic with exponential backoff
- Included "ResolveEndpointV2" in retryable error patterns
- Added connectivity test during initialization

### 3. Improved Configuration
- Multiple region fallback options (AWS_REGION → REGION → AWS_DEFAULT_REGION)
- Custom HTTP client with proper timeouts
- AWS SDK logging integration for better debugging

## To Update SDK Modules
If the issue persists, update all AWS SDK modules to compatible versions:
```bash
cd /path/to/status/api
./update_sdk.sh
```

This script will:
1. Remove go.sum for clean dependency resolution
2. Update all AWS SDK modules to latest compatible versions
3. Run go mod tidy to clean up dependencies

## Testing
After deployment, test the endpoint:
```bash
curl -X GET https://your-api-gateway-url/api/verifications/status/test-verification-id
```

## Environment Variables
Ensure these are set in Lambda:
- `DYNAMODB_VERIFICATION_TABLE`
- `DYNAMODB_CONVERSATION_TABLE`
- `REGION` or `AWS_REGION`
- `LOG_LEVEL=DEBUG` (for troubleshooting)

## Additional Notes
- The `FORCE_ENDPOINT_RESOLVER` environment variable in Lambda should be removed if present
- Consider using AWS SDK v2 modules all from the same release date for consistency
- Monitor CloudWatch logs for any endpoint resolution warnings