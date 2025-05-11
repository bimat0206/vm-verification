# Lambda Function Design Guide for API Gateway Integration

## Core Principles

When designing Lambda functions to work with Amazon API Gateway, follow these core principles:

1. **Understand the integration type** - API Gateway offers both proxy and non-proxy integrations, each requiring different Lambda function structures
2. **Properly handle event payloads** - Design your function to correctly parse the specific data structure that API Gateway sends
3. **Implement consistent response formatting** - Return responses in the exact format API Gateway expects
4. **Handle errors gracefully** - Provide meaningful error responses that API Gateway can correctly transform

## Integration Types and Required Structure

### Proxy Integration (`AWS_PROXY`)

```go
// Handler for API Gateway Proxy Integration
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // API Gateway proxy integration passes the entire HTTP request as-is
    // and expects a specific response format
    
    // 1. Parse the request body into your domain model
    var myRequest MyRequestType
    if err := json.Unmarshal([]byte(request.Body), &myRequest); err != nil {
        return events.APIGatewayProxyResponse{
            StatusCode: 400,
            Body: `{"error": "Invalid request format"}`,
            Headers: map[string]string{"Content-Type": "application/json"},
        }, nil
    }
    
    // 2. Process with your business logic
    result, err := processRequest(myRequest)
    if err != nil {
        return events.APIGatewayProxyResponse{
            StatusCode: 500,
            Body: fmt.Sprintf(`{"error": "%s"}`, err.Error()),
            Headers: map[string]string{"Content-Type": "application/json"},
        }, nil
    }
    
    // 3. Return formatted response
    responseBody, _ := json.Marshal(result)
    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Body: string(responseBody),
        Headers: map[string]string{"Content-Type": "application/json"},
    }, nil
}
```

### Non-Proxy Integration (`AWS`)

```go
// Handler for API Gateway Non-Proxy Integration
func Handler(ctx context.Context, wrappedRequest WrappedRequest) (interface{}, error) {
    // API Gateway non-proxy integration wraps your request in a structure
    
    // 1. Define the wrapped request structure API Gateway sends
    type WrappedRequest struct {
        Body        json.RawMessage         `json:"body"`
        Headers     map[string]string       `json:"headers"`
        Method      string                  `json:"method"`
        Params      map[string]string       `json:"params"`
        Query       map[string]string       `json:"query"`
    }
    
    // 2. Extract your actual request from the body field
    var myRequest MyRequestType
    if err := json.Unmarshal(wrappedRequest.Body, &myRequest); err != nil {
        return ErrorResponse{
            Error: "Invalid request format",
        }, nil
    }
    
    // 3. Process with your business logic
    result, err := processRequest(myRequest)
    if err != nil {
        return ErrorResponse{
            Error: err.Error(),
        }, nil
    }
    
    // 4. Return your domain object directly
    // API Gateway will handle the transformation based on mapping templates
    return result, nil
}
```

## Best Practices

1. **Always inspect your API Gateway setup first**
   - Check the integration type (AWS_PROXY vs AWS) before designing your Lambda
   - Understand how data will be passed to your function

2. **Use language-specific event objects**
   - In Go, use `events.APIGatewayProxyRequest`/`events.APIGatewayProxyResponse` from aws-lambda-go
   - In Node.js, use the event and context parameters
   - In Python, use the appropriate boto3 models

3. **Design for API Gateway's error handling**
   - Return errors in a consistent format
   - Use appropriate HTTP status codes
   - Include CORS headers in error responses if needed

4. **Test with actual API Gateway events**
   - Create sample events that match what API Gateway sends
   - Test locally with these events before deployment

5. **Handle path parameters, query strings, and headers properly**
   - API Gateway passes these differently based on integration type
   - Ensure your function extracts them correctly

6. **Return properly formatted responses**
   - For proxy integration: Include statusCode, body, and headers
   - For non-proxy: Return domain objects that match your mapping templates

7. **Implement proper logging**
   - Log request details at appropriate levels
   - Include request IDs for traceability

## Common Pitfalls to Avoid

1. **Mismatched integration type**
   - The most common issue is designing for proxy integration when using non-proxy or vice versa

2. **Incorrect response format**
   - Missing required fields in proxy integration responses
   - Returning nested objects when API Gateway expects a flat structure

3. **Not handling CORS correctly**
   - Forgetting to include CORS headers in responses
   - Failing to handle OPTIONS requests properly

4. **Poor error handling**
   - Returning raw errors instead of formatted responses
   - Not maintaining consistent error formats

5. **Ignoring path parameters or query strings**
   - Not extracting these from the event object
   - Using the wrong fields based on integration type

By following these guidelines, your Lambda functions will properly integrate with API Gateway, providing a reliable and consistent API experience.