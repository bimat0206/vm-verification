# API Images Browser - Issue Resolution

## Problem Summary
The Streamlit app was failing to browse the reference bucket with the following error:
```
Failed to browse reference bucket: Failed to get api/images/browser/: 500 Server Error: Internal Server Error for url: https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1/api/images/browser/?bucketType=reference
```

## Root Cause Analysis

### 1. Initial Investigation
- **Lambda Function**: ✅ Properly configured with correct environment variables
- **S3 Buckets**: ✅ Exist and are accessible
- **API Gateway**: ❌ Returning 405 Method Not Allowed instead of 500

### 2. Detailed Analysis
The diagnostic revealed:
- API Gateway was returning **405 Method Not Allowed** (not 500 as reported by Streamlit)
- Lambda function environment variables were correctly set:
  - `REFERENCE_BUCKET`: `kootoro-dev-s3-reference-f6d3xl`
  - `CHECKING_BUCKET`: `kootoro-dev-s3-checking-f6d3xl`
  - `LOG_LEVEL`: `INFO`

### 3. Terraform Configuration Issue
The Terraform plan revealed that Lambda functions were pointing to incorrect ECR image URIs:
- **Current**: `879654127886.dkr.ecr.us-east-1.amazonaws.com/vending-render:latest`
- **Expected**: `879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-api-images-browser-f6d3xl:latest`

## Solutions Implemented

### 1. Immediate Fix - API Gateway Deployment
**Script**: `fix_api_gateway.sh`
- Created new API Gateway deployment to ensure all configurations are active
- Verified Lambda permissions and resource configurations
- **Status**: ✅ Completed

### 2. Lambda Function Deployment
**Script**: `deploy.sh`
- Build and deploy the correct Docker image to ECR
- Update Lambda function with the correct image URI
- **Action Required**: Run deployment script

### 3. Terraform Configuration Fix
**Issue**: ECR image URIs in Terraform state are incorrect
- **Action Required**: Apply Terraform changes to fix image URIs

## Step-by-Step Resolution

### Step 1: Deploy Lambda Function (REQUIRED)
```bash
cd product-approach/api-function/api_images/browser
./deploy.sh
```

### Step 2: Apply Terraform Changes (RECOMMENDED)
```bash
cd product-approach/iac
terraform apply -target=module.lambda_functions
```

### Step 3: Verify Fix
```bash
cd product-approach/api-function/api_images/browser
./debug_lambda.sh test
```

## Expected Results

After completing the resolution steps:
1. **API Gateway**: Should return 200 OK for browser requests
2. **Lambda Function**: Should have correct ECR image URI
3. **Streamlit App**: Should successfully browse reference bucket

## Verification Commands

### Test API Endpoint Directly
```bash
curl "https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1/api/images/browser?bucketType=reference"
```

### Check Lambda Function Status
```bash
aws lambda get-function-configuration --function-name kootoro-dev-lambda-api-images-browser-f6d3xl --query 'ImageUri'
```

### Test from Streamlit App
1. Navigate to Image Browser page
2. Select "reference" bucket type
3. Click "Browse" button
4. Should display bucket contents without errors

## Prevention

To prevent similar issues in the future:
1. **Automated Testing**: Add integration tests for API endpoints
2. **Deployment Validation**: Verify ECR image URIs after Terraform apply
3. **Monitoring**: Set up CloudWatch alarms for API Gateway 4xx/5xx errors
4. **Documentation**: Keep deployment procedures up to date

## Files Created/Modified

### New Files
- `debug_lambda.sh` - Comprehensive diagnostic script
- `fix_api_gateway.sh` - API Gateway deployment fix
- `ISSUE_RESOLUTION.md` - This documentation

### Modified Files
- None (all fixes are deployment-related)

## Contact Information
For questions about this resolution, contact the development team or refer to the project documentation.
