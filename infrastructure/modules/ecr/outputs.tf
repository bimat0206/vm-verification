output "repository_url" {
  description = "The URL of the repository"
  value       = aws_ecr_repository.repository.repository_url
}

output "repository_arn" {
  description = "The ARN of the repository"
  value       = aws_ecr_repository.repository.arn
}

output "repository_name" {
  description = "The name of the repository"
  value       = aws_ecr_repository.repository.name
}

output "ecr_policy_arn" {
  description = "The ARN of the ECR IAM policy"
  value       = aws_iam_policy.ecr_policy.arn
} 