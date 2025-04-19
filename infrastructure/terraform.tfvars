# infrastructure/terraform.tfvars
# AWS Region
aws_region = "us-east-1"

# Environment
environment = "dev"

# S3 Configuration
s3_bucket_name = "vending-verification-images"

# DynamoDB Configuration
dynamodb_table_name = "VerificationResults"

# Lambda Configuration
lambda_timeout = 30
lambda_memory_size = 512

# ECR Configuration
ecr_repository_name = "vending-verification"
ecr_image_tag_mutability = "MUTABLE"
ecr_enable_scan_on_push = true
ecr_max_image_count = 30

# VPC Configuration
vpc_cidr = "172.16.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b"]

# Public Subnets (for ALB)
public_subnet_cidrs = [
  "172.16.1.0/24",
  "172.16.2.0/24"
]

# Private Subnets (for Lambda)
private_subnet_cidrs = [
  "172.16.3.0/24",
  "172.16.4.0/24"
]

# NAT Gateway Configuration
enable_nat_gateway = true
single_nat_gateway = true

# ALB Configuration
alb_certificate_arn = null  # Set this to your SSL certificate ARN if you have one

# Step Functions Configuration
step_functions_execution_timeout = 300

# API Gateway Configuration
enable_api_gateway_logging = true
api_gateway_throttling_rate_limit = 100
api_gateway_throttling_burst_limit = 50

# CloudWatch Configuration
cloudwatch_logs_retention_days = 90
enable_alarm_notifications = false
alarm_email = ""  # Set this to your email address if enable_alarm_notifications is true

# Centralized Tags
default_tags = {
  Project     = "vending-verification"
  Application = "kootoro-vending-verification"
  CostCenter  = "engineering"
  Environment = "dev"  # This will be overridden by the environment variable if different
}

additional_tags = {
  Owner       = "Kootoro"
  Team        = "AI Platform"
  Compliance  = "internal"
}

# Bedrock Configuration
# Do not set bedrock_api_key here for security reasons
# Use environment variables or AWS Secrets Manager
skip_lambda_functions = true 
push_placeholder_images = true