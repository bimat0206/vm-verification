output "service_id" {
  description = "ID of the App Runner service"
  value       = aws_apprunner_service.this.id
}

output "service_arn" {
  description = "ARN of the App Runner service"
  value       = aws_apprunner_service.this.arn
}

output "service_name" {
  description = "Name of the App Runner service"
  value       = aws_apprunner_service.this.service_name
}

output "service_url" {
  description = "URL of the App Runner service"
  value       = aws_apprunner_service.this.service_url
}

output "service_status" {
  description = "Status of the App Runner service"
  value       = aws_apprunner_service.this.status
}

output "app_runner_role_arn" {
  description = "ARN of the IAM role used by App Runner to access ECR"
  value       = aws_iam_role.app_runner_role.arn
}

output "app_runner_instance_role_arn" {
  description = "ARN of the IAM role used by App Runner instances"
  value       = aws_iam_role.app_runner_instance_role.arn
}

output "log_group_name" {
  description = "Name of the CloudWatch log group for App Runner"
  value       = aws_cloudwatch_log_group.app_runner.name
}

output "log_group_arn" {
  description = "ARN of the CloudWatch log group for App Runner"
  value       = aws_cloudwatch_log_group.app_runner.arn
}

output "auto_scaling_configuration_arn" {
  description = "ARN of the App Runner auto scaling configuration"
  value       = var.enable_auto_scaling ? aws_apprunner_auto_scaling_configuration_version.this[0].arn : null
}

output "custom_domain_certificates" {
  description = "Certificate verification records for custom domain"
  value       = var.custom_domain_name != "" ? aws_apprunner_custom_domain_association.this[0].certificate_validation_records : null
}

output "custom_domain_dns_target" {
  description = "DNS target for custom domain CNAME record"
  value       = var.custom_domain_name != "" ? aws_apprunner_custom_domain_association.this[0].dns_target : null
}

output "custom_domain_status" {
  description = "Status of the custom domain association"
  value       = var.custom_domain_name != "" ? aws_apprunner_custom_domain_association.this[0].status : null
}