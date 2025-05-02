output "service_id" {
  description = "ID of the Streamlit App Runner service"
  value       = aws_apprunner_service.streamlit.id
}

output "service_arn" {
  description = "ARN of the Streamlit App Runner service"
  value       = aws_apprunner_service.streamlit.arn
}

output "service_name" {
  description = "Name of the Streamlit App Runner service"
  value       = aws_apprunner_service.streamlit.service_name
}

output "service_url" {
  description = "URL of the Streamlit App Runner service"
  value       = aws_apprunner_service.streamlit.service_url
}

output "service_status" {
  description = "Status of the Streamlit App Runner service"
  value       = aws_apprunner_service.streamlit.status
}

output "service_roles" {
  description = "IAM roles used by the Streamlit App Runner service"
  value = {
    service_role  = aws_iam_role.streamlit_role.arn
    instance_role = aws_iam_role.streamlit_instance_role.arn
  }
}

output "log_group_name" {
  description = "Name of the CloudWatch log group for Streamlit App Runner"
  value       = aws_cloudwatch_log_group.streamlit_logs.name
}



output "auto_scaling_configuration" {
  description = "Auto scaling configuration information (if enabled)"
  value = var.enable_auto_scaling ? {
    name = aws_apprunner_auto_scaling_configuration_version.streamlit_scaling[0].auto_scaling_configuration_name
    arn  = aws_apprunner_auto_scaling_configuration_version.streamlit_scaling[0].arn
    min_size = var.min_size
    max_size = var.max_size
    max_concurrency = var.max_concurrency
  } : null
}