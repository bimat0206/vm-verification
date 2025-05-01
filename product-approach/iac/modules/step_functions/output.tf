output "state_machine_id" {
  description = "ID of the Step Functions state machine"
  value       = aws_sfn_state_machine.verification_workflow.id
}

output "state_machine_arn" {
  description = "ARN of the Step Functions state machine"
  value       = aws_sfn_state_machine.verification_workflow.arn
}

output "state_machine_name" {
  description = "Name of the Step Functions state machine"
  value       = aws_sfn_state_machine.verification_workflow.name
}

output "state_machine_creation_date" {
  description = "Creation date of the Step Functions state machine"
  value       = aws_sfn_state_machine.verification_workflow.creation_date
}

output "state_machine_status" {
  description = "Status of the Step Functions state machine"
  value       = aws_sfn_state_machine.verification_workflow.status
}

output "state_machine_role_arn" {
  description = "ARN of the IAM role used by the Step Functions state machine"
  value       = aws_iam_role.step_functions_role.arn
}

output "state_machine_role_name" {
  description = "Name of the IAM role used by the Step Functions state machine"
  value       = aws_iam_role.step_functions_role.name
}

output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group for Step Functions"
  value       = aws_cloudwatch_log_group.step_functions_logs.name
}

output "cloudwatch_log_group_arn" {
  description = "ARN of the CloudWatch log group for Step Functions"
  value       = aws_cloudwatch_log_group.step_functions_logs.arn
}