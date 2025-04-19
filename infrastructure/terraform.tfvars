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
lambda_memory_size = 256

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

# Optional: KMS key ARN for ECR encryption
# ecr_kms_key_arn = "arn:aws:kms:us-east-1:123456789012:key/your-key-id"

# ECR Image URI for Lambda container deployment
# Uncomment and set this to your ECR image URI (e.g., "123456789012.dkr.ecr.us-east-1.amazonaws.com/vending-verification:latest")
# ecr_image_uri = "YOUR_ECR_IMAGE_URI"

# VPC and Subnet IDs for ALB will be referenced from VPC module outputs
