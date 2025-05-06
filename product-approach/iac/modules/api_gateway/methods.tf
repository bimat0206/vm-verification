# modules/api_gateway/methods.tf

# 1. Verification Lookup - GET /api/verifications/lookup
resource "aws_api_gateway_method" "verifications_lookup_get" {
  rest_api_id      = aws_api_gateway_rest_api.api.id
  resource_id      = aws_api_gateway_resource.verifications_lookup.id
  http_method      = "GET"
  authorization    = var.use_api_key ? "NONE" : "NONE"
  api_key_required = var.use_api_key
}

resource "aws_api_gateway_method_response" "verifications_lookup_get" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications_lookup.id
  http_method = aws_api_gateway_method.verifications_lookup_get.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

# OPTIONS method for CORS
resource "aws_api_gateway_method" "verifications_lookup_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.verifications_lookup.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "verifications_lookup_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications_lookup.id
  http_method = aws_api_gateway_method.verifications_lookup_options.http_method
  type        = "MOCK"
  request_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }
}

resource "aws_api_gateway_method_response" "verifications_lookup_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications_lookup.id
  http_method = aws_api_gateway_method.verifications_lookup_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "verifications_lookup_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications_lookup.id
  http_method = aws_api_gateway_method.verifications_lookup_options.http_method
  status_code = aws_api_gateway_method_response.verifications_lookup_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'*'"
    "method.response.header.Access-Control-Allow-Methods" = "'*'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'"
  }
}

resource "aws_api_gateway_integration" "verifications_lookup_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.verifications_lookup.id
  http_method             = aws_api_gateway_method.verifications_lookup_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["fetch_historical_verification"]}/invocations"
}

# 2. Initiate Verification - POST /api/verifications
resource "aws_api_gateway_method" "verifications_post" {
  rest_api_id      = aws_api_gateway_rest_api.api.id
  resource_id      = aws_api_gateway_resource.verifications.id
  http_method      = "POST"
  authorization    = var.use_api_key ? "NONE" : "NONE"
  api_key_required = var.use_api_key

  # Add request validation
  request_validator_id = aws_api_gateway_request_validator.full_validator.id
  request_models = {
    "application/json" = aws_api_gateway_model.verification_request.name
  }
}

resource "aws_api_gateway_method_response" "verifications_post" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_post.http_method
  status_code = "202"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Location"                     = true
  }

  response_models = {
    "application/json" = aws_api_gateway_model.verification_result.name
  }
}

resource "aws_api_gateway_method_response" "verifications_post_400" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_post.http_method
  status_code = "400"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = aws_api_gateway_model.error.name
  }
}
resource "aws_api_gateway_method_response" "verifications_post_404" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_post.http_method
  status_code = "404"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = aws_api_gateway_model.error.name
  }
}

# OPTIONS method for verifications
resource "aws_api_gateway_method" "verifications_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.verifications.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "verifications_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_options.http_method
  type        = "MOCK"
  request_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }
}

resource "aws_api_gateway_method_response" "verifications_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "verifications_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_options.http_method
  status_code = aws_api_gateway_method_response.verifications_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'*'"
    "method.response.header.Access-Control-Allow-Methods" = "'*'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'"
  }
}

resource "aws_api_gateway_integration" "verifications_post" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.verifications.id
  http_method             = aws_api_gateway_method.verifications_post.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["initialize"]}/invocations"
  passthrough_behavior    = "WHEN_NO_TEMPLATES"

  # Add request templates for non-proxy integration
  request_templates = {
    "application/json" = <<EOF
{
  "body": $input.json('$'),
  "headers": {
    #foreach($header in $input.params().header.keySet())
    "$header": "$util.escapeJavaScript($input.params().header.get($header))" #if($foreach.hasNext),#end
    #end
  },
  "method": "$context.httpMethod",
  "params": {
    #foreach($param in $input.params().path.keySet())
    "$param": "$util.escapeJavaScript($input.params().path.get($param))" #if($foreach.hasNext),#end
    #end
  },
  "query": {
    #foreach($queryParam in $input.params().querystring.keySet())
    "$queryParam": "$util.escapeJavaScript($input.params().querystring.get($queryParam))" #if($foreach.hasNext),#end
    #end
  }
}
EOF
  }
}

# 3. List Verifications - GET /api/verifications
resource "aws_api_gateway_method" "verifications_get" {
  rest_api_id      = aws_api_gateway_rest_api.api.id
  resource_id      = aws_api_gateway_resource.verifications.id
  http_method      = "GET"
  authorization    = var.use_api_key ? "NONE" : "NONE"
  api_key_required = var.use_api_key

  # Add request parameter validation
  request_validator_id = aws_api_gateway_request_validator.params_only_validator.id
  request_parameters = {
    "method.request.querystring.vendingMachineId"   = false
    "method.request.querystring.verificationStatus" = false
    "method.request.querystring.fromDate"           = false
    "method.request.querystring.toDate"             = false
    "method.request.querystring.limit"              = false
    "method.request.querystring.offset"             = false
    "method.request.querystring.sortBy"             = false
  }
}

resource "aws_api_gateway_method_response" "verifications_get" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_get.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = aws_api_gateway_model.verification_list.name
  }
}

resource "aws_api_gateway_integration" "verifications_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.verifications.id
  http_method             = aws_api_gateway_method.verifications_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["list_verifications"]}/invocations"
}

# 4. Get Verification - GET /api/verifications/{verificationId}
resource "aws_api_gateway_method" "verification_id_get" {
  rest_api_id      = aws_api_gateway_rest_api.api.id
  resource_id      = aws_api_gateway_resource.verification_id.id
  http_method      = "GET"
  authorization    = var.use_api_key ? "NONE" : "NONE"
  api_key_required = var.use_api_key
  request_parameters = {
    "method.request.path.verificationId" = true
  }
}

# OPTIONS method for verification_id
resource "aws_api_gateway_method" "verification_id_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.verification_id.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "verification_id_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_id.id
  http_method = aws_api_gateway_method.verification_id_options.http_method
  type        = "MOCK"
  request_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }
}

resource "aws_api_gateway_method_response" "verification_id_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_id.id
  http_method = aws_api_gateway_method.verification_id_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

# Integration responses for POST /api/verifications
resource "aws_api_gateway_integration_response" "verifications_post_success" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_post.http_method
  status_code = aws_api_gateway_method_response.verifications_post.status_code

  # Default response (no selection pattern means this is the default)
  selection_pattern = ""

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
    "method.response.header.Location"                    = "integration.response.body.headers.Location"
  }

  # Transform the Lambda response to match expected format
  response_templates = {
    "application/json" = <<EOF
#set($inputRoot = $input.path('$'))
$input.json('$.body')
EOF
  }

  depends_on = [aws_api_gateway_integration.verifications_post]
}

resource "aws_api_gateway_integration_response" "verifications_post_bad_request" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_post.http_method
  status_code = aws_api_gateway_method_response.verifications_post_400.status_code

  # Match 400 error responses from Lambda
  selection_pattern = ".*[Bb]ad [Rr]equest.*|.*400.*"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
  }

  # Transform the error response
  response_templates = {
    "application/json" = <<EOF
#set($inputRoot = $input.path('$'))
{
  "message": "Bad Request",
  "details": $input.json('$.errorMessage')
}
EOF
  }

  depends_on = [aws_api_gateway_integration.verifications_post]
}

resource "aws_api_gateway_integration_response" "verifications_post_not_found" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verifications.id
  http_method = aws_api_gateway_method.verifications_post.http_method
  status_code = aws_api_gateway_method_response.verifications_post_404.status_code

  # Match 404 error responses from Lambda
  selection_pattern = ".*[Nn]ot [Ff]ound.*|.*404.*"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'${local.cors_origin}'"
  }

  # Transform the error response
  response_templates = {
    "application/json" = <<EOF
#set($inputRoot = $input.path('$'))
{
  "message": "Not Found",
  "details": $input.json('$.errorMessage')
}
EOF
  }

  depends_on = [aws_api_gateway_integration.verifications_post]
}
resource "aws_api_gateway_integration_response" "verification_id_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_id.id
  http_method = aws_api_gateway_method.verification_id_options.http_method
  status_code = aws_api_gateway_method_response.verification_id_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'*'"
    "method.response.header.Access-Control-Allow-Methods" = "'*'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'"
  }
}

resource "aws_api_gateway_integration" "verification_id_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.verification_id.id
  http_method             = aws_api_gateway_method.verification_id_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["get_verification"]}/invocations"
}

resource "aws_api_gateway_method_response" "verification_id_get" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_id.id
  http_method = aws_api_gateway_method.verification_id_get.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

# 5. Get Verification Conversation - GET /api/verifications/{verificationId}/conversation
resource "aws_api_gateway_method" "verification_conversation_get" {
  rest_api_id      = aws_api_gateway_rest_api.api.id
  resource_id      = aws_api_gateway_resource.verification_conversation.id
  http_method      = "GET"
  authorization    = var.use_api_key ? "NONE" : "NONE"
  api_key_required = var.use_api_key
  request_parameters = {
    "method.request.path.verificationId" = true
  }
}

# OPTIONS method for verification_conversation
resource "aws_api_gateway_method" "verification_conversation_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.verification_conversation.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "verification_conversation_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_conversation.id
  http_method = aws_api_gateway_method.verification_conversation_options.http_method
  type        = "MOCK"
  request_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }
}

resource "aws_api_gateway_method_response" "verification_conversation_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_conversation.id
  http_method = aws_api_gateway_method.verification_conversation_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "verification_conversation_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_conversation.id
  http_method = aws_api_gateway_method.verification_conversation_options.http_method
  status_code = aws_api_gateway_method_response.verification_conversation_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'*'"
    "method.response.header.Access-Control-Allow-Methods" = "'*'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'"
  }
}

resource "aws_api_gateway_integration" "verification_conversation_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.verification_conversation.id
  http_method             = aws_api_gateway_method.verification_conversation_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["get_conversation"]}/invocations"
}

resource "aws_api_gateway_method_response" "verification_conversation_get" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.verification_conversation.id
  http_method = aws_api_gateway_method.verification_conversation_get.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

# 6. Health Check - GET /api/health
resource "aws_api_gateway_method" "health_get" {
  rest_api_id      = aws_api_gateway_rest_api.api.id
  resource_id      = aws_api_gateway_resource.health.id
  http_method      = "GET"
  authorization    = var.use_api_key ? "NONE" : "NONE"
  api_key_required = var.use_api_key
}

# OPTIONS method for health
resource "aws_api_gateway_method" "health_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.health.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "health_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.health.id
  http_method = aws_api_gateway_method.health_options.http_method
  type        = "MOCK"
  request_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }
}

resource "aws_api_gateway_method_response" "health_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.health.id
  http_method = aws_api_gateway_method.health_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "health_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.health.id
  http_method = aws_api_gateway_method.health_options.http_method
  status_code = aws_api_gateway_method_response.health_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'*'"
    "method.response.header.Access-Control-Allow-Methods" = "'*'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'"
  }
}

resource "aws_api_gateway_integration" "health_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.health.id
  http_method             = aws_api_gateway_method.health_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["health_check"]}/invocations"
}

resource "aws_api_gateway_method_response" "health_get" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.health.id
  http_method = aws_api_gateway_method.health_get.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

# Health GET integration response is now defined in cors_integration_responses.tf

# 7. Get Image View - GET /api/images/{key}/view
resource "aws_api_gateway_method" "image_view_get" {
  rest_api_id      = aws_api_gateway_rest_api.api.id
  resource_id      = aws_api_gateway_resource.image_view.id
  http_method      = "GET"
  authorization    = var.use_api_key ? "NONE" : "NONE"
  api_key_required = var.use_api_key
  request_parameters = {
    "method.request.path.key" = true
  }
}

# OPTIONS method for image_view
resource "aws_api_gateway_method" "image_view_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.image_view.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "image_view_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_view.id
  http_method = aws_api_gateway_method.image_view_options.http_method
  type        = "MOCK"
  request_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }
}

resource "aws_api_gateway_method_response" "image_view_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_view.id
  http_method = aws_api_gateway_method.image_view_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "image_view_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_view.id
  http_method = aws_api_gateway_method.image_view_options.http_method
  status_code = aws_api_gateway_method_response.image_view_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'*'"
    "method.response.header.Access-Control-Allow-Methods" = "'*'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'"
  }
}

resource "aws_api_gateway_integration" "image_view_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.image_view.id
  http_method             = aws_api_gateway_method.image_view_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["fetch_images"]}/invocations"
}

resource "aws_api_gateway_method_response" "image_view_get" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_view.id
  http_method = aws_api_gateway_method.image_view_get.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

# 8. Get Image Browser - GET /api/images/browser
resource "aws_api_gateway_method" "image_browser_get" {
  rest_api_id      = aws_api_gateway_rest_api.api.id
  resource_id      = aws_api_gateway_resource.image_browser.id
  http_method      = "GET"
  authorization    = var.use_api_key ? "NONE" : "NONE"
  api_key_required = var.use_api_key
}

# OPTIONS method for image_browser
resource "aws_api_gateway_method" "image_browser_options" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.image_browser.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "image_browser_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_browser.id
  http_method = aws_api_gateway_method.image_browser_options.http_method
  type        = "MOCK"
  request_templates = {
    "application/json" = "{ \"statusCode\": 200 }"
  }
}

resource "aws_api_gateway_method_response" "image_browser_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_browser.id
  http_method = aws_api_gateway_method.image_browser_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "image_browser_options" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_browser.id
  http_method = aws_api_gateway_method.image_browser_options.http_method
  status_code = aws_api_gateway_method_response.image_browser_options.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'*'"
    "method.response.header.Access-Control-Allow-Methods" = "'*'"
    "method.response.header.Access-Control-Allow-Origin"  = "'${local.cors_origin}'"
  }
}

resource "aws_api_gateway_integration" "image_browser_get" {
  rest_api_id             = aws_api_gateway_rest_api.api.id
  resource_id             = aws_api_gateway_resource.image_browser.id
  http_method             = aws_api_gateway_method.image_browser_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_function_arns["fetch_images"]}/invocations"
}

resource "aws_api_gateway_method_response" "image_browser_get" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.image_browser.id
  http_method = aws_api_gateway_method.image_browser_get.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin"  = true
    "method.response.header.Access-Control-Allow-Headers" = true
  }

  response_models = {
    "application/json" = "Empty"
  }
}

# Lambda permissions for API Gateway to invoke Lambda functions
resource "aws_lambda_permission" "api_gateway_lambda" {
  for_each = var.lambda_function_arns

  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = each.value
  principal     = "apigateway.amazonaws.com"

  # Allow invocation from any method/resource in this API Gateway
  source_arn = "${aws_api_gateway_rest_api.api.execution_arn}/*/*/*"
}
