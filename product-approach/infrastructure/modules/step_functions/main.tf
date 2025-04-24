# Step Functions IAM Role Fix
# Add this to your modules/step_functions/main.tf file before the state machine resource

# Enhanced IAM role for Step Functions with proper logging permissions
resource "aws_iam_role" "step_functions_role" {
  name = "${var.state_machine_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "states.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# Allow Step Functions to invoke Lambda functions
resource "aws_iam_policy" "lambda_invoke_policy" {
  name = "${var.state_machine_name}-lambda-invoke"
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "lambda:InvokeFunction"
        ]
        Resource = [
          var.initialize_function_arn,
          var.fetch_images_function_arn,
          var.prepare_prompt_function_arn,
          var.invoke_bedrock_function_arn,
          var.process_results_function_arn,
          var.store_results_function_arn,
          var.notify_function_arn
        ]
      }
    ]
  })
}

# Allow Step Functions to write to CloudWatch Logs
resource "aws_iam_policy" "logs_policy" {
  name = "${var.state_machine_name}-logs"
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogDelivery",
          "logs:GetLogDelivery",
          "logs:UpdateLogDelivery",
          "logs:DeleteLogDelivery",
          "logs:ListLogDeliveries",
          "logs:PutLogEvents",
          "logs:PutResourcePolicy",
          "logs:DescribeResourcePolicies",
          "logs:DescribeLogGroups"
        ]
        Resource = "*"
      }
    ]
  })
}

# Attach policies to the Step Functions role
resource "aws_iam_role_policy_attachment" "lambda_invoke_attachment" {
  role       = aws_iam_role.step_functions_role.name
  policy_arn = aws_iam_policy.lambda_invoke_policy.arn
}

resource "aws_iam_role_policy_attachment" "logs_attachment" {
  role       = aws_iam_role.step_functions_role.name
  policy_arn = aws_iam_policy.logs_policy.arn
}

# Create log group for Step Functions
resource "aws_cloudwatch_log_group" "step_functions_logs" {
  name              = "/aws/stepfunctions/${var.state_machine_name}"
  retention_in_days = var.log_retention_days

  tags = var.tags
}

# Use the enhanced role in the state machine resource
resource "aws_sfn_state_machine" "verification_workflow" {
  name     = var.state_machine_name
  role_arn = aws_iam_role.step_functions_role.arn
  
definition = jsonencode({
    Comment = "Vending Machine Image Verification Workflow",
    StartAt = "Initialize",
    States = {
      Initialize = {
        Type = "Task",
        Resource = var.initialize_function_arn,
        Next = "FetchImages",
        Retry = [
          {
            ErrorEquals = ["States.ALL"],
            IntervalSeconds = 2,
            MaxAttempts = 3,
            BackoffRate = 2
          }
        ],
        Catch = [
          {
            ErrorEquals = ["States.ALL"],
            ResultPath = "$.error",
            Next = "WorkflowFailed"
          }
        ]
      },
      FetchImages = {
        Type = "Task",
        Resource = var.fetch_images_function_arn,
        Next = "PreparePrompt",
        Retry = [
          {
            ErrorEquals = ["States.ALL"],
            IntervalSeconds = 2,
            MaxAttempts = 3,
            BackoffRate = 2
          }
        ],
        Catch = [
          {
            ErrorEquals = ["States.ALL"],
            ResultPath = "$.error",
            Next = "WorkflowFailed"
          }
        ]
      },
      PreparePrompt = {
        Type = "Task",
        Resource = var.prepare_prompt_function_arn,
        Next = "InvokeBedrock",
        Retry = [
          {
            ErrorEquals = ["States.ALL"],
            IntervalSeconds = 2,
            MaxAttempts = 3,
            BackoffRate = 2
          }
        ],
        Catch = [
          {
            ErrorEquals = ["States.ALL"],
            ResultPath = "$.error",
            Next = "WorkflowFailed"
          }
        ]
      },
      InvokeBedrock = {
        Type = "Task",
        Resource = var.invoke_bedrock_function_arn,
        Next = "ProcessResults",
        Retry = [
          {
            ErrorEquals = ["States.ALL"],
            IntervalSeconds = 2,
            MaxAttempts = 3,
            BackoffRate = 2
          }
        ],
        Catch = [
          {
            ErrorEquals = ["States.ALL"],
            ResultPath = "$.error",
            Next = "WorkflowFailed"
          }
        ]
      },
      ProcessResults = {
        Type = "Task",
        Resource = var.process_results_function_arn,
        Next = "StoreResults",
        Retry = [
          {
            ErrorEquals = ["States.ALL"],
            IntervalSeconds = 2,
            MaxAttempts = 3,
            BackoffRate = 2
          }
        ],
        Catch = [
          {
            ErrorEquals = ["States.ALL"],
            ResultPath = "$.error",
            Next = "WorkflowFailed"
          }
        ]
      },
      StoreResults = {
        Type = "Task",
        Resource = var.store_results_function_arn,
        Next = "Notify",
        Retry = [
          {
            ErrorEquals = ["States.ALL"],
            IntervalSeconds = 2,
            MaxAttempts = 3,
            BackoffRate = 2
          }
        ],
        Catch = [
          {
            ErrorEquals = ["States.ALL"],
            ResultPath = "$.error",
            Next = "WorkflowFailed"
          }
        ]
      },
      Notify = {
        Type = "Task",
        Resource = var.notify_function_arn,
        End = true,
        Retry = [
          {
            ErrorEquals = ["States.ALL"],
            IntervalSeconds = 2,
            MaxAttempts = 3,
            BackoffRate = 2
          }
        ],
        Catch = [
          {
            ErrorEquals = ["States.ALL"],
            ResultPath = "$.error",
            Next = "WorkflowFailed"
          }
        ]
      },
      WorkflowFailed = {
        Type = "Fail",
        Error = "WorkflowFailedError",
        Cause = "One of the workflow steps failed. Check the error details."
      }
    }
  })

  
  logging_configuration {
    log_destination        = "${aws_cloudwatch_log_group.step_functions_logs.arn}:*"
    include_execution_data = true
    level                  = "ALL"
  }

  tags = var.tags
}