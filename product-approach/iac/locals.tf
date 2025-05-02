locals {
  # Always use a random suffix for resource names, combined with any user-provided suffix
  name_suffix = var.resource_name_suffix != "" ? "${random_string.suffix.result}-${var.resource_name_suffix}" : random_string.suffix.result

  # Standard naming convention for resources
  name_prefix = var.environment != "" ? "${var.project_name}-${var.environment}" : var.project_name

  # Common tags to be applied to all resources
  common_tags = merge(
    var.additional_tags,
    {
      Project     = var.project_name
      Environment = var.environment
    }
  )

  # S3 bucket names
  s3_buckets = {
    reference = lower(join("-", compact([local.name_prefix, "s3", "reference", local.name_suffix]))),
    checking  = lower(join("-", compact([local.name_prefix, "s3", "checking", local.name_suffix]))),
    results   = lower(join("-", compact([local.name_prefix, "s3", "results", local.name_suffix])))
  }

  # DynamoDB table names
  dynamodb_tables = {
    verification_results = lower(join("-", compact([local.name_prefix, "dynamodb", "verification-results", local.name_suffix]))),
    layout_metadata      = lower(join("-", compact([local.name_prefix, "dynamodb", "layout-metadata", local.name_suffix]))),
    conversation_history = lower(join("-", compact([local.name_prefix, "dynamodb", "conversation-history", local.name_suffix])))
  }

  # ECR Repository URL (without specific repository)
  ecr_repository_base_url = var.ecr.create_repositories ? "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com" : ""

  # ECR Repositories with standardized naming - one for each Lambda function
  ecr_repositories = {
    # Create repositories for each Lambda function
    initialize = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "initialize", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    fetch_historical_verification = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "fetch-historical-verification", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    fetch_images = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "fetch-images", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    prepare_system_prompt = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "prepare-system-prompt", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    prepare_turn_prompt = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "prepare-turn-prompt", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    invoke_bedrock = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "invoke-bedrock", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    process_turn1_response = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "process-turn1-response", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    process_turn2_response = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "process-turn2-response", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    finalize_results = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "finalize-results", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    store_results = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "store-results", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    notify = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "notify", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    handle_bedrock_error = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "handle-bedrock-error", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    finalize_with_error = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "finalize-with-error", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    render_layout = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "render-layout", local.name_suffix])))
      image_tag_mutability = "IMMUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    }
  }

  # Lambda Functions Configuration
  lambda_functions = {
    initialize = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "initialize", local.name_suffix]))),
      description           = "Initialize verification workflow",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    fetch_historical_verification = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "fetch-historical", local.name_suffix]))),
      description           = "Fetch historical verification data",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    fetch_images = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "fetch-images", local.name_suffix]))),
      description           = "Fetch images for verification",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    prepare_system_prompt = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "prepare-system-prompt", local.name_suffix]))),
      description           = "Prepare system prompt for Bedrock",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    prepare_turn_prompt = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "prepare-turn-prompt", local.name_suffix]))),
      description           = "Prepare turn prompt for Bedrock",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    invoke_bedrock = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "invoke-bedrock", local.name_suffix]))),
      description           = "Invoke Amazon Bedrock",
      memory_size           = 256,
      timeout               = 60,
      environment_variables = {}
    },
    process_turn1_response = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "process-turn1", local.name_suffix]))),
      description           = "Process turn 1 response from Bedrock",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    process_turn2_response = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "process-turn2", local.name_suffix]))),
      description           = "Process turn 2 response from Bedrock",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    finalize_results = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "finalize-results", local.name_suffix]))),
      description           = "Finalize verification results",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    store_results = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "store-results", local.name_suffix]))),
      description           = "Store verification results",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    notify = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "notify", local.name_suffix]))),
      description           = "Send notification about verification results",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    handle_bedrock_error = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "handle-bedrock-error", local.name_suffix]))),
      description           = "Handle Bedrock errors",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    },
    finalize_with_error = {
      name                  = lower(join("-", compact([local.name_prefix, "lambda", "finalize-with-error", local.name_suffix]))),
      description           = "Finalize workflow with error",
      memory_size           = 256,
      timeout               = 30,
      environment_variables = {}
    }
  }

  # Step function state machine name
  step_function_name = lower(join("-", compact([local.name_prefix, "sfn", "verification-workflow", local.name_suffix])))

  # API gateway name
  api_gateway_name = lower(join("-", compact([local.name_prefix, "api", "verification", local.name_suffix])))

  # App Runner service name
  app_runner_service_name = lower(join("-", compact([local.name_prefix, "apprunner", "frontend", local.name_suffix])))

  # Cloudwatch dashboard name
  dashboard_name = lower(join("-", compact([local.name_prefix, "dashboard", "verification", local.name_suffix])))
}

# Always generate a random suffix for resource names
resource "random_string" "suffix" {
  length  = 6
  special = false
  upper   = false
}

# Get current AWS account ID
data "aws_caller_identity" "current" {}
