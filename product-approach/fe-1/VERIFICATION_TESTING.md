# Verification Testing Guide

This guide helps you test the Initiate Verification functionality with the corrected API structure.

## ✅ Issue Resolution

The **405 Method Not Allowed** error has been **FIXED**! 

### What was wrong:
- API client was sending data directly instead of wrapping it in `verificationContext`
- Endpoint path was missing the `api/` prefix

### What was fixed:
- ✅ Request structure now matches API specification: `{"verificationContext": {...}}`
- ✅ Endpoint corrected to: `POST /api/verifications`
- ✅ All verification parameters properly wrapped

## 🧪 Testing the Fix

### Option 1: Using the Test Script
```bash
python3 test-verification.py
```

**Expected Results:**
- ✅ Request structure validation passes
- ⚠️ API call may return 400 Bad Request (due to test S3 URLs not existing)
- ❌ No more 405 Method Not Allowed errors

### Option 2: Using Streamlit App

1. **Start the app:**
   ```bash
   python3 -m streamlit run app.py
   ```

2. **Navigate to "Initiate Verification" page**

3. **Fill in the form with valid data:**

   **For Layout vs Checking verification:**
   ```
   Verification Type: Layout vs Checking
   Vending Machine ID: VM-001
   Reference Image URL: s3://kootoro-dev-s3-reference-f6d3xl/path/to/reference.jpg
   Checking Image URL: s3://kootoro-dev-s3-checking-f6d3xl/path/to/checking.jpg
   Layout ID: 12345
   Layout Prefix: test_prefix
   ```

   **For Previous vs Current verification:**
   ```
   Verification Type: Previous vs Current
   Vending Machine ID: VM-001
   Reference Image URL: s3://kootoro-dev-s3-checking-f6d3xl/path/to/previous.jpg
   Checking Image URL: s3://kootoro-dev-s3-checking-f6d3xl/path/to/current.jpg
   Previous Verification ID: (optional)
   ```

## 📋 Valid S3 URLs

To test with real data, use S3 URLs from your configured buckets:

### Reference Bucket: `kootoro-dev-s3-reference-f6d3xl`
```
s3://kootoro-dev-s3-reference-f6d3xl/[path-to-your-reference-image]
```

### Checking Bucket: `kootoro-dev-s3-checking-f6d3xl`
```
s3://kootoro-dev-s3-checking-f6d3xl/[path-to-your-checking-image]
```

### Finding Available Images

You can browse available images using the "Image Browser" page in the Streamlit app, or use AWS CLI:

```bash
# List reference images
aws s3 ls s3://kootoro-dev-s3-reference-f6d3xl/ --recursive

# List checking images  
aws s3 ls s3://kootoro-dev-s3-checking-f6d3xl/ --recursive
```

## 🎯 Expected Behavior

### ✅ Success Case
- **Status**: 200 OK
- **Response**: 
  ```json
  {
    "verificationId": "verif-20241220123456789",
    "verificationAt": "2024-12-20T12:34:56Z",
    "status": "PROCESSING",
    "message": "Verification has been successfully initiated."
  }
  ```

### ⚠️ Common Error Cases

**400 Bad Request - Invalid S3 URLs:**
- S3 objects don't exist
- Incorrect bucket names
- Invalid S3 URL format

**400 Bad Request - Missing Required Fields:**
- Layout ID/Prefix missing for Layout vs Checking
- Invalid verification type

**401/403 Authentication/Authorization:**
- Invalid API key
- Insufficient permissions

## 🔍 Debugging

### Check Request Structure
The API client now sends:
```json
{
  "verificationContext": {
    "verificationType": "LAYOUT_VS_CHECKING",
    "referenceImageUrl": "s3://bucket/path",
    "checkingImageUrl": "s3://bucket/path", 
    "vendingMachineId": "VM-001",
    "layoutId": 12345,
    "layoutPrefix": "test",
    "notificationEnabled": false
  }
}
```

### Monitor Logs
Check the Streamlit app logs for detailed error messages:
```bash
# In the terminal where Streamlit is running
# Look for API request/response logs
```

### Verify Configuration
```bash
python3 test-config.py
```

## 🎉 Success Indicators

1. **No 405 Method Not Allowed errors**
2. **Request reaches the API successfully**
3. **Proper error messages for invalid data (400 instead of 405)**
4. **Successful verification initiation with valid S3 URLs**

The verification API is now properly configured and should work as expected!
