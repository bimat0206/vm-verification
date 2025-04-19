# infrastructure/main.tf (updated with multi_ecr module)

# After the vpc, s3, dynamodb, and secrets modules, add:

# ECR Repositories for Lambda functions
module "lambda_ecr_repos" {
  source = "./modules/multi_ecr"

  repository_prefix  = var.ecr_repository_name
  environment        = var.environment
  aws_region         = var.aws_region
  image_tag_mutability = var.ecr_image_tag_mutability
  enable_scan_on_push = var.ecr_enable_scan_on_push
  kms_key_arn        = var.ecr_kms_key_arn
  max_image_count    = var.ecr_max_image_count
  
  # Set to true to automatically push placeholder nginx images
  push_placeholder_images = var.push_placeholder_images
  
  tags               = local.common_tags
}

# Lambda functions for the workflow
module "lambda_functions" {
  source = "./modules/multi_lambda"

  name_prefix      = "vending-verification"
  environment      = var.environment
  runtime          = "nodejs18.x"
  architectures    = ["arm64"]
  
  s3_bucket_name   = module.images_bucket.bucket_id
  s3_bucket_arn    = module.images_bucket.bucket_arn
  
  dynamodb_table_name = module.verification_results.table_name
  dynamodb_table_arn  = module.verification_results.table_arn
  
  aws_region       = var.aws_region
  secrets_arn      = module.secrets.secret_arn
  enable_secrets_access = true
  
  bedrock_model_id = "anthropic.claude-3-7-sonnet-20250219-v1:0"
  
  # Comment out or set to false if you want to create Lambda functions
  skip_lambda_function_creation = var.skip_lambda_functions
  
  # Use either a container image from the ECR module or a zip file
  # If using for specific functions, you can uncomment and specify:
  # ecr_image_uri    = module.lambda_ecr_repos.initialize_repository_url
  
  # Or use a map to specify different images for each function
  # Placeholder for now
  filename         = "dummy.zip"
  
  environment_variables = {
    SECRETS_ARN    = module.secrets.secret_arn
    API_STAGE      = "v1"
  }
  
  tags            = local.common_tags
}