terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "vending-verification-terraform-state"
    key            = "backend/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region
}


# VPC Module
module "vpc" {
  source = "./modules/vpc"

  environment         = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  public_subnet_cidrs = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
  enable_nat_gateway = var.enable_nat_gateway
  single_nat_gateway = var.single_nat_gateway
  tags = {
    Project     = "vending-verification"
    ManagedBy   = "terraform"
  }
}

# S3 bucket for storing images
module "images_bucket" {
  source = "./modules/s3"

  bucket_name     = var.s3_bucket_name
  environment     = var.environment
  enable_versioning = true
  encryption_algorithm = "AES256"
  tags = {
    Project     = "vending-verification"
    ManagedBy   = "terraform"
  }
}

# S3 bucket for storing Terraform state


# DynamoDB table for verification results
module "verification_results" {
  source = "./modules/dynamodb"

  table_name     = var.dynamodb_table_name
  environment    = var.environment
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "id"
  enable_streams = true
  attributes = [
    {
      name = "id"
      type = "S"
    }
  ]
  tags = {
    Project     = "vending-verification"
    ManagedBy   = "terraform"
  }
}

# Lambda function
module "api_lambda" {
  source = "./modules/lambda"

  function_name    = "vending-verification-api"
  environment      = var.environment
  architectures    = ["arm64"]
  timeout          = var.lambda_timeout
  memory_size      = var.lambda_memory_size
  ecr_image_uri    = "${module.ecr.repository_url}:latest"
  image_command    = ["./main"]
  
  # These are required for Zip deployment but not used with container images
  # Keeping them defined but with dummy values for module compatibility
  filename         = "dummy.zip"
  handler          = "dummy"
  runtime          = "provided.al2"
  environment_variables = {
    # AWS_REGION is a reserved environment variable and cannot be set manually
    # The region is automatically available to the Lambda function
    S3_BUCKET       = module.images_bucket.bucket_id
    DYNAMO_TABLE    = module.verification_results.table_name
    BEDROCK_MODEL   = "anthropic.claude-3-sonnet-20240229-v1:0"
    PORT            = "5000"
  }
  additional_policy_actions = [
    "s3:PutObject",
    "s3:GetObject",
    "s3:DeleteObject",
    "s3:ListBucket",
    "dynamodb:PutItem",
    "dynamodb:GetItem",
    "dynamodb:DeleteItem",
    "dynamodb:Query",
    "dynamodb:Scan",
    "bedrock:InvokeModel"
  ]
  additional_policy_resources = [
    module.images_bucket.bucket_arn,
    "${module.images_bucket.bucket_arn}/*",
    module.verification_results.table_arn,
    "arn:aws:bedrock:${var.aws_region}::foundation-model/anthropic.claude-3-sonnet-20240229-v1:0"
  ]
  tags = {
    Project     = "vending-verification"
    ManagedBy   = "terraform"
  }
}

# Security group for Lambda
resource "aws_security_group" "lambda" {
  name_prefix = "vending-verification-lambda-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 5000
    to_port     = 5000
    protocol    = "tcp"
    cidr_blocks = [module.vpc.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "vending-verification-lambda-sg"
    Environment = var.environment
    Project     = "vending-verification"
    ManagedBy   = "terraform"
  }
}

# Application Load Balancer
module "alb" {
  source = "./modules/alb"

  name            = "vending-verification-alb"
  environment     = var.environment
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.public_subnet_ids
  internal        = false
  target_port     = 80
  health_check_path = "/health"
  certificate_arn = var.alb_certificate_arn
  lambda_arn      = module.api_lambda.function_arn
  lambda_name     = module.api_lambda.function_name
  tags = {
    Project     = "vending-verification"
    ManagedBy   = "terraform"
  }
}

# ECR Repository for Docker images
module "ecr" {
  source = "./modules/ecr"

  repository_name     = var.ecr_repository_name
  environment        = var.environment
  image_tag_mutability = var.ecr_image_tag_mutability
  enable_scan_on_push = var.ecr_enable_scan_on_push
  kms_key_arn        = var.ecr_kms_key_arn
  max_image_count    = var.ecr_max_image_count
  tags = {
    Project     = "vending-verification"
    ManagedBy   = "terraform"
  }
}
