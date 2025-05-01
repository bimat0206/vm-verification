locals {
  # Generate a random suffix for resource names if not provided
  name_suffix = var.resource_name_suffix != "" ? var.resource_name_suffix : random_string.suffix[0].result

  # Standard naming convention for resources
  name_prefix = var.environment != "" ? "${var.project_name}-${var.environment}" : var.project_name

  # Resource name pattern function is defined in naming.tf
  
  # Common tags to be applied to all resources
  common_tags = merge(
    var.additional_tags,
    {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  )

  # S3 bucket names
  s3_buckets = {
    reference = local.generate_name("s3", "reference"),
    checking  = local.generate_name("s3", "checking"),
    results   = local.generate_name("s3", "results")
  }

  # DynamoDB table names
  dynamodb_tables = {
    verification_results = local.generate_name("dynamodb", "verification-results"),
    layout_metadata      = local.generate_name("dynamodb", "layout-metadata"),
    conversation_history = local.generate_name("dynamodb", "conversation-history")
  }

  # ECR Repository URL (without specific repository)
  ecr_repository_base_url = var.ecr.create_repositories ? "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com" : ""
}

# Generate a random suffix for resource names if not provided
resource "random_string" "suffix" {
  count   = var.resource_name_suffix == "" ? 1 : 0
  length  = 6
  special = false
  upper   = false
}

# Get current AWS account ID
data "aws_caller_identity" "current" {}
