# Terraform Variables Backup

**Date:** 2025-05-29 16:44:54
**Directory:** .
**File:** terraform.tfvars

## Variables Content
```hcl
# General Configuration
aws_region   = "us-east-1"
project_name = "kootoro"
environment  = "dev"
# resource_name_suffix is intentionally not set to use the auto-generated random suffix

additional_tags = {
  Owner   = "ManhDT"
  Project = "VendingMachineVerification"
}

# S3 Bucket Configuration
s3_buckets = {
  create_buckets = true
  force_destroy  = false # Set to false in production to prevent accidental deletion
  lifecycle_rules = {
    reference = [
      {
        id                                     = "expire-after-24-months"
        enabled                                = true
        expiration_days                        = 730
        noncurrent_version_expiration_days     = 90
        abort_incomplete_multipart_upload_days = 7
      }
    ],
    checking = [
      {
        id                                     = "expire-after-24-months"
        enabled                                = true
        expiration_days                        = 730
        noncurrent_version_expiration_days     = 90
        abort_incomplete_multipart_upload_days = 7
      }
    ],
    results = [
      {
        id                                     = "expire-after-12-months"
        enabled                                = true
        expiration_days                        = 365
        noncurrent_version_expiration_days     = 90
        abort_incomplete_multipart_upload_days = 7
      }
    ]
        temp_base64 = [
      {
        id                                     = "cleanup-temp-base64"
        enabled                                = true
        prefix                                 = "temp-base64/"
        expiration_days                        = 5
        abort_incomplete_multipart_upload_days = 1
      }
    ]
  }
}

# DynamoDB Configuration
dynamodb_tables = {
  create_tables          = true
  billing_mode           = "PAY_PER_REQUEST" # Changed to PAY_PER_REQUEST for scalability
  read_capacity          = 10
  write_capacity         = 10
  point_in_time_recovery = true
}

# ECR Configuration
ecr = {
  create_repositories = true
  repositories = {
    # Each Lambda function will get its repository
    # Production-specific settings
    initialize = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    fetch_historical_verification = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    fetch_images = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    prepare_system_prompt = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    execute_turn1_combined = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    execute_turn2_combined = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    finalize_results = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    store_results = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    notify = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    handle_bedrock_error = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    finalize_with_error = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    },
    render_layout = {
      force_delete         = false
      scan_on_push         = true
      image_tag_mutability = "mutable"
    }
  }
}

# Lambda Configuration
lambda_functions = {
  create_functions  = true
  use_ecr           = false # Set to false for initial deployment, then change to true after ECR repos are created and images are pushed
  image_tag         = "latest"
  default_image_uri = "879654127886.dkr.ecr.us-east-1.amazonaws.com/vending-render:latest" # Fallback image for all Lambda functions
  architectures     = ["arm64"]
  memory_sizes = {
    initialize                    = 1024
    fetch_historical_verification = 1024
    fetch_images                  = 1536
    prepare_system_prompt         = 1024
    execute_turn1_combined        = 1024
    execute_turn2_combined        = 1536
    finalize_results              = 1024
    store_results                 = 1024
    notify                        = 512
    handle_bedrock_error          = 512
    finalize_with_error           = 512
    render_layout                 = 2048
    list_verifications            = 1024
    get_verification              = 1024
    get_conversation              = 1024
    health_check                  = 512
  }
  timeouts = {
    initialize                    = 30
    fetch_historical_verification = 30
    fetch_images                  = 60
    prepare_system_prompt         = 30
    execute_turn1_combined        = 120
    execute_turn2_combined        = 150
    finalize_results              = 90
    store_results                 = 60
    notify                        = 60
    handle_bedrock_error          = 60
    finalize_with_error           = 60
    render_layout                 = 120
    list_verifications            = 30
    get_verification              = 30
    get_conversation              = 30
    health_check                  = 30
  }
  log_retention_days            = 90
  s3_trigger_functions          = ["render_layout"]
  eventbridge_trigger_functions = []
}

# API Gateway Configuration
api_gateway = {
  create_api_gateway     = true
  stage_name             = "v1"
  throttling_rate_limit  = 200
  throttling_burst_limit = 400
  cors_enabled           = true
  metrics_enabled        = true
  use_api_key            = true # Enable API key authentication
}

# Step Functions Configuration
# Step Functions Configuration
step_functions = {
  create_step_functions = true
  log_level             = "ALL"
  enable_x_ray_tracing  = true
}

# ECS Streamlit Configuration
streamlit_frontend = {
  create_streamlit                 = true
  service_name                     = "vm-fe"
  image_uri                        = "879654127886.dkr.ecr.us-east-1.amazonaws.com/vending-verification-streamlit-app:latest" # Replace with your image
  image_repository_type            = "ECR"
  cpu                              = 1024 # 1 vCPU = 1024 CPU units
  memory                           = 2048 # 2 GB = 2048 MB
  port                             = 8501
  auto_deployments_enabled         = false
  enable_auto_scaling              = true
  min_size                         = 1
  max_size                         = 3
  max_capacity                     = 10
  cpu_threshold                    = 70
  memory_threshold                 = 70
  theme_mode                       = "dark"
  log_retention_days               = 30
  health_check_path                = "/_stcore/health"
  health_check_interval            = 30
  health_check_timeout             = 5
  health_check_healthy_threshold   = 2
  health_check_unhealthy_threshold = 3
  enable_https                     = false
  internal_alb                     = false
  enable_container_insights        = true
  enable_execute_command           = true
  environment_variables = {
    STREAMLIT_THEME_PRIMARY_COLOR              = "#FF4B4B"
    STREAMLIT_THEME_BACKGROUND_COLOR           = "#0E1117"
    STREAMLIT_THEME_SECONDARY_BACKGROUND_COLOR = "#262730"
    STREAMLIT_THEME_TEXT_COLOR                 = "#FAFAFA"
    STREAMLIT_THEME_FONT                       = "sans serif"
    API_ENDPOINT                               = "" # Will be populated from API Gateway endpoint
  }
}

# VPC Configuration
vpc = {
  create_vpc         = true
  vpc_cidr           = "172.1.0.0/16"
  availability_zones = ["us-east-1a", "us-east-1b"]
  create_nat_gateway = true
}

# Bedrock Configuration
bedrock = {
  model_id          = "us.anthropic.claude-3-7-sonnet-20250219-v1:0"
  anthropic_version = "bedrock-2023-05-31"
  max_tokens        = 24000
  budget_tokens     = 16000
}

# Monitoring Configuration
monitoring = {
  create_dashboard      = true
  log_retention_days    = 12
  alarm_email_endpoints = ["manh.hoang@renovacloud.com" ]
}
```
