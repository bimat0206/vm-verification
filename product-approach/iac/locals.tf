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
    results   = lower(join("-", compact([local.name_prefix, "s3", "results", local.name_suffix]))),
    state     = lower(join("-", compact([local.name_prefix, "s3", "state", local.name_suffix])))
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
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    fetch_historical_verification = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "fetch-historical-verification", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    fetch_images = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "fetch-images", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    prepare_system_prompt = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "prepare-system-prompt", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    execute_turn1_combined = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "execute-turn1-combined", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    execute_turn2_combined = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "execute-turn2-combined", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    finalize_results = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "finalize-results", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },

    finalize_with_error = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "finalize-with-error", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    render_layout = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "render-layout", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    # Add new repositories for dedicated functions
    api_verifications_list = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "api-verifications-list", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },

    api_get_conversation = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "api-get-conversation", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    health_check = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "health-check", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    api_images_browser = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "api-images-browser", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
    api_images_view = {
      name                 = lower(join("-", compact([local.name_prefix, "ecr", "api-images-view", local.name_suffix])))
      image_tag_mutability = "MUTABLE"
      scan_on_push         = true
      force_delete         = false
      encryption_type      = "AES256"
      kms_key              = null
      lifecycle_policy     = null
      repository_policy    = null
    },
  }

  # Lambda Functions Configuration
  lambda_functions = {
    # Update to locals.tf to add Step Functions ARN to initialize Lambda
    initialize = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "initialize", local.name_suffix]))),
      description = "Initialize verification workflow and trigger Step Functions execution",
      memory_size = 512,
      timeout     = 30,
      environment_variables = {
        STEP_FUNCTIONS_STATE_MACHINE_ARN = "arn:aws:states:${var.aws_region}:${data.aws_caller_identity.current.account_id}:stateMachine:${local.step_function_name}"
        DYNAMODB_VERIFICATION_TABLE      = local.dynamodb_tables.verification_results
        DYNAMODB_CONVERSATION_TABLE      = local.dynamodb_tables.conversation_history
        DYNAMODB_LAYOUT_TABLE            = local.dynamodb_tables.layout_metadata
        REFERENCE_BUCKET                 = local.s3_buckets.reference
        CHECKING_BUCKET                  = local.s3_buckets.checking
        RESULTS_BUCKET                   = local.s3_buckets.results
        STATE_BUCKET                     = local.s3_buckets.state
      }
    },
    fetch_historical_verification = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "fetch-historical", local.name_suffix]))),
      description = "Fetch historical verification data",
      memory_size = 256,
      timeout     = 30,
      environment_variables = {
        DYNAMODB_VERIFICATION_TABLE = local.dynamodb_tables.verification_results
        DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
        LOG_LEVEL                   = "INFO"
        STATE_BUCKET                = local.s3_buckets.state
      }
    },
    fetch_images = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "fetch-images", local.name_suffix]))),
      description = "Fetch images for verification",
      memory_size = 256,
      timeout     = 30,
      environment_variables = {
        REFERENCE_BUCKET       = local.s3_buckets.reference
        CHECKING_BUCKET        = local.s3_buckets.checking
        DYNAMODB_LAYOUT_TABLE  = local.dynamodb_tables.layout_metadata
        LOG_LEVEL              = "INFO"
        STATE_BUCKET           = local.s3_buckets.state
        MAX_INLINE_BASE64_SIZE = "1048576" # 2MB in bytes
      }
    },
    prepare_system_prompt = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "prepare-system-prompt", local.name_suffix]))),
      description = "Prepare system prompt for Bedrock",
      memory_size = 256,
      timeout     = 30,
      environment_variables = {
        ANTHROPIC_VERSION           = var.bedrock.anthropic_version
        BEDROCK_MODEL               = var.bedrock.model_id
        MAX_TOKENS                  = var.bedrock.max_tokens
        BUDGET_TOKENS               = var.bedrock.budget_tokens
        THINKING_TYPE               = "enable"
        DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
        LOG_LEVEL                   = "INFO"
        REFERENCE_BUCKET            = local.s3_buckets.reference
        CHECKING_BUCKET             = local.s3_buckets.checking
        STATE_BUCKET                = local.s3_buckets.state
      }
    },
    execute_turn1_combined = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "execute-turn1-combined", local.name_suffix]))),
      description = "Combined function: prepare turn1 prompt, execute Bedrock call, and process response",
      memory_size = 1024,
      timeout     = 120,
      environment_variables = {
        ANTHROPIC_VERSION           = var.bedrock.anthropic_version
        BEDROCK_MODEL               = var.bedrock.model_id
        MAX_TOKENS                  = var.bedrock.max_tokens
        BUDGET_TOKENS               = var.bedrock.budget_tokens
        THINKING_TYPE               = "enabled"
        DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
        DYNAMODB_VERIFICATION_TABLE = local.dynamodb_tables.verification_results
        REFERENCE_BUCKET            = local.s3_buckets.reference
        CHECKING_BUCKET             = local.s3_buckets.checking
        STATE_BUCKET                = local.s3_buckets.state
        LOG_LEVEL                   = "INFO"
        TURN_NUMBER                 = "1"
        TEMPLATE_BASE_PATH          = "/opt/templates"
        RETRY_MAX_ATTEMPTS          = "3"
        RETRY_BASE_DELAY            = "2000"
        # Bedrock timeout configuration - increased to handle large payloads
        BEDROCK_CONNECT_TIMEOUT_SEC = "30"
        BEDROCK_CALL_TIMEOUT_SEC    = "120"
      }
    },
    execute_turn2_combined = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "execute-turn2-combined", local.name_suffix]))),
      description = "Combined function: prepare turn2 prompt, execute Bedrock call, and process response",
      memory_size = 1536,
      timeout     = 150,
      environment_variables = {
        ANTHROPIC_VERSION           = var.bedrock.anthropic_version
        BEDROCK_MODEL               = var.bedrock.model_id
        MAX_TOKENS                  = var.bedrock.max_tokens
        BUDGET_TOKENS               = var.bedrock.budget_tokens
        THINKING_TYPE               = "enabled"
        DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
        DYNAMODB_VERIFICATION_TABLE = local.dynamodb_tables.verification_results
        REFERENCE_BUCKET            = local.s3_buckets.reference
        CHECKING_BUCKET             = local.s3_buckets.checking
        STATE_BUCKET                = local.s3_buckets.state
        LOG_LEVEL                   = "INFO"
        TURN_NUMBER                 = "2"
        TEMPLATE_BASE_PATH          = "/opt/templates"
        RETRY_MAX_ATTEMPTS          = "3"
        RETRY_BASE_DELAY            = "2000"
        # Bedrock timeout configuration - increased to handle large payloads
        BEDROCK_CONNECT_TIMEOUT_SEC = "30"
        BEDROCK_CALL_TIMEOUT_SEC    = "120"
      }
    },
    finalize_results = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "finalize-results", local.name_suffix]))),
      description = "Finalize verification results",
      memory_size = 256,
      timeout     = 30,
      # Add explicit image URI to ensure it's not removed when updating
      image_uri = "879654127886.dkr.ecr.us-east-1.amazonaws.com/vending-render:latest"
      environment_variables = {
        DYNAMODB_VERIFICATION_TABLE = local.dynamodb_tables.verification_results
        DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
        LOG_LEVEL                   = "INFO"
        STATE_BUCKET                = local.s3_buckets.state
      }
    },

    finalize_with_error = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "finalize-with-error", local.name_suffix]))),
      description = "Finalize workflow with error",
      memory_size = 256,
      timeout     = 30,
      environment_variables = {
        DYNAMODB_VERIFICATION_TABLE = local.dynamodb_tables.verification_results
        LOG_LEVEL                   = "INFO"
        DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
        STATE_BUCKET                = local.s3_buckets.state
      }
    },
    # Add new Lambda functions
    api_verifications_list = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "api-verifications-list", local.name_suffix]))),
      description = "API endpoint for listing verification results with filtering and pagination",
      memory_size = 512,
      timeout     = 30,
      environment_variables = {
        DYNAMODB_VERIFICATION_TABLE = local.dynamodb_tables.verification_results
        DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
        LOG_LEVEL                   = "INFO"
      }
    },

    api_get_conversation = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "api-get-conversation", local.name_suffix]))),
      description = "Get verification conversation history",
      memory_size = 256,
      timeout     = 30,
      environment_variables = {
        DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
        DYNAMODB_VERIFICATION_TABLE = local.dynamodb_tables.verification_results
        LOG_LEVEL                   = "INFO"
      }
    },
    health_check = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "health-check", local.name_suffix]))),
      description = "System health check",
      memory_size = 256,
      timeout     = 30,
      environment_variables = {
        DYNAMODB_VERIFICATION_TABLE = local.dynamodb_tables.verification_results
        DYNAMODB_CONVERSATION_TABLE = local.dynamodb_tables.conversation_history
        REFERENCE_BUCKET            = local.s3_buckets.reference
        CHECKING_BUCKET             = local.s3_buckets.checking
        RESULTS_BUCKET              = local.s3_buckets.results
        BEDROCK_MODEL               = var.bedrock.model_id
        LOG_LEVEL                   = "INFO"
      }
    },
    render_layout = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "render-layout", local.name_suffix]))),
      description = "Render layout visualization",
      memory_size = 2048,
      timeout     = 120,
      environment_variables = {
        REFERENCE_BUCKET = local.s3_buckets.reference
        CHECKING_BUCKET  = local.s3_buckets.checking
        RESULTS_BUCKET   = local.s3_buckets.results
        STATE_BUCKET     = local.s3_buckets.state
        LOG_LEVEL        = "INFO"
      }
    },
    api_images_browser = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "api-images-browser", local.name_suffix]))),
      description = "Browse images in S3 buckets via REST API",
      memory_size = 256,
      timeout     = 30,
      image_uri   = "879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-api-images-browser-f6d3xl:latest"
      environment_variables = {
        REFERENCE_BUCKET = local.s3_buckets.reference
        CHECKING_BUCKET  = local.s3_buckets.checking
        LOG_LEVEL        = "INFO"
      }
    },
    api_images_view = {
      name        = lower(join("-", compact([local.name_prefix, "lambda", "api-images-view", local.name_suffix]))),
      description = "Generate presigned URLs for S3 image viewing via REST API",
      memory_size = 128,
      timeout     = 30,
      image_uri   = "879654127886.dkr.ecr.us-east-1.amazonaws.com/kootoro-dev-ecr-api-images-view-f6d3xl:latest"
      environment_variables = {
        REFERENCE_BUCKET = local.s3_buckets.reference
        CHECKING_BUCKET  = local.s3_buckets.checking
        LOG_LEVEL        = "INFO"
      }
    }
  }

  # Step function state machine name
  step_function_name = lower(join("-", compact([local.name_prefix, "sfn", "verification-workflow", local.name_suffix])))

  # API gateway name
  api_gateway_name = lower(join("-", compact([local.name_prefix, "api", "verification", local.name_suffix])))

  # VPC name
  vpc_name = lower(join("-", compact([local.name_prefix, "vpc", local.name_suffix])))

  # ECS service name
  ecs_service_name = lower(join("-", compact([local.name_prefix, "ecs", "streamlit", local.name_suffix])))

  # ALB name
  alb_name = lower(join("-", compact([local.name_prefix, "alb", "streamlit", local.name_suffix])))

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
