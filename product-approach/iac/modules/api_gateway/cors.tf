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
  cors_headers = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'"
  cors_methods = {
    "verifications_lookup" = "'GET,OPTIONS'",
    "verifications" = "'POST,OPTIONS'",
    "verification_id" = "'GET,OPTIONS'",
    "verification_conversation" = "'POST,OPTIONS'",
    "health" = "'GET,OPTIONS'",
    "image_view" = "'GET,OPTIONS'",
    "image_browser_path" = "'GET,OPTIONS'"
  }
}

# Create OPTIONS method, integration, method response, and integration response for each resource
# These resources are already defined in methods.tf for each endpoint
# The cors.tf file now serves as documentation for CORS-enabled resources
# and provides the common configuration values used across the module