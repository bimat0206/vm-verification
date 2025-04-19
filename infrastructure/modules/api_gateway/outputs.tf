# infrastructure/modules/api_gateway/outputs.tf
output "api_id" {
  description = "ID of the API Gateway REST API"
  value       = aws_api_gateway_rest_api.verification_api.id
}

output "api_arn" {
  description = "ARN of the API Gateway REST API"
  value       = aws_api_gateway_rest_api.verification_api.execution_arn
}

output "invoke_url" {
  description = "Base URL for invoking the API"
  value       = "${aws_api_gateway_deployment.deployment.invoke_url}${var.stage_name}"
}

output "api_endpoint" {
  description = "The endpoint URL of the API Gateway"
  value       = aws_api_gateway_deployment.deployment.invoke_url
}

output "stage_name" {
  description = "The stage name of the API deployment"
  value       = var.stage_name
}