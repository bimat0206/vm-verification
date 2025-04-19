# infrastructure/modules/streamlit_frontend/main.tf

# Create ECR repository for Streamlit container
resource "aws_ecr_repository" "streamlit_app" {
  name                 = "${var.name_prefix}-streamlit-app"
  image_tag_mutability = var.image_tag_mutability

  image_scanning_configuration {
    scan_on_push = var.enable_scan_on_push
  }

  encryption_configuration {
    encryption_type = var.kms_key_arn != null ? "KMS" : "AES256"
    kms_key         = var.kms_key_arn
  }

  tags = merge(
    {
      Name        = "${var.name_prefix}-streamlit-app"
      Environment = var.environment
    },
    var.tags
  )
}

# Add lifecycle policy to ECR repository
resource "aws_ecr_lifecycle_policy" "streamlit_app_policy" {
  repository = aws_ecr_repository.streamlit_app.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1,
        action = {
          type = "expire"
        }
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = var.max_image_count
        }
        description = "Keep only the latest ${var.max_image_count} images"
      }
    ]
  })
}

# Create a secret to store Streamlit app configuration
resource "aws_secretsmanager_secret" "streamlit_config" {
  name        = "${var.name_prefix}-streamlit-config"
  description = "Configuration for Streamlit frontend application"
  
  recovery_window_in_days = 7
  
  tags = merge(
    {
      Name        = "${var.name_prefix}-streamlit-config"
      Environment = var.environment
    },
    var.tags
  )
}

# Store API endpoints and other configuration in the secret
resource "aws_secretsmanager_secret_version" "streamlit_config_version" {
  secret_id = aws_secretsmanager_secret.streamlit_config.id
  
  secret_string = jsonencode({
    api_endpoint          = var.api_endpoint,
    dynamodb_table_name   = var.dynamodb_table_name,
    s3_bucket_name        = var.s3_bucket_name,
    step_functions_arn    = var.step_functions_arn,
    additional_config     = var.additional_config
  })
}

# IAM role for the App Runner service
resource "aws_iam_role" "app_runner_role" {
  name = "${var.name_prefix}-apprunner-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "build.apprunner.amazonaws.com"
        }
      },
      {
        Action = "sts:AssumeRole",
        Effect = "Allow",
        Principal = {
          Service = "tasks.apprunner.amazonaws.com"
        }
      }
    ]
  })
  
  tags = merge(
    {
      Name        = "${var.name_prefix}-apprunner-role"
      Environment = var.environment
    },
    var.tags
  )
}

# Policy for App Runner to access ECR
resource "aws_iam_policy" "app_runner_ecr_policy" {
  name        = "${var.name_prefix}-apprunner-ecr-policy"
  description = "Allow App Runner to pull images from ECR"
  
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchCheckLayerAvailability",
          "ecr:BatchGetImage",
          "ecr:DescribeImages",
          "ecr:GetAuthorizationToken"
        ],
        Resource = "*"
      }
    ]
  })
}

# Policy for App Runner to access Secrets Manager
resource "aws_iam_policy" "app_runner_secrets_policy" {
  name        = "${var.name_prefix}-apprunner-secrets-policy"
  description = "Allow App Runner to read secrets from Secrets Manager"
  
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "secretsmanager:GetSecretValue"
        ],
        Resource = [
          aws_secretsmanager_secret.streamlit_config.arn
        ]
      }
    ]
  })
}

# Attach policies to IAM role
resource "aws_iam_role_policy_attachment" "app_runner_ecr_attachment" {
  role       = aws_iam_role.app_runner_role.name
  policy_arn = aws_iam_policy.app_runner_ecr_policy.arn
}

resource "aws_iam_role_policy_attachment" "app_runner_secrets_attachment" {
  role       = aws_iam_role.app_runner_role.name
  policy_arn = aws_iam_policy.app_runner_secrets_policy.arn
}

# Create App Runner service using the ECR repository
resource "aws_apprunner_service" "streamlit_service" {
  service_name = "${var.name_prefix}-streamlit"
  
  source_configuration {
    authentication_configuration {
      access_role_arn = aws_iam_role.app_runner_role.arn
    }
    
    image_repository {
      image_configuration {
        port = var.container_port
        runtime_environment_variables = {
          SECRET_ARN = aws_secretsmanager_secret.streamlit_config.arn
          REGION     = var.aws_region
        }
      }
      image_identifier      = "${aws_ecr_repository.streamlit_app.repository_url}:${var.image_tag}"
      image_repository_type = "ECR"
    }
    
    auto_deployments_enabled = var.auto_deployments_enabled
  }
  
  instance_configuration {
    cpu               = var.cpu
    memory            = var.memory
    instance_role_arn = aws_iam_role.app_runner_role.arn
  }
  
health_check_configuration {
  protocol = "HTTP"
  path     = "/"
  interval = 20
  timeout  = 10
  healthy_threshold = 1
  unhealthy_threshold = 5
}
  
  tags = merge(
    {
      Name        = "${var.name_prefix}-streamlit"
      Environment = var.environment
    },
    var.tags
  )
  
  # Wait for initial deployment to complete
  auto_scaling_configuration_arn = aws_apprunner_auto_scaling_configuration_version.app_scaling.arn
}

# App Runner auto scaling configuration
# App Runner auto scaling configuration
resource "aws_apprunner_auto_scaling_configuration_version" "app_scaling" {
  auto_scaling_configuration_name = "${substr(var.name_prefix, 0, 20)}-scaling"
  
  max_concurrency = var.max_concurrency
  max_size        = var.max_size
  min_size        = var.min_size
  
  tags = merge(
    {
      Name        = "${var.name_prefix}-apprunner-scaling"
      Environment = var.environment
    },
    var.tags
  )
}

# Null resource to build and push initial Docker image to ECR
# Note: This is a local-exec that requires Docker and AWS CLI to be installed
resource "null_resource" "docker_build_push" {
  count = var.build_and_push_image ? 1 : 0
  
  triggers = {
    ecr_repository_url = aws_ecr_repository.streamlit_app.repository_url
    source_code_hash   = var.source_code_hash != "" ? var.source_code_hash : timestamp()
  }
  
  provisioner "local-exec" {
    working_dir = var.app_source_path
    command     = <<EOF
      aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${aws_ecr_repository.streamlit_app.repository_url}
      docker build -t ${aws_ecr_repository.streamlit_app.repository_url}:${var.image_tag} .
      docker push ${aws_ecr_repository.streamlit_app.repository_url}:${var.image_tag}
    EOF
  }
  
  depends_on = [aws_ecr_repository.streamlit_app]
}
# Additional policy for App Runner
resource "aws_iam_policy" "app_runner_additional_policy" {
  name        = "${var.name_prefix}-apprunner-additional-policy"
  description = "Additional permissions for App Runner"
  
  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Resource = "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "app_runner_additional_attachment" {
  role       = aws_iam_role.app_runner_role.name
  policy_arn = aws_iam_policy.app_runner_additional_policy.arn
}