output "verification_results_table_name" {
  description = "Name of the verification results DynamoDB table"
  value       = aws_dynamodb_table.verification_results.name
}

output "verification_results_table_arn" {
  description = "ARN of the verification results DynamoDB table"
  value       = aws_dynamodb_table.verification_results.arn
}

output "verification_results_table_id" {
  description = "ID of the verification results DynamoDB table"
  value       = aws_dynamodb_table.verification_results.id
}

output "layout_metadata_table_name" {
  description = "Name of the layout metadata DynamoDB table"
  value       = aws_dynamodb_table.layout_metadata.name
}

output "layout_metadata_table_arn" {
  description = "ARN of the layout metadata DynamoDB table"
  value       = aws_dynamodb_table.layout_metadata.arn
}

output "layout_metadata_table_id" {
  description = "ID of the layout metadata DynamoDB table"
  value       = aws_dynamodb_table.layout_metadata.id
}

output "conversation_history_table_name" {
  description = "Name of the conversation history DynamoDB table"
  value       = aws_dynamodb_table.conversation_history.name
}

output "conversation_history_table_arn" {
  description = "ARN of the conversation history DynamoDB table"
  value       = aws_dynamodb_table.conversation_history.arn
}

output "conversation_history_table_id" {
  description = "ID of the conversation history DynamoDB table"
  value       = aws_dynamodb_table.conversation_history.id
}

output "billing_mode" {
  description = "Billing mode used for DynamoDB tables"
  value       = var.billing_mode
}

output "read_capacity" {
  description = "Read capacity units for DynamoDB tables (if applicable)"
  value       = var.billing_mode == "PROVISIONED" ? var.read_capacity : "N/A (pay-per-request mode)"
}

output "write_capacity" {
  description = "Write capacity units for DynamoDB tables (if applicable)"
  value       = var.billing_mode == "PROVISIONED" ? var.write_capacity : "N/A (pay-per-request mode)"
}

output "point_in_time_recovery_enabled" {
  description = "Whether point-in-time recovery is enabled for the DynamoDB tables"
  value       = var.point_in_time_recovery
}