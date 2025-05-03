
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
  
  response_models = {
    "application/json" = "Empty"
  }
  
  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = true
  }
}