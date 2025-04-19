variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
  default     = "dev"
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for storing images"
  type        = string
  default     = "vending-verification-images"
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table for verification results"
  type        = string
  default     = "VerificationResults"
}

variable "lambda_zip_path" {
  description = "Path to the Lambda function deployment package"
  type        = string
  default     = "backend/dist/vending-verification.zip"
}

variable "lambda_timeout" {
  description = "Lambda function timeout in seconds"
  type        = number
  default     = 30
}

variable "lambda_memory_size" {
  description = "Lambda function memory size in MB"
  type        = number
  default     = 256
}

variable "ecr_repository_name" {
  description = "Name of the ECR repository"
  type        = string
  default     = "vending-verification"
}

variable "ecr_image_tag_mutability" {
  description = "Image tag mutability for the ECR repository"
  type        = string
  default     = "MUTABLE"
}

variable "ecr_enable_scan_on_push" {
  description = "Whether to enable image scanning on push"
  type        = bool
  default     = true
}

variable "ecr_kms_key_arn" {
  description = "ARN of the KMS key for ECR encryption"
  type        = string
  default     = null
}

variable "ecr_max_image_count" {
  description = "Maximum number of images to keep in the ECR repository"
  type        = number
  default     = 30
}

variable "alb_certificate_arn" {
  description = "ARN of the SSL certificate for ALB HTTPS"
  type        = string
  default     = null
}

# VPC Variables
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "172.16.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["172.16.1.0/24", "172.16.2.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["172.16.3.0/24", "172.16.4.0/24"]
}

variable "enable_nat_gateway" {
  description = "Whether to enable NAT Gateway for private subnets"
  type        = bool
  default     = true
}

variable "single_nat_gateway" {
  description = "Whether to use a single NAT Gateway for all private subnets"
  type        = bool
  default     = true
}

variable "ami_id" {
  description = "AMI ID for the EC2 instance"
  type        = string
  default     = "ami-0c7217cdde317cfec" # Amazon Linux 2023 AMI
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}

variable "ecr_image_uri" {
  description = "ECR image URI for Lambda container deployment"
  type        = string
  default     = null
}
