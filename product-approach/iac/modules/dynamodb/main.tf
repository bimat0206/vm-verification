# VerificationResults Table
resource "aws_dynamodb_table" "verification_results" {
  name         = var.verification_results_table_name
  billing_mode = var.billing_mode
  hash_key     = "verificationId"
  range_key    = "verificationAt"

  # Only set capacity if using PROVISIONED billing mode
  read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
  write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null

  # Primary Key attributes
  attribute {
    name = "verificationId"
    type = "S"
  }

  attribute {
    name = "verificationAt"
    type = "S"
  }

  # Attributes for GSIs
  attribute {
    name = "layoutId"
    type = "N"
  }

  attribute {
    name = "verificationType"
    type = "S"
  }

  attribute {
    name = "verificationStatus"
    type = "S"
  }

  attribute {
    name = "checkingImageUrl"
    type = "S"
  }

  attribute {
    name = "referenceImageUrl"
    type = "S"
  }

  # GSI1: Query by layoutId and verification time
  global_secondary_index {
    name            = "LayoutIndex"
    hash_key        = "layoutId"
    range_key       = "verificationAt"
    projection_type = "ALL"
    
    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
    write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
  }

  # GSI2: Query by verification type and time
  global_secondary_index {
    name            = "VerificationTypeIndex"
    hash_key        = "verificationType"
    range_key       = "verificationAt"
    projection_type = "ALL"
    
    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
    write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
  }

  # GSI3: Query by verification status and time
  global_secondary_index {
    name            = "VerificationStatusIndex"
    hash_key        = "verificationStatus"
    range_key       = "verificationAt"
    projection_type = "INCLUDE"
    non_key_attributes = ["vendingMachineId", "location", "verificationSummary"]
    
    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
    write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
  }

  # GSI4: Query by checking image URL
  global_secondary_index {
    name            = "CheckingImageIndex"
    hash_key        = "checkingImageUrl"
    range_key       = "verificationAt"
    projection_type = "ALL"
    
    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
    write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
  }

  # GSI5: Query by reference image URL
  global_secondary_index {
    name            = "ReferenceImageIndex"
    hash_key        = "referenceImageUrl"
    range_key       = "verificationAt"
    projection_type = "ALL"
    
    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
    write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
  }

  point_in_time_recovery {
    enabled = var.point_in_time_recovery
  }

  # TTL for automatic expiration (if needed)
  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  tags = merge(
    var.common_tags,
    {
      Name = var.verification_results_table_name
    }
  )
}

# LayoutMetadata Table
resource "aws_dynamodb_table" "layout_metadata" {
  name         = var.layout_metadata_table_name
  billing_mode = var.billing_mode
  hash_key     = "layoutId"
  range_key    = "layoutPrefix"

  # Only set capacity if using PROVISIONED billing mode
  read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
  write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null

  # Primary Key attributes
  attribute {
    name = "layoutId"
    type = "N"
  }

  attribute {
    name = "layoutPrefix"
    type = "S"
  }

  # Attributes for GSIs and LSIs
  attribute {
    name = "createdAt"
    type = "S"
  }

  attribute {
    name = "vendingMachineId"
    type = "S"
  }

  # LSI1: Sort by creation date for a specific layoutId
  local_secondary_index {
    name            = "CreatedAtIndex"
    range_key       = "createdAt"
    projection_type = "ALL"
  }

  # GSI1: Query by creation date
  global_secondary_index {
    name            = "CreatedAtGSI"
    hash_key        = "createdAt"
    range_key       = "layoutId"
    projection_type = "KEYS_ONLY"
    
    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
    write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
  }

  # GSI2: Query by vending machine ID
  global_secondary_index {
    name            = "VendingMachineIdIndex"
    hash_key        = "vendingMachineId"
    range_key       = "createdAt"
    projection_type = "ALL"
    
    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
    write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
  }

  point_in_time_recovery {
    enabled = var.point_in_time_recovery
  }

  tags = merge(
    var.common_tags,
    {
      Name = var.layout_metadata_table_name
    }
  )
}

# ConversationHistory Table
resource "aws_dynamodb_table" "conversation_history" {
  name         = var.conversation_history_table_name
  billing_mode = var.billing_mode
  hash_key     = "verificationId"
  range_key    = "conversationAt"

  # Only set capacity if using PROVISIONED billing mode
  read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
  write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null

  # Primary Key attributes
  attribute {
    name = "verificationId"
    type = "S"
  }

  attribute {
    name = "conversationAt"
    type = "S"
  }

  # Attributes for GSIs
  attribute {
    name = "turnStatus"
    type = "S"
  }

  attribute {
    name = "layoutId"
    type = "N"
  }

  # GSI1: Query by turn status
  global_secondary_index {
    name            = "TurnStatusIndex"
    hash_key        = "turnStatus"
    range_key       = "conversationAt"
    projection_type = "INCLUDE"
    non_key_attributes = ["verificationId", "vendingMachineId", "confidenceScore"]
    
    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
    write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
  }

  # GSI2: Query by layout ID
  global_secondary_index {
    name            = "LayoutIdIndex"
    hash_key        = "layoutId"
    range_key       = "conversationAt"
    projection_type = "KEYS_ONLY"
    
    # Only set capacity if using PROVISIONED billing mode
    read_capacity  = var.billing_mode == "PROVISIONED" ? var.read_capacity : null
    write_capacity = var.billing_mode == "PROVISIONED" ? var.write_capacity : null
  }

  point_in_time_recovery {
    enabled = var.point_in_time_recovery
  }

  # TTL for automatic expiration
  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  tags = merge(
    var.common_tags,
    {
      Name = var.conversation_history_table_name
    }
  )
}