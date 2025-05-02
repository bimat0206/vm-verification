output "api_id" {
  description = "ID of the API Gateway"
  value       = aws_api_gateway_rest_api.this.id
}

output "api_name" {
  description = "Name of the API Gateway"
  value       = aws_api_gateway_rest_api.this.name
}

output "api_arn" {
  description = "ARN of the API Gateway"
  value       = aws_api_gateway_rest_api.this.arn
}

output "api_endpoint" {
  description = "Endpoint URL of the API Gateway"
  value       = aws_api_gateway_rest_api.this.execution_arn
}

output "stage_id" {
  description = "ID of the API Gateway stage"
  value       = aws_api_gateway_stage.this.id
}

output "stage_name" {
  description = "Name of the API Gateway stage"
  value       = aws_api_gateway_stage.this.stage_name
}

output "stage_arn" {
  description = "ARN of the API Gateway stage"
  value       = aws_api_gateway_stage.this.arn
}

output "api_key_id" {
  description = "ID of the API key (if used)"
  value       = var.use_api_key ? aws_api_gateway_api_key.this[0].id : null
}

output "api_key_value" {
  description = "Value of the API key (if used)"
  value       = var.use_api_key ? aws_api_gateway_api_key.this[0].value : null
  sensitive   = true
}

output "invoke_url" {
  description = "Base URL for invoking the API"
  value       = "${aws_api_gateway_stage.this.invoke_url}/api"
}

output "execution_arn" {
  description = "Execution ARN of the API Gateway"
  value       = aws_api_gateway_rest_api.this.execution_arn
}

output "log_group_name" {
  description = "Name of the CloudWatch log group for API Gateway"
  value       = aws_cloudwatch_log_group.api_gateway.name
}

output "log_group_arn" {
  description = "ARN of the CloudWatch log group for API Gateway"
  value       = aws_cloudwatch_log_group.api_gateway.arn
}
