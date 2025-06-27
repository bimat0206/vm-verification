# Outputs for ECS React module

output "service_url" {
  description = "URL of the ALB for the React service"
  value       = "http://${aws_lb.react.dns_name}"
}

output "service_https_url" {
  description = "HTTPS URL of the ALB for the React service (if HTTPS is enabled)"
  value       = var.enable_https ? "https://${aws_lb.react.dns_name}" : null
}

output "alb_dns_name" {
  description = "DNS name of the ALB"
  value       = aws_lb.react.dns_name
}

output "alb_zone_id" {
  description = "Zone ID of the ALB"
  value       = aws_lb.react.zone_id
}

output "alb_arn" {
  description = "ARN of the ALB"
  value       = aws_lb.react.arn
}

output "target_group_arn" {
  description = "ARN of the target group"
  value       = aws_lb_target_group.react.arn
}

output "ecs_cluster_id" {
  description = "ID of the ECS cluster"
  value       = aws_ecs_cluster.react.id
}

output "ecs_cluster_arn" {
  description = "ARN of the ECS cluster"
  value       = aws_ecs_cluster.react.arn
}

output "ecs_service_id" {
  description = "ID of the ECS service"
  value       = aws_ecs_service.react.id
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.react.name
}

output "task_definition_arn" {
  description = "ARN of the task definition"
  value       = aws_ecs_task_definition.react.arn
}

output "task_definition_family" {
  description = "Family of the task definition"
  value       = aws_ecs_task_definition.react.family
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
  value       = aws_cloudwatch_log_group.react_logs.name
}

output "cloudwatch_log_group_arn" {
  description = "ARN of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.react_logs.arn
}
