# modules/api_gateway/resources.tf

# Create the main API Gateway REST API
resource "aws_api_gateway_rest_api" "api" {
  name        = var.api_name
  description = var.api_description
  
  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = var.common_tags
}

# Create API resources for each endpoint path

# Root resource is created by default
# /api
resource "aws_api_gateway_resource" "api" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "api"
}

# /api/v1
resource "aws_api_gateway_resource" "v1" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.api.id
  path_part   = "v1"
}

# /api/v1/verifications
resource "aws_api_gateway_resource" "verifications" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "verifications"
}

# /api/v1/verifications/lookup
resource "aws_api_gateway_resource" "verifications_lookup" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.verifications.id
  path_part   = "lookup"
}

# /api/v1/verifications/{verificationId}
resource "aws_api_gateway_resource" "verification_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.verifications.id
  path_part   = "{verificationId}"
}

# /api/v1/verifications/{verificationId}/conversation
resource "aws_api_gateway_resource" "verification_conversation" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.verification_id.id
  path_part   = "conversation"
}

# /api/v1/health
resource "aws_api_gateway_resource" "health" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "health"
}

# /api/v1/images
resource "aws_api_gateway_resource" "images" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "images"
}

# /api/v1/images/{key}
resource "aws_api_gateway_resource" "image_key" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.images.id
  path_part   = "{key}"
}

# /api/v1/images/{key}/view
resource "aws_api_gateway_resource" "image_view" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.image_key.id
  path_part   = "view"
}

# /api/v1/images/browser
resource "aws_api_gateway_resource" "image_browser" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.images.id
  path_part   = "browser"
}

# /api/v1/images/browser/{path+}
resource "aws_api_gateway_resource" "image_browser_path" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.image_browser.id
  path_part   = "{path+}"
}