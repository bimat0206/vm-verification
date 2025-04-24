# infrastructure/main.tf
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
  
  # Apply default tags to all resources
  default_tags {
    tags = merge(
      var.default_tags,
      {
        Environment = var.environment
      }
    )
  }
}