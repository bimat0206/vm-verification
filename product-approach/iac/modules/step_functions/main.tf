
# Create the state machine definition template file
resource "local_file" "state_machine_definition" {
  count = var.create_definition_file ? 1 : 0
  
  content  = templatefile("${path.module}/templates/state_machine_definition.tftpl", {
    function_arns = var.lambda_function_arns
    region = var.region
    account_id = data.aws_caller_identity.current.account_id
    dynamodb_verification_table = var.dynamodb_verification_table
    dynamodb_conversation_table = var.dynamodb_conversation_table
  })
  filename = "${path.module}/generated_definition.json"
}

# Get current AWS account ID
data "aws_caller_identity" "current" {}

# Step Functions State Machine with enhanced definition
resource "aws_sfn_state_machine" "verification_workflow" {
  name     = var.state_machine_name
  role_arn = aws_iam_role.step_functions_role.arn

  definition = templatefile("${path.module}/templates/state_machine_definition.tftpl", {
    function_arns = var.lambda_function_arns
    region = var.region
    account_id = data.aws_caller_identity.current.account_id
    dynamodb_verification_table = var.dynamodb_verification_table
    dynamodb_conversation_table = var.dynamodb_conversation_table
  })

  logging_configuration {
    log_destination        = "${aws_cloudwatch_log_group.step_functions_logs.arn}:*"
    include_execution_data = true
    level                  = var.log_level
  }

  tracing_configuration {
    enabled = var.enable_x_ray_tracing
  }

  type = "STANDARD"

  tags = merge(
    var.common_tags,
    {
      Name = var.state_machine_name
    }
  )
}

# API Gateway integration with Step Functions
resource "aws_api_gateway_integration" "step_functions_start" {
  count               = var.create_api_gateway_integration ? 1 : 0
  rest_api_id         = var.api_gateway_id
  resource_id         = aws_api_gateway_resource.step_functions[0].id
  http_method         = aws_api_gateway_method.step_functions_start[0].http_method
  type                = "AWS"
  integration_http_method = "POST"
  uri                 = "arn:aws:apigateway:${var.region}:states:action/StartExecution"
  credentials         = aws_iam_role.api_gateway_step_functions_role[0].arn
  
  request_templates = {
    "application/json" = <<EOF
{
  "input": "$util.escapeJavaScript($input.json('$'))",
  "stateMachineArn": "${aws_sfn_state_machine.verification_workflow.arn}"
}
EOF
  }
}

# API Gateway integration response
resource "aws_api_gateway_integration_response" "step_functions_start" {
  count               = var.create_api_gateway_integration ? 1 : 0
  rest_api_id         = var.api_gateway_id
  resource_id         = aws_api_gateway_resource.step_functions[0].id
  http_method         = aws_api_gateway_method.step_functions_start[0].http_method
  status_code         = "200"
  
  response_templates = {
    "application/json" = <<EOF
{
  "executionArn": "$input.path('$.executionArn')",
  "startDate": "$input.path('$.startDate')"
}
EOF
  }
  
  depends_on = [
    aws_api_gateway_integration.step_functions_start
  ]
}

# API Gateway method response
resource "aws_api_gateway_method_response" "step_functions_start" {
  count       = var.create_api_gateway_integration ? 1 : 0
  rest_api_id = var.api_gateway_id
  resource_id = aws_api_gateway_resource.step_functions[0].id
  http_method = aws_api_gateway_method.step_functions_start[0].http_method
  status_code = "200"
  
  response_models = {}
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = true
  }
}
