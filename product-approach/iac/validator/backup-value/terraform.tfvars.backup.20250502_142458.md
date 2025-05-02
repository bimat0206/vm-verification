# Terraform Variables Backup

**Date:** 2025-05-02 14:24:58
**Directory:** .
**File:** terraform.tfvars

## Variables Content
```hcl
# General Configuration
aws_region          = "us-east-1"
project_name        = "kootoro"
environment         = "dev"
# resource_name_suffix is intentionally not set to use the auto-generated random suffix

additional_tags = {
  Owner       = "ManhDT"
  Project     = "VendingMachineVerification"
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
  }
}

# DynamoDB Configuration
dynamodb_tables = {
  create_tables         = true
  billing_mode          = "PROVISIONED"
  read_capacity         = 10
  write_capacity        = 10
  point_in_time_recovery = true
}

# ECR Configuration
ecr = {
  create_repositories = true
  repositories = {
    # Each Lambda function will get its repository
    # Production-specific settings
    initialize = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    fetch_historical_verification = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    fetch_images = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    prepare_system_prompt = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    prepare_turn_prompt = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    invoke_bedrock = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
      lifecycle_policy = <<EOF
{
  "rules": [
    {
      "rulePriority": 1,
      "description": "Keep only tagged images and remove untagged after 7 days",
      "selection": {
        "tagStatus": "untagged",
        "countType": "sinceImagePushed",
        "countUnit": "days",
        "countNumber": 7
      },
      "action": {
        "type": "expire"
      }
    },
    {
      "rulePriority": 2,
      "description": "Keep the last 10 images",
      "selection": {
        "tagStatus": "any",
        "countType": "imageCountMoreThan",
        "countNumber": 10
      },
      "action": {
        "type": "expire"
      }
    }
  ]
}
EOF
    },
    process_turn1_response = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    process_turn2_response = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    finalize_results = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    store_results = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    notify = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    handle_bedrock_error = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    finalize_with_error = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    },
    render_layout = {
      force_delete = false
      scan_on_push = true
      image_tag_mutability = "IMMUTABLE"
    }
  }
}

# Lambda Configuration
lambda_functions = {
  create_functions = true
  use_ecr          = false  # Set to false to use default_image_uri instead of ECR
  image_tag        = "latest"
  default_image_uri = "879654127886.dkr.ecr.us-east-1.amazonaws.com/vending-render:latest" # Placeholder image for first deployment
  architectures    = ["arm64"]
  memory_sizes = {
    initialize                  = 1024
    fetch_historical_verification = 1024
    fetch_images                = 1536
    prepare_system_prompt       = 1024
    prepare_turn_prompt         = 512
    invoke_bedrock              = 2048
    process_turn1_response      = 1024
    process_turn2_response      = 1024
    finalize_results            = 1024
    store_results               = 1024
    notify                      = 512
    handle_bedrock_error        = 512
    finalize_with_error         = 512
    render_layout               = 2048
  }
  timeouts = {
    initialize                  = 30
    fetch_historical_verification = 30
    fetch_images                = 60
    prepare_system_prompt       = 30
    prepare_turn_prompt         = 30
    invoke_bedrock              = 150
    process_turn1_response      = 90
    process_turn2_response      = 90
    finalize_results            = 90
    store_results               = 60
    notify                      = 60
    handle_bedrock_error        = 60
    finalize_with_error         = 60
    render_layout               = 120
  }
  log_retention_days = 90
  s3_trigger_functions = ["render_layout"]
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
}

# Step Functions Configuration
step_functions = {
  create_step_functions = true
  log_level             = "ALL"
}

# App Runner Configuration
app_runner = {
  create_app_runner = true
  image_uri         = "public.ecr.aws/nginx/nginx:latest" # Placeholder image for first deployment
  cpu               = 2
  memory            = 4096
  environment_variables = {
    NODE_ENV = "production"
  }
  auto_deployments_enabled = false
}

# Bedrock Configuration
bedrock = {
  model_id          = "anthropic.claude-3-7-sonnet-20250219"
  anthropic_version = "bedrock-2023-05-31"
  max_tokens        = 24000
  budget_tokens     = 16000
}

# Monitoring Configuration
monitoring = {
  create_dashboard    = true
  log_retention_days  = 90
  alarm_email_endpoints = ["ops-alerts@example.com", "on-call@example.com"]
}```
