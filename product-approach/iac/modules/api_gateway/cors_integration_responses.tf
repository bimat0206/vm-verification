# modules/api_gateway/cors_integration_responses.tf

# Integration responses for GET methods to include CORS headers
# These are only created when CORS is enabled

# 1. Verifications Lookup - GET /api/verifications/lookup
resource "aws_api_gateway_integration_response" "verifications_lookup_get" {
  count = var.cors_enabled ? 1 : 0

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications_lookup.id
  http_method = aws_api_gateway_method.verifications_lookup_get.http_method
  status_code = aws_api_gateway_method_response.verifications_lookup_get.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'",
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Api-Key,Authorization'"
  }

  depends_on = [aws_api_gateway_integration.verifications_lookup_get]
}

# 2. Verifications - GET /api/verifications
resource "aws_api_gateway_integration_response" "verifications_get" {
  count = var.cors_enabled ? 1 : 0

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_get.http_method
  status_code = aws_api_gateway_method_response.verifications_get.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'",
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Api-Key,Authorization'"
  }

  depends_on = [aws_api_gateway_integration.verifications_get]
}

# 3. Verification ID - GET /api/verifications/{id}
resource "aws_api_gateway_integration_response" "verification_id_get" {
  count = var.cors_enabled ? 1 : 0

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_id.id
  http_method = aws_api_gateway_method.verification_id_get.http_method
  status_code = aws_api_gateway_method_response.verification_id_get.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'",
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Api-Key,Authorization'"
  }

  depends_on = [aws_api_gateway_integration.verification_id_get]
}

# 4. Verification Conversation - GET /api/verifications/{id}/conversation
resource "aws_api_gateway_integration_response" "verification_conversation_get" {
  count = var.cors_enabled ? 1 : 0

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_conversation.id
  http_method = aws_api_gateway_method.verification_conversation_get.http_method
  status_code = aws_api_gateway_method_response.verification_conversation_get.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'",
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Api-Key,Authorization'"
  }

  depends_on = [aws_api_gateway_integration.verification_conversation_get]
}

# 5. Image View - GET /api/images/{key}/view
resource "aws_api_gateway_integration_response" "image_view_get" {
  count = var.cors_enabled ? 1 : 0

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_view.id
  http_method = aws_api_gateway_method.image_view_get.http_method
  status_code = aws_api_gateway_method_response.image_view_get.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'",
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Api-Key,Authorization'"
  }

  depends_on = [aws_api_gateway_integration.image_view_get]
}

# 6. Image Browser - GET /api/images/browser
resource "aws_api_gateway_integration_response" "image_browser_get" {
  count = var.cors_enabled ? 1 : 0

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_browser.id
  http_method = aws_api_gateway_method.image_browser_get.http_method
  status_code = aws_api_gateway_method_response.image_browser_get.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'",
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Api-Key,Authorization'"
  }

  depends_on = [aws_api_gateway_integration.image_browser_get]
}

# 7. Verifications - POST /api/verifications
resource "aws_api_gateway_integration_response" "verifications_post" {
  count = var.cors_enabled ? 1 : 0

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_post.http_method
  status_code = aws_api_gateway_method_response.verifications_post.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'",
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Api-Key,Authorization'"
  }

  depends_on = [aws_api_gateway_integration.verifications_post]
}

# 9. Health Check - GET /api/health
resource "aws_api_gateway_integration_response" "health_get" {
  count = var.cors_enabled ? 1 : 0

  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.health.id
  http_method = aws_api_gateway_method.health_get.http_method
  status_code = aws_api_gateway_method_response.health_get.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'",
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Api-Key,Authorization'"
  }

  depends_on = [aws_api_gateway_integration.health_get]
}
