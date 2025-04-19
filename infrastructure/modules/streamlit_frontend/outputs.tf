# infrastructure/modules/streamlit_frontend/outputs.tf

output "ecr_repository_url" {
  description = "URL of the ECR repository for Streamlit app"
  value       = aws_ecr_repository.streamlit_app.repository_url
}

output "ecr_repository_arn" {
  description = "ARN of the ECR repository for Streamlit app"
  value       = aws_ecr_repository.streamlit_app.arn
}

output "secret_arn" {
  description = "ARN of the secret containing Streamlit app configuration"
  value       = aws_secretsmanager_secret.streamlit_config.arn
}

output "secret_name" {
  description = "Name of the secret containing Streamlit app configuration"
  value       = aws_secretsmanager_secret.streamlit_config.name
}

output "app_runner_service_id" {
  description = "ID of the App Runner service"
  value       = aws_apprunner_service.streamlit_service.id
}

output "app_runner_service_arn" {
  description = "ARN of the App Runner service"
  value       = aws_apprunner_service.streamlit_service.arn
}

output "app_runner_service_url" {
  description = "URL of the App Runner service"
  value       = aws_apprunner_service.streamlit_service.service_url
}

output "app_runner_service_status" {
  description = "Status of the App Runner service"
  value       = aws_apprunner_service.streamlit_service.status
}

output "app_runner_role_arn" {
  description = "ARN of the IAM role for App Runner"
  value       = aws_iam_role.app_runner_role.arn
}

output "app_runner_role_name" {
  description = "Name of the IAM role for App Runner"
  value       = aws_iam_role.app_runner_role.name
}