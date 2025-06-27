# API Backend Updates for Previous Analysis Feature

## Overview
Updated both the List and Conversation API functions to include the `previousVerificationId` field, enabling the frontend Previous Analysis feature for "Previous vs Current" verification types.

## Changes Made

### 1. List API Function (`../api-function/api_verifications/list/main.go`)

**File**: `../api-function/api_verifications/list/main.go`

**Change**: Added `PreviousVerificationID` field to the `VerificationRecord` struct

```go
type VerificationRecord struct {
    VerificationID       string                 `json:"verificationId" dynamodbav:"verificationId"`
    VerificationAt       string                 `json:"verificationAt" dynamodbav:"verificationAt"`
    VerificationStatus   string                 `json:"verificationStatus" dynamodbav:"verificationStatus"`
    VerificationType     string                 `json:"verificationType" dynamodbav:"verificationType"`
    VendingMachineID     string                 `json:"vendingMachineId" dynamodbav:"vendingMachineId"`
    ReferenceImageURL    string                 `json:"referenceImageUrl" dynamodbav:"referenceImageUrl"`
    CheckingImageURL     string                 `json:"checkingImageUrl" dynamodbav:"checkingImageUrl"`
    LayoutID             *int                   `json:"layoutId,omitempty" dynamodbav:"layoutId,omitempty"`
    LayoutPrefix         *string                `json:"layoutPrefix,omitempty" dynamodbav:"layoutPrefix,omitempty"`
    OverallAccuracy      *float64               `json:"overallAccuracy,omitempty" dynamodbav:"overallAccuracy,omitempty"`
    CorrectPositions     *int                   `json:"correctPositions,omitempty" dynamodbav:"correctPositions,omitempty"`
    DiscrepantPositions  *int                   `json:"discrepantPositions,omitempty" dynamodbav:"discrepantPositions,omitempty"`
    Result               map[string]interface{} `json:"result,omitempty" dynamodbav:"result,omitempty"`
    VerificationSummary  map[string]interface{} `json:"verificationSummary,omitempty" dynamodbav:"verificationSummary,omitempty"`
    CreatedAt            string                 `json:"createdAt,omitempty" dynamodbav:"createdAt,omitempty"`
    UpdatedAt            string                 `json:"updatedAt,omitempty" dynamodbav:"updatedAt,omitempty"`
    // ✅ NEW FIELD ADDED
    PreviousVerificationID *string              `json:"previousVerificationId,omitempty" dynamodbav:"previousVerificationId,omitempty"`
}
```

### 2. Conversation API Function (`../api-function/api_verifications/conversation/main.go`)

**File**: `../api-function/api_verifications/conversation/main.go`

**Change**: Added `PreviousVerificationID` field to the `VerificationRecord` struct

```go
type VerificationRecord struct {
    VerificationID       string                 `json:"verificationId" dynamodbav:"verificationId"`
    VerificationAt       string                 `json:"verificationAt" dynamodbav:"verificationAt"`
    VerificationStatus   string                 `json:"verificationStatus" dynamodbav:"verificationStatus"`
    VerificationType     string                 `json:"verificationType" dynamodbav:"verificationType"`
    VendingMachineID     string                 `json:"vendingMachineId" dynamodbav:"vendingMachineId"`
    ReferenceImageURL    string                 `json:"referenceImageUrl" dynamodbav:"referenceImageUrl"`
    CheckingImageURL     string                 `json:"checkingImageUrl" dynamodbav:"checkingImageUrl"`
    LayoutID             *int                   `json:"layoutId,omitempty" dynamodbav:"layoutId,omitempty"`
    LayoutPrefix         *string                `json:"layoutPrefix,omitempty" dynamodbav:"layoutPrefix,omitempty"`
    OverallAccuracy      *float64               `json:"overallAccuracy,omitempty" dynamodbav:"overallAccuracy,omitempty"`
    CorrectPositions     *int                   `json:"correctPositions,omitempty" dynamodbav:"correctPositions,omitempty"`
    DiscrepantPositions  *int                   `json:"discrepantPositions,omitempty" dynamodbav:"discrepantPositions,omitempty"`
    Result               map[string]interface{} `json:"result,omitempty" dynamodbav:"result,omitempty"`
    VerificationSummary  map[string]interface{} `json:"verificationSummary,omitempty" dynamodbav:"verificationSummary,omitempty"`
    CreatedAt            string                 `json:"createdAt,omitempty" dynamodbav:"createdAt,omitempty"`
    UpdatedAt            string                 `json:"updatedAt,omitempty" dynamodbav:"updatedAt,omitempty"`
    // ✅ NEW FIELD ADDED
    PreviousVerificationID *string              `json:"previousVerificationId,omitempty" dynamodbav:"previousVerificationId,omitempty"`
    // Processed paths from verification metadata
    Turn1ProcessedPath   string                 `json:"turn1ProcessedPath,omitempty" dynamodbav:"turn1ProcessedPath,omitempty"`
    Turn2ProcessedPath   string                 `json:"turn2ProcessedPath,omitempty" dynamodbav:"turn2ProcessedPath,omitempty"`
}
```

## Field Details

- **Field Name**: `PreviousVerificationID`
- **JSON Key**: `previousVerificationId`
- **DynamoDB Attribute**: `previousVerificationId`
- **Type**: `*string` (pointer to string, allowing nil values)
- **Tags**: `omitempty` - field will be omitted from JSON/DynamoDB if nil or empty

## Impact

### ✅ What This Enables:
1. **Frontend Previous Analysis Feature**: The frontend can now retrieve the `previousVerificationId` from API responses
2. **Previous vs Current Verification Support**: Enables the comparison view between current and previous verification analyses
3. **Accordion UI Components**: Supports the expandable/collapsible Previous Analysis and Current Analysis sections

### 🔄 API Response Changes:
- List API (`/verifications`) will now include `previousVerificationId` in each verification record
- Conversation API (`/verifications/{id}/conversation`) will include `previousVerificationId` in the verification metadata

### 📋 Next Steps:
1. **Deploy Updated APIs**: Both API functions need to be redeployed to AWS Lambda
2. **Test Integration**: Verify that the frontend Previous Analysis feature works with real data
3. **Data Population**: Ensure that `previousVerificationId` values are properly set in the DynamoDB table for "Previous vs Current" verification types

## Backup Files Created
- `../api-function/api_verifications/list/main.go.backup`
- `../api-function/api_verifications/conversation/main.go.backup`

## Validation
Both Go files have been syntax-checked with `go fmt` and are ready for deployment.
