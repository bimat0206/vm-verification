# Vending Machine Verification API Guide

This guide explains how to call the verification API endpoints from your application.

## API Client Setup

The API client is already configured in `clients/api_client.py` with the correct base URL and authentication:

```python
def __init__(self, api_endpoint=None, api_key=None):
    self.api_endpoint = api_endpoint or os.environ.get('API_ENDPOINT')
    self.api_key = api_key or os.environ.get('API_KEY')
    
    if not self.api_endpoint or not self.api_key:
        raise ValueError("API_ENDPOINT and API_KEY must be provided")
```

## Common API Endpoints

### 1. Health Check

```python
# Import the API client
from clients.api_client import APIClient

# Initialize the client
api_client = APIClient()

# Make health check request
response = api_client.health_check()
# Returns: {"status": "healthy", "version": "1.x.x"}
```

**Example Response:**
```json
{
  "status": "healthy",
  "version": "1.2.3",
  "environment": "development",
  "timestamp": "2025-06-04T08:22:15Z"
}
```

### 2. Initiate Verification

```python
# Simplified verification API call with only required fields
response = api_client.initiate_verification(
    verificationType="LAYOUT_VS_CHECKING",  # or "PREVIOUS_VS_CURRENT"
    referenceImageUrl="s3://kootoro-dev-s3-reference-f6d3xl/path/to/reference.jpg",
    checkingImageUrl="s3://kootoro-dev-s3-checking-f6d3xl/path/to/checking.jpg",
    notificationEnabled=False  # Optional
)

# Response contains verification ID and status
verification_id = response.get('verificationId')
status = response.get('status')  # Usually "PROCESSING"
```

**Example Request:**
```json
POST /api/verifications
{
  "verificationContext": {
    "verificationType": "LAYOUT_VS_CHECKING",
    "referenceImageUrl": "s3://kootoro-dev-s3-reference-f6d3xl/path/to/reference.jpg",
    "checkingImageUrl": "s3://kootoro-dev-s3-checking-f6d3xl/path/to/checking.jpg",
    "notificationEnabled": false
  }
}
```

**Example Response:**
```json
{
  "verificationId": "verif-20250604082215-c689",
  "status": "PROCESSING",
  "createdAt": "2025-06-04T08:22:15Z",
  "estimatedCompletionTime": "2025-06-04T08:23:15Z"
}
```

### 3. Get Verification Details

```python
# Fetch details for a specific verification
verification_details = api_client.get_verification_details("a041e458-3171-43e9-a149-f63c5916d3a2")

# Access verification properties
status = verification_details.get('verificationStatus')
accuracy = verification_details.get('overallAccuracy')
```

**Example Request:**
```
GET /api/verifications/verif-20250604082215-c689
```

**Example Response:**
```json
{
  "verificationId": "verif-20250604082215-c689",
  "verificationStatus": "COMPLETED",
  "verificationType": "LAYOUT_VS_CHECKING",
  "referenceImageUrl": "s3://kootoro-dev-s3-reference-f6d3xl/path/to/reference.jpg",
  "checkingImageUrl": "s3://kootoro-dev-s3-checking-f6d3xl/path/to/checking.jpg",
  "overallAccuracy": 0.92,
  "verificationResults": {
    "layoutMatches": true,
    "productMatches": [
      {
        "productId": "A1",
        "referencePosition": {"row": 1, "column": 1},
        "checkingPosition": {"row": 1, "column": 1},
        "match": true
      },
      {
        "productId": "B2",
        "referencePosition": {"row": 2, "column": 3},
        "checkingPosition": {"row": 2, "column": 3},
        "match": true
      }
    ],
    "missingProducts": [],
    "extraProducts": []
  },
  "createdAt": "2025-06-04T08:22:15Z",
  "completedAt": "2025-06-04T08:23:05Z"
}
```

### 4. List Verifications

```python
# Get list of verifications with optional pagination
verifications = api_client.list_verifications(limit=20, offset=0)

# Access the results
results = verifications.get('results', [])
pagination = verifications.get('pagination', {})
```

**Example Request:**
```
GET /api/verifications?limit=20&offset=0
```

**Example Response:**
```json
{
  "results": [
    {
      "verificationId": "verif-20250604082215-c689",
      "verificationStatus": "COMPLETED",
      "verificationType": "LAYOUT_VS_CHECKING",
      "overallAccuracy": 0.92,
      "createdAt": "2025-06-04T08:22:15Z",
      "completedAt": "2025-06-04T08:23:05Z"
    },
    {
      "verificationId": "verif-20250604075512-a123",
      "verificationStatus": "COMPLETED",
      "verificationType": "PREVIOUS_VS_CURRENT",
      "overallAccuracy": 0.87,
      "createdAt": "2025-06-04T07:55:12Z",
      "completedAt": "2025-06-04T07:56:30Z"
    }
  ],
  "pagination": {
    "total": 42,
    "limit": 20,
    "offset": 0,
    "nextOffset": 20
  }
}
```

### 5. Lookup Verification

```python
# Find verifications by checking image URL and optional vending machine ID
lookup_results = api_client.lookup_verification(
    checkingImageUrl="s3://kootoro-dev-s3-checking-f6d3xl/AACZ_3.png",
    vendingMachineId="VM-001",  # Optional
    limit=10  # Optional
)
```

**Example Request:**
```
GET /api/verifications/lookup?checkingImageUrl=s3://kootoro-dev-s3-checking-f6d3xl/AACZ_3.png&vendingMachineId=VM-001&limit=10
```

**Example Response:**
```json
{
  "results": [
    {
      "verificationId": "verif-20250604082215-c689",
      "verificationStatus": "COMPLETED",
      "verificationType": "LAYOUT_VS_CHECKING",
      "createdAt": "2025-06-04T08:22:15Z"
    }
  ],
  "pagination": {
    "total": 1,
    "limit": 10,
    "offset": 0
  }
}
```

### 6. Get Verification Conversation

```python
# Retrieve conversation history for a verification
conversation = api_client.get_verification_conversation("a041e458-3171-43e9-a149-f63c5916d3a2")

# Access conversation data
history = conversation.get('history', [])
current_turn = conversation.get('currentTurn')
```

**Example Request:**
```
GET /api/verifications/verif-20250604082215-c689/conversation
```

**Example Response:**
```json
{
  "verificationId": "verif-20250604082215-c689",
  "conversationId": "conv-456",
  "currentTurn": 2,
  "history": [
    {
      "turn": 1,
      "role": "system",
      "content": "Analyzing reference image...",
      "timestamp": "2025-06-04T08:22:20Z",
      "metadata": {
        "tokenUsage": {
          "input": 4252,
          "output": 1837,
          "thinking": 876,
          "total": 6965
        },
        "latencyMs": 2840
      }
    },
    {
      "turn": 2,
      "role": "system",
      "content": "Comparing reference and checking images...",
      "timestamp": "2025-06-04T08:22:45Z",
      "metadata": {
        "tokenUsage": {
          "input": 5120,
          "output": 2856,
          "thinking": 1234,
          "total": 9210
        },
        "latencyMs": 3450
      }
    }
  ],
  "createdAt": "2025-06-04T08:22:15Z",
  "updatedAt": "2025-06-04T08:25:30Z"
}
```

### 7. Browse Images

```python
# Browse images in S3 buckets
images = api_client.browse_images(
    path="",  # Optional path prefix
    bucketType="reference"  # or "checking"
)

# Returns folder structure and files
```

**Example Request:**
```
GET /api/images/browse?bucketType=reference&path=machines/
```

**Example Response:**
```json
{
  "path": "machines/",
  "folders": [
    "VM-001/",
    "VM-002/",
    "VM-003/"
  ],
  "files": [
    {
      "key": "machines/index.json",
      "lastModified": "2025-05-30T12:34:56Z",
      "size": 2048
    }
  ]
}
```

### 8. Get Image URL

```python
# Generate a presigned URL for viewing an image
image_url_response = api_client.get_image_url(
    key="folder/image.jpg",
    bucketType="reference"  # or "checking"
)

# Access the URL
presigned_url = image_url_response.get('presignedUrl')
```

**Example Request:**
```
GET /api/images/url?key=machines/VM-001/front.jpg&bucketType=reference
```

**Example Response:**
```json
{
  "key": "machines/VM-001/front.jpg",
  "bucketType": "reference",
  "presignedUrl": "https://kootoro-dev-s3-reference-f6d3xl.s3.amazonaws.com/machines/VM-001/front.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=...",
  "expiresAt": "2025-06-04T09:22:15Z"
}
```

## Important Notes

1. All API endpoints now include the `/api/` prefix
2. Verification requests must wrap data in a `verificationContext` object (handled by the API client)
3. The API client handles authentication with the API key
4. S3 URLs should use the correct bucket names: `kootoro-dev-s3-reference-f6d3xl` and `kootoro-dev-s3-checking-f6d3xl`

## Common Error Codes

- **400**: Bad Request - Invalid parameters or S3 URLs
- **401/403**: Authentication/Authorization failure
- **404**: Resource not found
- **405**: Method Not Allowed - Wrong HTTP method or endpoint
- **500**: Internal Server Error - Often expected for non-existent verification IDs

## Error Response Example

```json
{
  "error": "RESOURCE_NOT_FOUND",
  "message": "Verification with ID 'verif-invalid' not found",
  "details": {
    "verificationId": "verif-invalid",
    "requestId": "req-abc123"
  }
}
```
