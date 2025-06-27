provider "aws" {
  region = var.aws_region

  default_tags {
    tags = local.common_tags
  }
}

provider "random" {}


terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0.0"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.0.0"
    }
  }
    # Using local backend for testing
  backend "s3" {
    bucket  = "vending-verification-terraform-state"
    key     = "state/terraform.tfstate"
    region  = "us-east-1"
  }
}