# API Quick Reference Card
## Vending Machine Verification System

## 🚀 Base Setup

```javascript
// .env.local (Next.js) or .env (React)
NEXT_PUBLIC_API_BASE_URL=https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1/api
NEXT_PUBLIC_API_KEY=WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH

// For Create React App
REACT_APP_API_BASE_URL=https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1/api
REACT_APP_API_KEY=WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH

// Basic API call
import apiClient from './services/apiClient';
const response = await apiClient.get('/health');
```

## 📁 File Upload API

```javascript
// Upload file
POST /api/images/upload?bucketType=reference&fileName=image.jpg&path=uploads
Body: { "fileContent": "base64-encoded-content" }

// React example
const result = await uploadService.uploadFile(file, 'reference', 'uploads/2024');
```

## 📋 Verification Management

```javascript
// List verifications
GET /api/verifications?verificationStatus=CORRECT&limit=20&offset=0

// Get specific verification
GET /api/verifications/{verificationId}

// Get conversation
GET /api/verifications/{verificationId}/conversation

// Get verification status
GET /api/verifications/status/{verificationId}

// Create verification
POST /api/verifications
Body: { vendingMachineId: "VM001", referenceImageUrl: "...", checkingImageUrl: "..." }

// React examples
const verifications = await verificationService.listVerifications({
  verificationStatus: 'CORRECT',
  limit: 20,
  offset: 0
});

const conversation = await verificationService.getConversation('verification-id');
const status = await verificationService.getStatus('verification-id');
```

## 🖼️ Image Management

```javascript
// Browse images
GET /api/images/browser?bucketType=reference
GET /api/images/browser/{path}?bucketType=checking

// Get image view URL
GET /api/images/{imageKey}/view?bucketType=reference

// React examples
const items = await imageService.browseImages('reference', 'folder/path');
const viewUrl = await imageService.getImageViewUrl('image-key', 'reference');
```

## 🏥 Health Check

```javascript
// Check system health
GET /api/health

// React example
const health = await healthService.checkHealth();
console.log(health.status); // 'healthy', 'degraded', or 'unhealthy'
```

## 🔧 Common Patterns

### Error Handling
```javascript
try {
  const result = await apiClient.get('/endpoint');
} catch (error) {
  console.error('API Error:', error.message);
  // Handle specific error codes
  if (error.status === 404) {
    // Handle not found
  }
}
```

### Loading States
```javascript
const [loading, setLoading] = useState(false);
const [error, setError] = useState(null);

const fetchData = async () => {
  setLoading(true);
  setError(null);
  try {
    const data = await apiCall();
    // Handle success
  } catch (err) {
    setError(err.message);
  } finally {
    setLoading(false);
  }
};
```

### File Upload with Progress
```javascript
const handleUpload = async (file) => {
  try {
    // Validate file
    if (file.size > 10 * 1024 * 1024) {
      throw new Error('File too large (max 10MB)');
    }
    
    // Upload
    const result = await uploadService.uploadFile(file, 'reference');
    console.log('Uploaded:', result.s3Key);
  } catch (error) {
    console.error('Upload failed:', error.message);
  }
};
```

### Pagination
```javascript
const [pagination, setPagination] = useState({ offset: 0, limit: 20 });

const loadMore = () => {
  if (pagination.nextOffset) {
    setPagination(prev => ({ ...prev, offset: pagination.nextOffset }));
  }
};
```

## 📊 Request & Response Examples

### 🏥 Health Check
**Request:**
```http
GET /api/health
X-Api-Key: WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH
```

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "dynamodb": {
      "status": "healthy",
      "message": "Tables accessible",
      "details": {
        "verification_table": "kootoro-dev-verification-table",
        "conversation_table": "kootoro-dev-conversation-table"
      }
    },
    "s3": {
      "status": "healthy",
      "message": "Buckets accessible",
      "details": {
        "reference_bucket": "kootoro-dev-reference-bucket",
        "checking_bucket": "kootoro-dev-checking-bucket",
        "results_bucket": "kootoro-dev-results-bucket"
      }
    },
    "bedrock": {
      "status": "healthy",
      "message": "Bedrock model available"
    }
  }
}
```

### 📋 List Verifications
**Request:**
```http
GET /api/verifications?verificationStatus=CORRECT&limit=20&offset=0
X-Api-Key: WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH
```

**Response:**
```json
{
  "results": [
    {
      "verificationId": "550e8400-e29b-41d4-a716-446655440000",
      "verificationStatus": "CORRECT",
      "vendingMachineId": "VM001",
      "verificationAt": "2024-01-01T12:00:00Z",
      "overallAccuracy": 0.95,
      "referenceImageUrl": "s3://reference-bucket/images/ref_001.jpg",
      "checkingImageUrl": "s3://checking-bucket/images/check_001.jpg",
      "turn1ProcessedPath": "s3://results-bucket/processed/turn1_001.md",
      "turn2ProcessedPath": "s3://results-bucket/processed/turn2_001.md"
    }
  ],
  "pagination": {
    "total": 100,
    "limit": 20,
    "offset": 0,
    "nextOffset": 20
  }
}
```

### 📝 Get Single Verification
**Request:**
```http
GET /api/verifications/550e8400-e29b-41d4-a716-446655440000
X-Api-Key: WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH
```

**Response:**
```json
{
  "verificationId": "550e8400-e29b-41d4-a716-446655440000",
  "verificationStatus": "CORRECT",
  "vendingMachineId": "VM001",
  "verificationAt": "2024-01-01T12:00:00Z",
  "overallAccuracy": 0.95,
  "referenceImageUrl": "s3://reference-bucket/images/ref_001.jpg",
  "checkingImageUrl": "s3://checking-bucket/images/check_001.jpg",
  "turn1ProcessedPath": "s3://results-bucket/processed/turn1_001.md",
  "turn2ProcessedPath": "s3://results-bucket/processed/turn2_001.md",
  "createdAt": "2024-01-01T11:30:00Z",
  "updatedAt": "2024-01-01T12:00:00Z"
}
```

### 💬 Get Conversation
**Request:**
```http
GET /api/verifications/550e8400-e29b-41d4-a716-446655440000/conversation
X-Api-Key: WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH
```

**Response:**
```json
{
  "turn1": "# Turn 1 Analysis\n\n## Reference Image Analysis\nThe reference image shows a vending machine with the following items:\n- Slot A1: Coca Cola (12 units)\n- Slot A2: Pepsi (8 units)\n- Slot B1: Water (15 units)\n\n## Checking Image Analysis\nThe checking image shows:\n- Slot A1: Coca Cola (11 units) - 1 unit sold\n- Slot A2: Pepsi (8 units) - No change\n- Slot B1: Water (15 units) - No change\n\n## Conclusion\nOne Coca Cola unit was sold from slot A1.",
  "turn2": "# Turn 2 Verification\n\n## Cross-validation\nComparing the analysis with transaction logs:\n- Transaction ID: TXN_001\n- Item: Coca Cola\n- Slot: A1\n- Timestamp: 2024-01-01T11:45:00Z\n\n## Final Verification\nThe analysis is CORRECT. One Coca Cola unit was indeed sold from slot A1 as indicated by both image analysis and transaction logs."
}
```

### ➕ Create Verification
**Request:**
```http
POST /api/verifications
Content-Type: application/json
X-Api-Key: WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH

{
  "vendingMachineId": "VM001",
  "referenceImageUrl": "s3://reference-bucket/images/ref_002.jpg",
  "checkingImageUrl": "s3://checking-bucket/images/check_002.jpg"
}
```

**Response:**
```json
{
  "verificationId": "660e8400-e29b-41d4-a716-446655440001",
  "verificationStatus": "PENDING",
  "vendingMachineId": "VM001",
  "verificationAt": "2024-01-01T13:00:00Z",
  "referenceImageUrl": "s3://reference-bucket/images/ref_002.jpg",
  "checkingImageUrl": "s3://checking-bucket/images/check_002.jpg",
  "createdAt": "2024-01-01T13:00:00Z"
}
```

### � Get Verification Status
**Request:**
```http
GET /api/verifications/status/660e8400-e29b-41d4-a716-446655440001
X-Api-Key: WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH
```

**Response (COMPLETED):**
```json
{
  "verificationId": "660e8400-e29b-41d4-a716-446655440001",
  "status": "COMPLETED",
  "currentStatus": "COMPLETED",
  "verificationStatus": "CORRECT",
  "s3References": {
    "turn1Processed": "s3://bucket/path/turn1-processed-response.md",
    "turn2Processed": "s3://bucket/path/turn2-processed-response.md"
  },
  "summary": {
    "message": "Verification completed successfully - No discrepancies found",
    "verificationAt": "2025-06-05T08:52:05Z",
    "verificationStatus": "CORRECT",
    "overallAccuracy": 0.833,
    "correctPositions": 35,
    "discrepantPositions": 7
  },
  "llmResponse": "# Verification Analysis\n\n## Summary\nThe verification process has been completed...",
  "verificationSummary": {
    "overall_confidence": "100%",
    "total_positions_checked": 42,
    "verification_outcome": "CORRECT"
  }
}
```

**Response (RUNNING):**
```json
{
  "verificationId": "660e8400-e29b-41d4-a716-446655440001",
  "status": "RUNNING",
  "currentStatus": "TURN1_COMPLETED",
  "verificationStatus": "PENDING",
  "s3References": {
    "turn1Processed": "s3://bucket/path/turn1-processed-response.md",
    "turn2Processed": ""
  },
  "summary": {
    "message": "Verification is currently in progress",
    "verificationAt": "2025-06-05T08:52:05Z",
    "verificationStatus": "PENDING"
  }
}
```

### �🔍 Lookup Verification
**Request:**
```http
GET /api/verifications/lookup?vendingMachineId=VM001&startDate=2024-01-01&endDate=2024-01-02
X-Api-Key: WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH
```

**Response:**
```json
[
  {
    "verificationId": "550e8400-e29b-41d4-a716-446655440000",
    "verificationStatus": "CORRECT",
    "vendingMachineId": "VM001",
    "verificationAt": "2024-01-01T12:00:00Z",
    "overallAccuracy": 0.95
  },
  {
    "verificationId": "660e8400-e29b-41d4-a716-446655440001",
    "verificationStatus": "PENDING",
    "vendingMachineId": "VM001",
    "verificationAt": "2024-01-01T13:00:00Z"
  }
]
```

## 🎨 CSS Classes

```css
/* Component containers */
.file-upload, .verification-list, .image-browser, .health-check

/* States */
.loading, .error, .success-message, .error-message

/* Interactive elements */
.upload-btn, .refresh-button, .pagination-controls

/* Layout */
.verification-grid, .items-grid, .filters, .browser-controls
```

## 🧪 Testing Snippets

```javascript
// Mock API client
jest.mock('./services/apiClient');

// Test file validation
expect(() => uploadService.validateFile(largeFile))
  .toThrow('File size must be less than 10MB');

// Test component rendering
render(<FileUpload />);
expect(screen.getByText('File Upload')).toBeInTheDocument();
```

## 🚨 Common Issues & Solutions

### CORS Errors
- Ensure API Gateway has proper CORS configuration
- Check that all required headers are included

### File Upload Failures
- Verify file size < 10MB
- Check file type is in allowed list
- Ensure base64 encoding is correct

### Authentication Issues
- Verify API key is set in environment variables
- Check X-Api-Key header is being sent

### Network Timeouts
- Increase timeout for large file uploads
- Implement retry logic for failed requests

## 📱 Mobile Considerations

```css
@media (max-width: 768px) {
  .verification-grid { grid-template-columns: 1fr; }
  .filters { flex-direction: column; }
  .browser-controls { flex-direction: column; }
}
```

## 🔒 Security Best Practices

- Store API keys in environment variables
- Validate file types and sizes on frontend
- Implement proper error handling
- Use HTTPS for all API calls
- Sanitize user inputs

---

## � API Gateway Configuration Summary

- **API Gateway ID**: `hpux2uegnd`
- **Base URL**: `https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1`
- **Stage**: `v1`
- **Region**: `us-east-1`
- **API Key**: `WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH`
- **Rate Limits**: 400 burst, 200/sec sustained
- **CORS**: Enabled for all endpoints

�📖 **Full Documentation**: See `REACT_API_INTEGRATION_GUIDE.md` for complete TypeScript examples and design patterns.
