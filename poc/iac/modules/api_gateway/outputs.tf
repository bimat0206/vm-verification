# modules/api_gateway/output.tf

output "api_id" {
  description = "ID of the API Gateway"
  value       = aws_api_gateway_rest_api.api.id
}

output "api_name" {
  description = "Name of the API Gateway"
  value       = aws_api_gateway_rest_api.api.name
}

output "api_endpoint" {
  description = "Endpoint URL of the API Gateway"
  value       = "https://${aws_api_gateway_rest_api.api.id}.execute-api.${var.region}.amazonaws.com/${var.stage_name}"
}

output "api_arn" {
  description = "ARN of the API Gateway"
  value       = aws_api_gateway_rest_api.api.arn
}

output "api_execution_arn" {
  description = "Execution ARN of the API Gateway"
  value       = aws_api_gateway_rest_api.api.execution_arn
}

output "invoke_url" {
  description = "Invoke URL of the API Gateway stage"
  value       = "https://${aws_api_gateway_rest_api.api.id}.execute-api.${var.region}.amazonaws.com/${var.stage_name}"
}

output "api_key_value" {
  description = "Value of the API key"
  value       = var.use_api_key ? aws_api_gateway_api_key.api_key[0].value : null
  sensitive   = true
}

output "root_resource_id" {
  description = "ID of the API Gateway root resource"
  value       = aws_api_gateway_rest_api.api.root_resource_id
}

output "step_functions_role_arn" {
  description = "ARN of the IAM role for API Gateway to invoke Step Functions"
  value       = aws_iam_role.api_gateway_step_functions_role.arn
}
