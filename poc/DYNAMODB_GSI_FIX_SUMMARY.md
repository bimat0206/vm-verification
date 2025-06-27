# DynamoDB GSI Fix Summary

## Problem Identified
The Initialize function was failing with error:
```
LAYOUT_LOOKUP_FAILED: cannot retrieve layout metadata from referenceImageUrl: GSI query failed: operation error DynamoDB: Query, https response error StatusCode: 400, RequestID: ..., api error ValidationException: The table does not have the specified index: ReferenceImageIndex-gsi
```

## Root Cause Analysis
1. **Code Expectation**: The `layout_repo.go` code expects a GSI named `ReferenceImageIndex-gsi` on the `layout_metadata` table to query layouts by `referenceImageUrl`.

2. **Current State**: The `layout_metadata` table only had these GSIs:
   - `VendingMachineIdIndex`
   - `CreatedAtGSI`

3. **Missing GSI**: The `ReferenceImageIndex-gsi` GSI was missing from the table definition.

## Solution Implemented

### 1. Terraform Configuration Update
Updated `iac/modules/dynamodb/main.tf` to add the missing GSI to the `layout_metadata` table:

```hcl
# Added attribute definition
attribute {
  name = "referenceImageUrl"
  type = "S"
}

# Added GSI3: Query by reference image URL
global_secondary_index {
  name            = "ReferenceImageIndex-gsi"
  hash_key        = "referenceImageUrl"
  range_key       = "createdAt"
  projection_type = "ALL"
  
  # Only set capacity if using PROVISIONED billing mode
  read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
  write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
}
```

### 2. Next Steps Required

#### Apply Terraform Changes
```bash
cd iac
terraform plan -target='module.dynamodb_tables[0].aws_dynamodb_table.layout_metadata'
terraform apply -target='module.dynamodb_tables[0].aws_dynamodb_table.layout_metadata'
```

#### Verify GSI Creation
After applying, verify the GSI exists:
```bash
aws dynamodb describe-table \
  --table-name kootoro-dev-dynamodb-layout-metadata-f6d3xl \
  --query "Table.GlobalSecondaryIndexes[].IndexName"
```

Expected output should include: `["VendingMachineIdIndex", "CreatedAtGSI", "ReferenceImageIndex-gsi"]`

#### Test the Fix
1. Wait for GSI to become ACTIVE (may take a few minutes)
2. Re-run the Initialize function with the same test data
3. Verify that layout lookup now succeeds

## Technical Details

### GSI Configuration
- **Name**: `ReferenceImageIndex-gsi`
- **Hash Key**: `referenceImageUrl` (String)
- **Range Key**: `createdAt` (String)
- **Projection**: ALL (includes all attributes)
- **Billing**: Follows table billing mode (ON_DEMAND or PROVISIONED)

### Code Flow
1. Initialize function receives request with `referenceImageUrl`
2. When `layoutId`/`layoutPrefix` are missing, calls `GetLayoutByReferenceImage()`
3. Function queries `layout_metadata` table using `ReferenceImageIndex-gsi`
4. Returns layout metadata with `layoutId` and `layoutPrefix`
5. These values are populated in verification context for downstream processes

## Expected Outcome
After applying this fix:
- Initialize function should successfully lookup layout metadata by `referenceImageUrl`
- `layoutId` and `layoutPrefix` will be properly populated in the state output
- Downstream processes will receive complete layout information
- No more `RESOURCE_VALIDATION_FAILED` errors due to missing layout metadata

## ✅ IMPLEMENTATION COMPLETED

### Results
- **Terraform Apply**: Successfully completed after 4m3s
- **GSI Creation**: `ReferenceImageIndex-gsi` created and ACTIVE
- **Verification**: GSI confirmed present in table with ACTIVE status

### Current GSI Status
```bash
$ aws dynamodb describe-table --table-name kootoro-dev-dynamodb-layout-metadata-f6d3xl --query "Table.GlobalSecondaryIndexes[].IndexName"
[
    "ReferenceImageIndex-gsi",
    "VendingMachineIdIndex", 
    "CreatedAtGSI"
]
```

### Next Steps
1. **Test the Fix**: Re-run the Initialize function with the same test data that previously failed
2. **Verify Success**: Confirm that layout lookup now succeeds and `layoutId`/`layoutPrefix` are populated
3. **Monitor**: Watch for successful downstream processing without `RESOURCE_VALIDATION_FAILED` errors

The DynamoDB infrastructure is now ready to support the layout lookup functionality.
