# Create REST API Gateway (v1)
resource "aws_api_gateway_rest_api" "this" {
  name        = var.api_name
  description = var.api_description

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = merge(
    var.common_tags,
    {
      Name = var.api_name
    }
  )
}

# Create root resource for /api
resource "aws_api_gateway_resource" "api" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_rest_api.this.root_resource_id
  path_part   = "api"
}

# Create resource for /api/{stage_name}
resource "aws_api_gateway_resource" "stage" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.api.id
  path_part   = var.stage_name
}

# Create resource for /api/{stage_name}/verifications
resource "aws_api_gateway_resource" "verifications" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.stage.id
  path_part   = "verifications"
}

# Create resource for /api/{stage_name}/verifications/lookup
resource "aws_api_gateway_resource" "verifications_lookup" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.verifications.id
  path_part   = "lookup"
}

# Create resource for /api/{stage_name}/verifications/{id}
resource "aws_api_gateway_resource" "verification_id" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.verifications.id
  path_part   = "{id}"
}

# Create resource for /api/{stage_name}/verifications/{id}/conversation
resource "aws_api_gateway_resource" "verification_conversation" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.verification_id.id
  path_part   = "conversation"
}

# Create resource for /api/{stage_name}/health
resource "aws_api_gateway_resource" "health" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.stage.id
  path_part   = "health"
}

# Create resource for /api/{stage_name}/images
resource "aws_api_gateway_resource" "images" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.stage.id
  path_part   = "images"
}

# Create resource for /api/{stage_name}/images/{key}
resource "aws_api_gateway_resource" "image_key" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.images.id
  path_part   = "{key}"
}

# Create resource for /api/{stage_name}/images/{key}/view
resource "aws_api_gateway_resource" "image_view" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.image_key.id
  path_part   = "view"
}

# Create resource for /api/{stage_name}/images/browser
resource "aws_api_gateway_resource" "image_browser" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.images.id
  path_part   = "browser"
}

# Create resource for /api/{stage_name}/images/browser/{path+}
resource "aws_api_gateway_resource" "image_browser_path" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  parent_id   = aws_api_gateway_resource.image_browser.id
  path_part   = "{path+}"
}

# POST method for /api/{stage_name}/verifications
resource "aws_api_gateway_method" "verifications_post" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.verifications.id
  http_method   = "POST"
  authorization = "NONE"
  api_key_required = var.use_api_key
}

# Integration for POST /api/{stage_name}/verifications
resource "aws_api_gateway_integration" "verifications_post" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.verifications.id
  http_method             = aws_api_gateway_method.verifications_post.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${lookup(var.lambda_function_arns, "initialize", null)}:$LATEST/invocations"
}

# GET method for /api/{stage_name}/verifications
resource "aws_api_gateway_method" "verifications_get" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.verifications.id
  http_method   = "GET"
  authorization = "NONE"
  api_key_required = var.use_api_key
}

# Integration for GET /api/{stage_name}/verifications
resource "aws_api_gateway_integration" "verifications_get" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.verifications.id
  http_method             = aws_api_gateway_method.verifications_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${lookup(var.lambda_function_arns, "initialize", null)}:$LATEST/invocations"
}

# GET method for /api/{stage_name}/verifications/lookup
resource "aws_api_gateway_method" "verification_lookup" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.verifications_lookup.id
  http_method   = "GET"
  authorization = "NONE"
  api_key_required = var.use_api_key
}

# Integration for GET /api/{stage_name}/verifications/lookup
resource "aws_api_gateway_integration" "verification_lookup" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.verifications_lookup.id
  http_method             = aws_api_gateway_method.verification_lookup.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${lookup(var.lambda_function_arns, "initialize", null)}:$LATEST/invocations"
}

# GET method for /api/{stage_name}/verifications/{id}
resource "aws_api_gateway_method" "verification_by_id" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.verification_id.id
  http_method   = "GET"
  authorization = "NONE"
  api_key_required = var.use_api_key

  request_parameters = {
    "method.request.path.id" = true
  }
}

# Integration for GET /api/{stage_name}/verifications/{id}
resource "aws_api_gateway_integration" "verification_by_id" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.verification_id.id
  http_method             = aws_api_gateway_method.verification_by_id.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${lookup(var.lambda_function_arns, "initialize", null)}:$LATEST/invocations"
}

# GET method for /api/{stage_name}/verifications/{id}/conversation
resource "aws_api_gateway_method" "verification_conversation" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.verification_conversation.id
  http_method   = "GET"
  authorization = "NONE"
  api_key_required = var.use_api_key

  request_parameters = {
    "method.request.path.id" = true
  }
}

# Integration for GET /api/{stage_name}/verifications/{id}/conversation
resource "aws_api_gateway_integration" "verification_conversation" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.verification_conversation.id
  http_method             = aws_api_gateway_method.verification_conversation.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${lookup(var.lambda_function_arns, "initialize", null)}:$LATEST/invocations"
}

# GET method for /api/{stage_name}/health
resource "aws_api_gateway_method" "health" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.health.id
  http_method   = "GET"
  authorization = "NONE"
}

# Integration for GET /api/{stage_name}/health
resource "aws_api_gateway_integration" "health" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.health.id
  http_method             = aws_api_gateway_method.health.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${lookup(var.lambda_function_arns, "initialize", null)}:$LATEST/invocations"
}

# GET method for /api/{stage_name}/images/{key}/view
resource "aws_api_gateway_method" "image_view" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.image_view.id
  http_method   = "GET"
  authorization = "NONE"
  api_key_required = var.use_api_key

  request_parameters = {
    "method.request.path.key" = true
  }
}

# Integration for GET /api/{stage_name}/images/{key}/view
resource "aws_api_gateway_integration" "image_view" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.image_view.id
  http_method             = aws_api_gateway_method.image_view.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${lookup(var.lambda_function_arns, "initialize", null)}:$LATEST/invocations"
}

# GET method for /api/{stage_name}/images/browser/{path+}
resource "aws_api_gateway_method" "image_browser" {
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.image_browser_path.id
  http_method   = "GET"
  authorization = "NONE"
  api_key_required = var.use_api_key

  request_parameters = {
    "method.request.path.path" = true
  }
}

# Integration for GET /api/{stage_name}/images/browser/{path+}
resource "aws_api_gateway_integration" "image_browser" {
  rest_api_id             = aws_api_gateway_rest_api.this.id
  resource_id             = aws_api_gateway_resource.image_browser_path.id
  http_method             = aws_api_gateway_method.image_browser.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${lookup(var.lambda_function_arns, "initialize", null)}:$LATEST/invocations"
}

# Define local variable for CORS resources
locals {
  cors_resources = var.cors_enabled ? {
    "verifications" = {
      resource_id = aws_api_gateway_resource.verifications.id
      path = "/api/${var.stage_name}/verifications"
    },
    "verifications_lookup" = {
      resource_id = aws_api_gateway_resource.verifications_lookup.id
      path = "/api/${var.stage_name}/verifications/lookup"
    },
    "verification_id" = {
      resource_id = aws_api_gateway_resource.verification_id.id
      path = "/api/${var.stage_name}/verifications/{id}"
    },
    "verification_conversation" = {
      resource_id = aws_api_gateway_resource.verification_conversation.id
      path = "/api/${var.stage_name}/verifications/{id}/conversation"
    },
    "health" = {
      resource_id = aws_api_gateway_resource.health.id
      path = "/api/${var.stage_name}/health"
    },
    "image_view" = {
      resource_id = aws_api_gateway_resource.image_view.id
      path = "/api/${var.stage_name}/images/{key}/view"
    },
    "image_browser_path" = {
      resource_id = aws_api_gateway_resource.image_browser_path.id
      path = "/api/${var.stage_name}/images/browser/{path+}"
    }
  } : {}
}

# Enable CORS for all resources if enabled - individual resources approach
# Create OPTIONS method for verifications
resource "aws_api_gateway_method" "verifications_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.verifications.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "verifications_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_options[0].http_method
  type        = "MOCK"
  
  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }
}

resource "aws_api_gateway_method_response" "verifications_options_200" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_options[0].http_method
  status_code = "200"
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true,
    "method.response.header.Access-Control-Allow-Methods" = true,
    "method.response.header.Access-Control-Allow-Origin"  = true,
    "method.response.header.Access-Control-Allow-Credentials" = true,
    "method.response.header.Access-Control-Max-Age" = true
  }
}

resource "aws_api_gateway_integration_response" "verifications_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_options[0].http_method
  status_code = aws_api_gateway_method_response.verifications_options_200[0].status_code
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
    "method.response.header.Access-Control-Allow-Methods" = "'DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT'",
    "method.response.header.Access-Control-Allow-Origin"  = var.cors_enabled ? join(",", [for origin in var.cors_allowed_origins : "'${origin}'"]) : "''",
    "method.response.header.Access-Control-Allow-Credentials" = "'true'",
    "method.response.header.Access-Control-Max-Age" = "'7200'"
  }
}

# Create OPTIONS method for verifications_lookup
resource "aws_api_gateway_method" "verifications_lookup_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.verifications_lookup.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "verifications_lookup_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verifications_lookup.id
  http_method = aws_api_gateway_method.verifications_lookup_options[0].http_method
  type        = "MOCK"
  
  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }
}

resource "aws_api_gateway_method_response" "verifications_lookup_options_200" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verifications_lookup.id
  http_method = aws_api_gateway_method.verifications_lookup_options[0].http_method
  status_code = "200"
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true,
    "method.response.header.Access-Control-Allow-Methods" = true,
    "method.response.header.Access-Control-Allow-Origin"  = true,
    "method.response.header.Access-Control-Allow-Credentials" = true,
    "method.response.header.Access-Control-Max-Age" = true
  }
}

resource "aws_api_gateway_integration_response" "verifications_lookup_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verifications_lookup.id
  http_method = aws_api_gateway_method.verifications_lookup_options[0].http_method
  status_code = aws_api_gateway_method_response.verifications_lookup_options_200[0].status_code
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
    "method.response.header.Access-Control-Allow-Methods" = "'DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT'",
    "method.response.header.Access-Control-Allow-Origin"  = var.cors_enabled ? join(",", [for origin in var.cors_allowed_origins : "'${origin}'"]) : "''",
    "method.response.header.Access-Control-Allow-Credentials" = "'true'",
    "method.response.header.Access-Control-Max-Age" = "'7200'"
  }
}

# Create OPTIONS method for verification_id
resource "aws_api_gateway_method" "verification_id_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.verification_id.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "verification_id_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verification_id.id
  http_method = aws_api_gateway_method.verification_id_options[0].http_method
  type        = "MOCK"
  
  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }
}

resource "aws_api_gateway_method_response" "verification_id_options_200" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verification_id.id
  http_method = aws_api_gateway_method.verification_id_options[0].http_method
  status_code = "200"
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true,
    "method.response.header.Access-Control-Allow-Methods" = true,
    "method.response.header.Access-Control-Allow-Origin"  = true,
    "method.response.header.Access-Control-Allow-Credentials" = true,
    "method.response.header.Access-Control-Max-Age" = true
  }
}

resource "aws_api_gateway_integration_response" "verification_id_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verification_id.id
  http_method = aws_api_gateway_method.verification_id_options[0].http_method
  status_code = aws_api_gateway_method_response.verification_id_options_200[0].status_code
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
    "method.response.header.Access-Control-Allow-Methods" = "'DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT'",
    "method.response.header.Access-Control-Allow-Origin"  = var.cors_enabled ? join(",", [for origin in var.cors_allowed_origins : "'${origin}'"]) : "''",
    "method.response.header.Access-Control-Allow-Credentials" = "'true'",
    "method.response.header.Access-Control-Max-Age" = "'7200'"
  }
}

# Create OPTIONS method for verification_conversation
resource "aws_api_gateway_method" "verification_conversation_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.verification_conversation.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "verification_conversation_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verification_conversation.id
  http_method = aws_api_gateway_method.verification_conversation_options[0].http_method
  type        = "MOCK"
  
  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }
}

resource "aws_api_gateway_method_response" "verification_conversation_options_200" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verification_conversation.id
  http_method = aws_api_gateway_method.verification_conversation_options[0].http_method
  status_code = "200"
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true,
    "method.response.header.Access-Control-Allow-Methods" = true,
    "method.response.header.Access-Control-Allow-Origin"  = true,
    "method.response.header.Access-Control-Allow-Credentials" = true,
    "method.response.header.Access-Control-Max-Age" = true
  }
}

resource "aws_api_gateway_integration_response" "verification_conversation_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.verification_conversation.id
  http_method = aws_api_gateway_method.verification_conversation_options[0].http_method
  status_code = aws_api_gateway_method_response.verification_conversation_options_200[0].status_code
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
    "method.response.header.Access-Control-Allow-Methods" = "'DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT'",
    "method.response.header.Access-Control-Allow-Origin"  = var.cors_enabled ? join(",", [for origin in var.cors_allowed_origins : "'${origin}'"]) : "''",
    "method.response.header.Access-Control-Allow-Credentials" = "'true'",
    "method.response.header.Access-Control-Max-Age" = "'7200'"
  }
}

# Create OPTIONS method for health
resource "aws_api_gateway_method" "health_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.health.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "health_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.health.id
  http_method = aws_api_gateway_method.health_options[0].http_method
  type        = "MOCK"
  
  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }
}

resource "aws_api_gateway_method_response" "health_options_200" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.health.id
  http_method = aws_api_gateway_method.health_options[0].http_method
  status_code = "200"
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true,
    "method.response.header.Access-Control-Allow-Methods" = true,
    "method.response.header.Access-Control-Allow-Origin"  = true,
    "method.response.header.Access-Control-Allow-Credentials" = true,
    "method.response.header.Access-Control-Max-Age" = true
  }
}

resource "aws_api_gateway_integration_response" "health_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.health.id
  http_method = aws_api_gateway_method.health_options[0].http_method
  status_code = aws_api_gateway_method_response.health_options_200[0].status_code
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
    "method.response.header.Access-Control-Allow-Methods" = "'DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT'",
    "method.response.header.Access-Control-Allow-Origin"  = var.cors_enabled ? join(",", [for origin in var.cors_allowed_origins : "'${origin}'"]) : "''",
    "method.response.header.Access-Control-Allow-Credentials" = "'true'",
    "method.response.header.Access-Control-Max-Age" = "'7200'"
  }
}

# Create OPTIONS method for image_view
resource "aws_api_gateway_method" "image_view_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.image_view.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "image_view_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.image_view.id
  http_method = aws_api_gateway_method.image_view_options[0].http_method
  type        = "MOCK"
  
  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }
}

resource "aws_api_gateway_method_response" "image_view_options_200" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.image_view.id
  http_method = aws_api_gateway_method.image_view_options[0].http_method
  status_code = "200"
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true,
    "method.response.header.Access-Control-Allow-Methods" = true,
    "method.response.header.Access-Control-Allow-Origin"  = true,
    "method.response.header.Access-Control-Allow-Credentials" = true,
    "method.response.header.Access-Control-Max-Age" = true
  }
}

resource "aws_api_gateway_integration_response" "image_view_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.image_view.id
  http_method = aws_api_gateway_method.image_view_options[0].http_method
  status_code = aws_api_gateway_method_response.image_view_options_200[0].status_code
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
    "method.response.header.Access-Control-Allow-Methods" = "'DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT'",
    "method.response.header.Access-Control-Allow-Origin"  = var.cors_enabled ? join(",", [for origin in var.cors_allowed_origins : "'${origin}'"]) : "''",
    "method.response.header.Access-Control-Allow-Credentials" = "'true'",
    "method.response.header.Access-Control-Max-Age" = "'7200'"
  }
}

# Create OPTIONS method for image_browser_path
resource "aws_api_gateway_method" "image_browser_path_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id   = aws_api_gateway_rest_api.this.id
  resource_id   = aws_api_gateway_resource.image_browser_path.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "image_browser_path_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.image_browser_path.id
  http_method = aws_api_gateway_method.image_browser_path_options[0].http_method
  type        = "MOCK"
  
  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }
}

resource "aws_api_gateway_method_response" "image_browser_path_options_200" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.image_browser_path.id
  http_method = aws_api_gateway_method.image_browser_path_options[0].http_method
  status_code = "200"
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true,
    "method.response.header.Access-Control-Allow-Methods" = true,
    "method.response.header.Access-Control-Allow-Origin"  = true,
    "method.response.header.Access-Control-Allow-Credentials" = true,
    "method.response.header.Access-Control-Max-Age" = true
  }
}

resource "aws_api_gateway_integration_response" "image_browser_path_options" {
  count = var.cors_enabled ? 1 : 0
  
  rest_api_id = aws_api_gateway_rest_api.this.id
  resource_id = aws_api_gateway_resource.image_browser_path.id
  http_method = aws_api_gateway_method.image_browser_path_options[0].http_method
  status_code = aws_api_gateway_method_response.image_browser_path_options_200[0].status_code
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
    "method.response.header.Access-Control-Allow-Methods" = "'DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT'",
    "method.response.header.Access-Control-Allow-Origin"  = var.cors_enabled ? join(",", [for origin in var.cors_allowed_origins : "'${origin}'"]) : "''",
    "method.response.header.Access-Control-Allow-Credentials" = "'true'",
    "method.response.header.Access-Control-Max-Age" = "'7200'"
  }
}

# Create API Gateway deployment
resource "aws_api_gateway_deployment" "this" {
  rest_api_id = aws_api_gateway_rest_api.this.id
  
  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.api.id,
      aws_api_gateway_resource.stage.id,
      aws_api_gateway_resource.verifications.id,
      aws_api_gateway_resource.verifications_lookup.id,
      aws_api_gateway_resource.verification_id.id,
      aws_api_gateway_resource.verification_conversation.id,
      aws_api_gateway_resource.health.id,
      aws_api_gateway_resource.images.id,
      aws_api_gateway_resource.image_key.id,
      aws_api_gateway_resource.image_view.id,
      aws_api_gateway_resource.image_browser.id,
      aws_api_gateway_resource.image_browser_path.id,
      aws_api_gateway_method.verifications_post.id,
      aws_api_gateway_integration.verifications_post.id,
      aws_api_gateway_method.verifications_get.id,
      aws_api_gateway_integration.verifications_get.id,
      aws_api_gateway_method.verification_lookup.id,
      aws_api_gateway_integration.verification_lookup.id,
      aws_api_gateway_method.verification_by_id.id,
      aws_api_gateway_integration.verification_by_id.id,
      aws_api_gateway_method.verification_conversation.id,
      aws_api_gateway_integration.verification_conversation.id,
      aws_api_gateway_method.health.id,
      aws_api_gateway_integration.health.id,
      aws_api_gateway_method.image_view.id,
      aws_api_gateway_integration.image_view.id,
      aws_api_gateway_method.image_browser.id,
      aws_api_gateway_integration.image_browser.id
    ]))
  }
  
  lifecycle {
    create_before_destroy = true
  }
}

# Create API Gateway stage
resource "aws_api_gateway_stage" "this" {
  deployment_id = aws_api_gateway_deployment.this.id
  rest_api_id   = aws_api_gateway_rest_api.this.id
  stage_name    = var.stage_name
  
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId               = "$context.requestId"
      sourceIp                = "$context.identity.sourceIp"
      requestTime             = "$context.requestTime"
      protocol                = "$context.protocol"
      httpMethod              = "$context.httpMethod"
      resourcePath            = "$context.resourcePath"
      routeKey                = "$context.routeKey"
      status                  = "$context.status"
      responseLength          = "$context.responseLength"
      integrationErrorMessage = "$context.integrationErrorMessage"
    })
  }
  
  variables = {
    "throttlingRateLimit"  = var.throttling_rate_limit
    "throttlingBurstLimit" = var.throttling_burst_limit
    "metricsEnabled"       = var.metrics_enabled ? "true" : "false"
  }
}

# CloudWatch log group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.api_name}"
  retention_in_days = var.log_retention_days
}

# API Key (optional)
resource "aws_api_gateway_api_key" "this" {
  count = var.use_api_key ? 1 : 0
  name  = "${var.api_name}-key"
  enabled = true
  description = "API key for ${var.api_name}"
}

# Create usage plan if API key is used
resource "aws_api_gateway_usage_plan" "this" {
  count = var.use_api_key ? 1 : 0
  
  name        = "${var.api_name}-usage-plan"
  description = "Usage plan for ${var.api_name}"
  
  api_stages {
    api_id = aws_api_gateway_rest_api.this.id
    stage  = aws_api_gateway_stage.this.stage_name
  }
  
  quota_settings {
    limit  = var.api_quota_limit
    period = var.api_quota_period
  }
  
  throttle_settings {
    burst_limit = var.throttling_burst_limit
    rate_limit  = var.throttling_rate_limit
  }
}

# Associate API key with usage plan
resource "aws_api_gateway_usage_plan_key" "this" {
  count = var.use_api_key ? 1 : 0
  
  key_id        = aws_api_gateway_api_key.this[0].id
  key_type      = "API_KEY"
  usage_plan_id = aws_api_gateway_usage_plan.this[0].id
}

# Permission for API Gateway to invoke Lambda
resource "aws_lambda_permission" "api_gateway" {
  for_each = var.lambda_function_names

  statement_id  = "AllowExecutionFromAPIGateway-${each.key}"
  action        = "lambda:InvokeFunction"
  function_name = each.value
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.this.execution_arn}/*/*/*"
}
