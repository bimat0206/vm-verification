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
}

module "cors" {
  source   = "../cors"
  count    = length(local.cors_resources)
  
  api_id          = aws_api_gateway_rest_api.api.id
  api_resource_id = local.cors_resources[count.index]
  allow_origin    = local.cors_origin
}