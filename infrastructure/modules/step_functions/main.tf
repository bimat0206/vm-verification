# infrastructure/modules/step_functions/main.tf
resource "aws_sfn_state_machine" "verification_workflow" {
  name     = var.state_machine_name
  role_arn = aws_iam_role.step_functions_role.arn
  
  # Use the definition directly instead of reading from a file with comments
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

  tags = merge(
    {
      Name        = var.state_machine_name
      Environment = var.environment
    },
    var.tags
  )
}

resource "aws_cloudwatch_log_group" "step_functions_logs" {
  name              = "/aws/states/${var.state_machine_name}"
  retention_in_days = 30

  tags = merge(
    {
      Name        = "/aws/states/${var.state_machine_name}"
      Environment = var.environment
    },
    var.tags
  )
}

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

  tags = merge(
    {
      Name        = "${var.state_machine_name}-role"
      Environment = var.environment
    },
    var.tags
  )
}

resource "aws_iam_policy" "step_functions_policy" {
  name = "${var.state_machine_name}-policy"
  
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
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogGroups",
          "logs:DescribeLogStreams"
        ]
        Resource = "arn:aws:logs:*:*:*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "step_functions_policy_attachment" {
  role       = aws_iam_role.step_functions_role.name
  policy_arn = aws_iam_policy.step_functions_policy.arn
}