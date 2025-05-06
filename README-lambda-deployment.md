# Lambda Deployment with Docker Images

## Issue

The Terraform deployment is failing with errors like:

```
Error: creating Lambda Function: InvalidParameterValueException: Source image 879654127886.dkr.ecr.us-east-1.amazonaws.com/finalize_with_error:latest does not exist. Provide a valid source image.
```

This is happening because the Lambda functions are configured to use Docker images from ECR (Elastic Container Registry), but the images don't exist in the specified ECR repository.

## Solution

There are two approaches to fix this issue:

### Option 1: Build and Push Docker Images First (Recommended)

1. Run the provided script to build and push Docker images to ECR:

```bash
./build-and-push-images.sh
```

This script will:
- Create ECR repositories if they don't exist
- Build Docker images for all Lambda functions using placeholder Nginx images
- Push the images to ECR

2. After the images are pushed, run Terraform to deploy the Lambda functions:

```bash
cd product-approach/iac
terraform apply
```

### Option 2: Modify Terraform Configuration to Use Public Images

If you prefer not to build and push Docker images first, you can modify the Terraform configuration to use public images for the initial deployment:

1. The Lambda module has been updated to use the default_image_uri for all Lambda functions:

```hcl
# In terraform.tfvars:
lambda_functions = {
  create_functions = true
  use_ecr          = false  # Set to false to use default_image_uri instead of ECR
  image_tag        = "latest"
  default_image_uri = "public.ecr.aws/nginx/nginx:latest" # Placeholder image for first deployment
  # ... rest of the configuration ...
}

# In modules/lambda/main.tf:
resource "aws_lambda_function" "this" {
  for_each = var.functions_config

  function_name = each.value.name
  description   = each.value.description
  role          = var.execution_role_arn
  
  # Image configuration for container-based Lambda
  package_type  = "Image"
  image_uri     = var.default_image_uri  # Using default_image_uri from tfvars
  
  # ... rest of the configuration ...
}
```

2. Run Terraform to deploy the Lambda functions:

```bash
cd product-approach/iac
terraform apply
```

3. After the initial deployment, you can build and push your actual Docker images to ECR and update the Lambda functions.

## Next Steps

After the initial deployment, you should:

1. Replace the placeholder Docker images with your actual application images
2. Update the Lambda functions to use the new images

## Notes

- The placeholder images use the public Nginx image (`public.ecr.aws/nginx/nginx:latest`)
- For production deployments, you should replace these with your actual application images
- The ECR repositories are created with the same names as the Lambda functions
