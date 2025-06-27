# Outputs for ECS Streamlit module

output "service_url" {
  description = "URL of the ALB for the Streamlit service"
  value       = "http://${aws_lb.streamlit.dns_name}"
}

output "service_https_url" {
  description = "HTTPS URL of the ALB for the Streamlit service (if HTTPS is enabled)"
  value       = var.enable_https ? "https://${aws_lb.streamlit.dns_name}" : null
}

output "alb_dns_name" {
  description = "DNS name of the ALB"
  value       = aws_lb.streamlit.dns_name
}

output "alb_zone_id" {
  description = "Zone ID of the ALB"
  value       = aws_lb.streamlit.zone_id
}

output "alb_arn" {
  description = "ARN of the ALB"
  value       = aws_lb.streamlit.arn
}

output "target_group_arn" {
  description = "ARN of the target group"
  value       = aws_lb_target_group.streamlit.arn
}

output "ecs_cluster_id" {
  description = "ID of the ECS cluster"
  value       = aws_ecs_cluster.streamlit.id
}

output "ecs_cluster_arn" {
  description = "ARN of the ECS cluster"
  value       = aws_ecs_cluster.streamlit.arn
}

output "ecs_service_id" {
  description = "ID of the ECS service"
  value       = aws_ecs_service.streamlit.id
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.streamlit.name
}

output "task_definition_arn" {
  description = "ARN of the task definition"
  value       = aws_ecs_task_definition.streamlit.arn
}

output "task_definition_family" {
  description = "Family of the task definition"
  value       = aws_ecs_task_definition.streamlit.family
}

output "execution_role_arn" {
  description = "ARN of the ECS execution role"
  value       = aws_iam_role.ecs_execution_role.arn
}

output "task_role_arn" {
  description = "ARN of the ECS task role"
  value       = aws_iam_role.ecs_task_role.arn
}

output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.streamlit_logs.name
}

output "cloudwatch_log_group_arn" {
  description = "ARN of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.streamlit_logs.arn
}
