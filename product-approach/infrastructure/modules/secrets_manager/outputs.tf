# infrastructure/modules/secrets_manager/outputs.tf
output "secret_arn" {
  description = "ARN of the secret"
  value       = aws_secretsmanager_secret.secret.arn
}

output "secret_id" {
  description = "ID of the secret"
  value       = aws_secretsmanager_secret.secret.id
}

output "bedrock_policy_arn" {
  description = "ARN of the Bedrock policy, if created"
  value       = var.create_bedrock_policy ? aws_iam_policy.bedrock_policy[0].arn : null
}