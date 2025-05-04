# modules/api_gateway/error_responses.tf

# Custom Gateway Responses for API errors
resource "aws_api_gateway_gateway_response" "bad_request_body" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  response_type = "BAD_REQUEST_BODY"
  status_code   = "400"
  
  response_templates = {
    "application/json" = jsonencode({
      error = "BadRequest",
      message = "Invalid request body",
      details = {
        requestId = "$context.requestId"
      }
    })
  }
  
  response_parameters = {
    "gatewayresponse.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
    "gatewayresponse.header.Content-Type" = "'application/json'"
  }
}

resource "aws_api_gateway_gateway_response" "unauthorized" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  response_type = "UNAUTHORIZED"
  status_code   = "401"
  
  response_templates = {
    "application/json" = jsonencode({
      error = "Unauthorized",
      message = "Missing or invalid authentication token",
      details = {
        requestId = "$context.requestId"
      }
    })
  }
  
  response_parameters = {
    "gatewayresponse.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
    "gatewayresponse.header.Content-Type" = "'application/json'"
  }
}

resource "aws_api_gateway_gateway_response" "access_denied" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  response_type = "ACCESS_DENIED"
  status_code   = "403"
  
  response_templates = {
    "application/json" = jsonencode({
      error = "AccessDenied",
      message = "You don't have permission to access this resource",
      details = {
        requestId = "$context.requestId"
      }
    })
  }
  
  response_parameters = {
    "gatewayresponse.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
    "gatewayresponse.header.Content-Type" = "'application/json'"
  }
}

resource "aws_api_gateway_gateway_response" "resource_not_found" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  response_type = "RESOURCE_NOT_FOUND"
  status_code   = "404"
  
  response_templates = {
    "application/json" = jsonencode({
      error = "ResourceNotFound",
      message = "The requested resource does not exist",
      details = {
        requestId = "$context.requestId"
      }
    })
  }
  
  response_parameters = {
    "gatewayresponse.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
    "gatewayresponse.header.Content-Type" = "'application/json'"
  }
}

resource "aws_api_gateway_gateway_response" "method_not_allowed" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  response_type = "METHOD_NOT_ALLOWED"
  status_code   = "405"
  
  response_templates = {
    "application/json" = jsonencode({
      error = "MethodNotAllowed",
      message = "The requested method is not allowed for this resource",
      details = {
        requestId = "$context.requestId"
      }
    })
  }
  
  response_parameters = {
    "gatewayresponse.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
    "gatewayresponse.header.Content-Type" = "'application/json'"
  }
}

resource "aws_api_gateway_gateway_response" "integration_timeout" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  response_type = "INTEGRATION_TIMEOUT"
  status_code   = "504"
  
  response_templates = {
    "application/json" = jsonencode({
      error = "IntegrationTimeout",
      message = "The backend service did not respond in time",
      details = {
        requestId = "$context.requestId"
      }
    })
  }
  
  response_parameters = {
    "gatewayresponse.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
    "gatewayresponse.header.Content-Type" = "'application/json'"
  }
}

resource "aws_api_gateway_gateway_response" "default_5xx" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  response_type = "DEFAULT_5XX"
  status_code   = "500"
  
  response_templates = {
    "application/json" = jsonencode({
      error = "ServerError",
      message = "An unexpected error occurred",
      details = {
        requestId = "$context.requestId"
      }
    })
  }
  
  response_parameters = {
    "gatewayresponse.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
    "gatewayresponse.header.Content-Type" = "'application/json'"
  }
}