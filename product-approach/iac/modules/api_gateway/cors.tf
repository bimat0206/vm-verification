# modules/api_gateway/cors.tf

# CORS Configuration for all endpoints
locals {
  cors_resources = var.cors_enabled ? [
    aws_api_gateway_resource.verifications_lookup.id,
    aws_api_gateway_resource.verifications.id,
    aws_api_gateway_resource.verification_id.id,
    aws_api_gateway_resource.verification_conversation.id,
    aws_api_gateway_resource.health.id,
    aws_api_gateway_resource.image_view.id,
    aws_api_gateway_resource.image_browser_path.id
  ] : []
  
  # Common CORS headers and methods
  cors_headers = "'*'"
  cors_methods = {
    "verifications_lookup" = "'*'",
    "verifications" = "'*'",
    "verification_id" = "'*'",
    "verification_conversation" = "'*'",
    "health" = "'*'",
    "image_view" = "'*'",
    "image_browser_path" = "'*'"
  }
  
  # Resource to endpoint mapping for easier reference
  resource_to_endpoint = {
    "verifications_lookup" = aws_api_gateway_resource.verifications_lookup.id,
    "verifications" = aws_api_gateway_resource.verifications.id,
    "verification_id" = aws_api_gateway_resource.verification_id.id,
    "verification_conversation" = aws_api_gateway_resource.verification_conversation.id,
    "health" = aws_api_gateway_resource.health.id,
    "image_view" = aws_api_gateway_resource.image_view.id,
    "image_browser_path" = aws_api_gateway_resource.image_browser_path.id
  }
  
  # HTTP methods that need CORS headers in their responses
  cors_enabled_methods = var.cors_enabled ? ["*"] : []
}

# Create OPTIONS method, integration, method response, and integration response for each resource
# These resources are already defined in methods.tf for each endpoint
# The cors.tf file now serves as documentation for CORS-enabled resources
# and provides the common configuration values used across the module
