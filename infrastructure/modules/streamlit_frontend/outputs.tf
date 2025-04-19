# infrastructure/modules/streamlit_frontend_ecs/outputs.tf

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

output "ecs_cluster_id" {
  description = "ID of the ECS cluster"
  value       = aws_ecs_cluster.streamlit_cluster.id
}

output "ecs_cluster_arn" {
  description = "ARN of the ECS cluster"
  value       = aws_ecs_cluster.streamlit_cluster.arn
}

output "ecs_service_id" {
  description = "ID of the ECS service"
  value       = aws_ecs_service.streamlit_service.id
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.streamlit_service.name
}

output "ecs_task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = aws_ecs_task_definition.streamlit_task.arn
}

output "alb_dns_name" {
  description = "DNS name of the ALB"
  value       = aws_lb.streamlit_alb.dns_name
}

output "alb_zone_id" {
  description = "Zone ID of the ALB"
  value       = aws_lb.streamlit_alb.zone_id
}

output "alb_arn" {
  description = "ARN of the ALB"
  value       = aws_lb.streamlit_alb.arn
}

output "security_group_alb_id" {
  description = "ID of the ALB security group"
  value       = aws_security_group.alb_sg.id
}

output "security_group_ecs_id" {
  description = "ID of the ECS security group"
  value       = aws_security_group.ecs_sg.id
}

output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.streamlit_logs.name
}

output "iam_role_ecs_task_execution_arn" {
  description = "ARN of the ECS task execution role"
  value       = aws_iam_role.ecs_task_execution_role.arn
}

output "iam_role_ecs_task_arn" {
  description = "ARN of the ECS task role"
  value       = aws_iam_role.ecs_task_role.arn
}