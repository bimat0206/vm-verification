
# Create the state machine definition template file
resource "local_file" "state_machine_definition" {
  count = var.create_definition_file ? 1 : 0

  content = templatefile("${path.module}/templates/state_machine_definition.tftpl", {
    function_arns               = var.lambda_function_arns
    region                      = var.region
    account_id                  = data.aws_caller_identity.current.account_id
    dynamodb_verification_table = var.dynamodb_verification_table
    dynamodb_conversation_table = var.dynamodb_conversation_table
  })
  filename = "${path.module}/generated_definition.json"
}

# Get current AWS account ID
data "aws_caller_identity" "current" {}


# Step Functions State Machine with enhanced definition
resource "aws_sfn_state_machine" "verification_workflow" {
  name     = var.state_machine_name
  role_arn = aws_iam_role.step_functions_role.arn

  definition = templatefile("${path.module}/templates/state_machine_definition.tftpl", {
    function_arns               = var.lambda_function_arns
    region                      = var.region
    account_id                  = data.aws_caller_identity.current.account_id
    dynamodb_verification_table = var.dynamodb_verification_table
    dynamodb_conversation_table = var.dynamodb_conversation_table
  })

  logging_configuration {
    log_destination        = "${aws_cloudwatch_log_group.step_functions_logs.arn}:*"
    include_execution_data = true
    level                  = var.log_level
  }

  tracing_configuration {
    enabled = var.enable_x_ray_tracing
  }

  type = "STANDARD"

  tags = merge(
    var.common_tags,
    {
      Name = var.state_machine_name
    }
  )
}

