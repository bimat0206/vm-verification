# infrastructure/modules/step_functions/outputs.tf
output "state_machine_arn" {
  description = "ARN of the Step Functions state machine"
  value       = aws_sfn_state_machine.verification_workflow.arn
}

output "state_machine_name" {
  description = "Name of the Step Functions state machine"
  value       = aws_sfn_state_machine.verification_workflow.name
}

output "role_arn" {
  description = "ARN of the IAM role for the Step Functions state machine"
  value       = aws_iam_role.step_functions_role.arn
}