# infrastructure/modules/api_gateway/main.tf
resource "aws_api_gateway_rest_api" "verification_api" {
  name        = var.api_name
  description = "API for vending machine verification"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = merge(
    {
      Name        = var.api_name
      Environment = var.environment
    },
    var.tags
  )
}

# API Resources
resource "aws_api_gateway_resource" "comparisons" {
  rest_api_id = aws_api_gateway_rest_api.verification_api.id
  parent_id   = aws_api_gateway_rest_api.verification_api.root_resource_id
  path_part   = "comparisons"
}

resource "aws_api_gateway_resource" "comparison_id" {
  rest_api_id = aws_api_gateway_rest_api.verification_api.id
  parent_id   = aws_api_gateway_resource.comparisons.id
  path_part   = "{id}"
}

resource "aws_api_gateway_resource" "images" {
  rest_api_id = aws_api_gateway_rest_api.verification_api.id
  parent_id   = aws_api_gateway_rest_api.verification_api.root_resource_id
  path_part   = "images"
}

# POST /comparisons
resource "aws_api_gateway_method" "post_comparisons" {
  rest_api_id   = aws_api_gateway_rest_api.verification_api.id
  resource_id   = aws_api_gateway_resource.comparisons.id
  http_method   = "POST"
  authorization = "NONE"

  request_validator_id = aws_api_gateway_request_validator.validator.id
  request_models = {
    "application/json" = aws_api_gateway_model.comparison_request.name
  }
}

resource "aws_api_gateway_integration" "post_comparisons_integration" {
  rest_api_id             = aws_api_gateway_rest_api.verification_api.id
  resource_id             = aws_api_gateway_resource.comparisons.id
  http_method             = aws_api_gateway_method.post_comparisons.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = var.step_functions_invoke_arn
  credentials             = aws_iam_role.api_gateway_step_functions_role.arn
  
  request_templates = {
    "application/json" = jsonencode({
      input = "$util.escapeJavaScript($input.json('$'))"
      stateMachineArn = var.step_functions_state_machine_arn
    })
  }
}

resource "aws_api_gateway_method_response" "post_comparisons_response_200" {
  rest_api_id = aws_api_gateway_rest_api.verification_api.id
  resource_id = aws_api_gateway_resource.comparisons.id
  http_method = aws_api_gateway_method.post_comparisons.http_method
  status_code = "200"
  
  response_models = {
    "application/json" = "Empty"
  }
}

# API Gateway Integration Response fixes
# The integration response resource should depend on the integration resource
# Add this to your modules/api_gateway/main.tf file

# Fix for the API Gateway Integration Response in modules/api_gateway/main.tf
# Replace the current post_comparisons_integration_response resource with this corrected version:

resource "aws_api_gateway_integration_response" "post_comparisons_integration_response" {
  count = var.skip_api_gateway_integration_response ? 0 : 1
  
  rest_api_id         = aws_api_gateway_rest_api.verification_api.id
  resource_id         = aws_api_gateway_resource.comparisons.id
  http_method         = aws_api_gateway_method.post_comparisons.http_method
  status_code         = aws_api_gateway_method_response.post_comparisons_response_200.status_code
  selection_pattern   = "200"
  
  response_templates = {
    "application/json" = jsonencode({
      executionArn = "$input.path('$.executionArn')"
      startDate    = "$input.path('$.startDate')"
    })
  }
  
  depends_on = [
    aws_api_gateway_integration.post_comparisons_integration
  ]
}

# GET /comparisons/{id}
resource "aws_api_gateway_method" "get_comparison" {
  rest_api_id   = aws_api_gateway_rest_api.verification_api.id
  resource_id   = aws_api_gateway_resource.comparison_id.id
  http_method   = "GET"
  authorization = "NONE"
  
  request_parameters = {
    "method.request.path.id" = true
  }
}

resource "aws_api_gateway_integration" "get_comparison_integration" {
  rest_api_id             = aws_api_gateway_rest_api.verification_api.id
  resource_id             = aws_api_gateway_resource.comparison_id.id
  http_method             = aws_api_gateway_method.get_comparison.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = var.get_comparison_lambda_invoke_arn
  
  request_templates = {
    "application/json" = jsonencode({
      id = "$input.params('id')"
    })
  }
}

resource "aws_api_gateway_method_response" "get_comparison_response_200" {
  rest_api_id = aws_api_gateway_rest_api.verification_api.id
  resource_id = aws_api_gateway_resource.comparison_id.id
  http_method = aws_api_gateway_method.get_comparison.http_method
  status_code = "200"
  
  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "get_comparison_integration_response" {
  rest_api_id = aws_api_gateway_rest_api.verification_api.id
  resource_id = aws_api_gateway_resource.comparison_id.id
  http_method = aws_api_gateway_method.get_comparison.http_method
  status_code = aws_api_gateway_method_response.get_comparison_response_200.status_code
  
  response_templates = {
    "application/json" = ""
  }
}

# GET /images
resource "aws_api_gateway_method" "get_images" {
  rest_api_id   = aws_api_gateway_rest_api.verification_api.id
  resource_id   = aws_api_gateway_resource.images.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "get_images_integration" {
  rest_api_id             = aws_api_gateway_rest_api.verification_api.id
  resource_id             = aws_api_gateway_resource.images.id
  http_method             = aws_api_gateway_method.get_images.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = var.get_images_lambda_invoke_arn
  
  request_templates = {
    "application/json" = jsonencode({
      machineId = "$input.params('machineId')",
      type = "$input.params('type')"
    })
  }
}

resource "aws_api_gateway_method_response" "get_images_response_200" {
  rest_api_id = aws_api_gateway_rest_api.verification_api.id
  resource_id = aws_api_gateway_resource.images.id
  http_method = aws_api_gateway_method.get_images.http_method
  status_code = "200"
  
  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_integration_response" "get_images_integration_response" {
  rest_api_id = aws_api_gateway_rest_api.verification_api.id
  resource_id = aws_api_gateway_resource.images.id
  http_method = aws_api_gateway_method.get_images.http_method
  status_code = aws_api_gateway_method_response.get_images_response_200.status_code
  
  response_templates = {
    "application/json" = ""
  }
}

# Request Validator
resource "aws_api_gateway_request_validator" "validator" {
  name                        = "validator"
  rest_api_id                 = aws_api_gateway_rest_api.verification_api.id
  validate_request_body       = true
  validate_request_parameters = true
}

# Models
resource "aws_api_gateway_model" "comparison_request" {
  rest_api_id  = aws_api_gateway_rest_api.verification_api.id
  name         = "ComparisonRequest"
  description  = "JSON schema for comparison request"
  content_type = "application/json"

  schema = jsonencode({
    "$schema" = "http://json-schema.org/draft-04/schema#",
    "type" = "object",
    "required" = ["referenceImageKey", "checkingImageKey", "vendingMachineId"],
    "properties" = {
      "referenceImageKey" = {
        "type" = "string"
      },
      "checkingImageKey" = {
        "type" = "string"
      },
      "vendingMachineId" = {
        "type" = "string"
      },
      "location" = {
        "type" = "string"
      }
    }
  })
}

# API Gateway Stage for deployment
resource "aws_api_gateway_stage" "api_stage" {
  deployment_id = aws_api_gateway_deployment.deployment.id
  rest_api_id   = aws_api_gateway_rest_api.verification_api.id
  stage_name    = var.stage_name
  
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway_logs.arn
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
  
  tags = merge(
    {
      Name        = "${var.api_name}-${var.stage_name}"
      Environment = var.environment
    },
    var.tags
  )
}

# CloudWatch Log Group for API Gateway logs
resource "aws_cloudwatch_log_group" "api_gateway_logs" {
  name              = "API-Gateway-Execution-Logs_${aws_api_gateway_rest_api.verification_api.id}/${var.stage_name}"
  retention_in_days = 30
  
  tags = merge(
    {
      Name        = "API-Gateway-Logs-${var.api_name}"
      Environment = var.environment
    },
    var.tags
  )
}

# API Deployment
resource "aws_api_gateway_deployment" "deployment" {
  depends_on = [
    aws_api_gateway_integration.post_comparisons_integration,
    aws_api_gateway_integration.get_comparison_integration,
    aws_api_gateway_integration.get_images_integration
  ]

  rest_api_id = aws_api_gateway_rest_api.verification_api.id
  
  lifecycle {
    create_before_destroy = true
  }
}

# IAM Role for API Gateway to invoke Step Functions
resource "aws_iam_role" "api_gateway_step_functions_role" {
  name = "api-gateway-step-functions-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "apigateway.amazonaws.com"
        }
      }
    ]
  })
  
  tags = merge(
    {
      Environment = var.environment
    },
    var.tags
  )
}

resource "aws_iam_policy" "api_gateway_step_functions_policy" {
  name = "api-gateway-step-functions-policy"
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "states:StartExecution"
        ]
        Resource = [
          var.step_functions_state_machine_arn
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "api_gateway_step_functions_policy_attachment" {
  role       = aws_iam_role.api_gateway_step_functions_role.name
  policy_arn = aws_iam_policy.api_gateway_step_functions_policy.arn
}

# Lambda Permissions
resource "aws_lambda_permission" "get_comparison_lambda_permission" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = var.get_comparison_lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.verification_api.execution_arn}/*/${aws_api_gateway_method.get_comparison.http_method}${aws_api_gateway_resource.comparison_id.path}"
}

resource "aws_lambda_permission" "get_images_lambda_permission" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = var.get_images_lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.verification_api.execution_arn}/*/${aws_api_gateway_method.get_images.http_method}${aws_api_gateway_resource.images.path}"
}