# infrastructure/modules/multi_lambda/outputs.tf

data "aws_caller_identity" "current" {}

output "function_names" {
  description = "Map of function keys to their names"
  value = var.skip_lambda_function_creation ? {
    for key, lambda in local.lambda_functions : key => lambda.name
  } : {
    for key, lambda in aws_lambda_function.function : key => lambda.function_name
  }
}

output "function_arns" {
  description = "Map of function keys to their ARNs"
  value = var.skip_lambda_function_creation ? {
    for key, lambda in local.lambda_functions : key => "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${lambda.name}"
  } : {
    for key, lambda in aws_lambda_function.function : key => lambda.arn
  }
}

output "invoke_arns" {
  description = "Map of function keys to their invoke ARNs"
  value = var.skip_lambda_function_creation ? {
    for key, lambda in local.lambda_functions : key => "arn:aws:apigateway:${var.aws_region}:lambda:path/2015-03-31/functions/arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${lambda.name}/invocations"
  } : {
    for key, lambda in aws_lambda_function.function : key => lambda.invoke_arn
  }
}

# Individual function outputs for specific functions
output "initialize_function_name" {
  description = "Name of the Initialize Lambda function"
  value       = var.skip_lambda_function_creation ? local.lambda_functions["initialize"].name : aws_lambda_function.function["initialize"].function_name
}

output "initialize_function_arn" {
  description = "ARN of the Initialize Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["initialize"].name}" : aws_lambda_function.function["initialize"].arn
}

output "fetch_images_function_name" {
  description = "Name of the FetchImages Lambda function"
  value       = var.skip_lambda_function_creation ? local.lambda_functions["fetch_images"].name : aws_lambda_function.function["fetch_images"].function_name
}

output "fetch_images_function_arn" {
  description = "ARN of the FetchImages Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["fetch_images"].name}" : aws_lambda_function.function["fetch_images"].arn
}

output "prepare_prompt_function_name" {
  description = "Name of the PreparePrompt Lambda function"
  value       = var.skip_lambda_function_creation ? local.lambda_functions["prepare_prompt"].name : aws_lambda_function.function["prepare_prompt"].function_name
}

output "prepare_prompt_function_arn" {
  description = "ARN of the PreparePrompt Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["prepare_prompt"].name}" : aws_lambda_function.function["prepare_prompt"].arn
}

output "invoke_bedrock_function_name" {
  description = "Name of the InvokeBedrock Lambda function"
  value       = var.skip_lambda_function_creation ? local.lambda_functions["invoke_bedrock"].name : aws_lambda_function.function["invoke_bedrock"].function_name
}

output "invoke_bedrock_function_arn" {
  description = "ARN of the InvokeBedrock Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["invoke_bedrock"].name}" : aws_lambda_function.function["invoke_bedrock"].arn
}

output "process_results_function_name" {
  description = "Name of the ProcessResults Lambda function"
  value       = var.skip_lambda_function_creation ? local.lambda_functions["process_results"].name : aws_lambda_function.function["process_results"].function_name
}

output "process_results_function_arn" {
  description = "ARN of the ProcessResults Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["process_results"].name}" : aws_lambda_function.function["process_results"].arn
}

output "store_results_function_name" {
  description = "Name of the StoreResults Lambda function"
  value       = var.skip_lambda_function_creation ? local.lambda_functions["store_results"].name : aws_lambda_function.function["store_results"].function_name
}

output "store_results_function_arn" {
  description = "ARN of the StoreResults Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["store_results"].name}" : aws_lambda_function.function["store_results"].arn
}

output "notify_function_name" {
  description = "Name of the Notify Lambda function"
  value       = var.skip_lambda_function_creation ? local.lambda_functions["notify"].name : aws_lambda_function.function["notify"].function_name
}

output "notify_function_arn" {
  description = "ARN of the Notify Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["notify"].name}" : aws_lambda_function.function["notify"].arn
}

output "get_comparison_function_name" {
  description = "Name of the GetComparison Lambda function"
  value       = var.skip_lambda_function_creation ? local.lambda_functions["get_comparison"].name : aws_lambda_function.function["get_comparison"].function_name
}

output "get_comparison_function_arn" {
  description = "ARN of the GetComparison Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["get_comparison"].name}" : aws_lambda_function.function["get_comparison"].arn
}

output "get_comparison_invoke_arn" {
  description = "Invoke ARN of the GetComparison Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:apigateway:${var.aws_region}:lambda:path/2015-03-31/functions/arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["get_comparison"].name}/invocations" : aws_lambda_function.function["get_comparison"].invoke_arn
}

output "get_images_function_name" {
  description = "Name of the GetImages Lambda function"
  value       = var.skip_lambda_function_creation ? local.lambda_functions["get_images"].name : aws_lambda_function.function["get_images"].function_name
}

output "get_images_function_arn" {
  description = "ARN of the GetImages Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["get_images"].name}" : aws_lambda_function.function["get_images"].arn
}

output "get_images_invoke_arn" {
  description = "Invoke ARN of the GetImages Lambda function"
  value       = var.skip_lambda_function_creation ? "arn:aws:apigateway:${var.aws_region}:lambda:path/2015-03-31/functions/arn:aws:lambda:${var.aws_region}:${data.aws_caller_identity.current.account_id}:function:${local.lambda_functions["get_images"].name}/invocations" : aws_lambda_function.function["get_images"].invoke_arn
}

output "lambda_role_name" {
  description = "Name of the IAM role for the Lambda functions"
  value       = aws_iam_role.lambda_role.name
}

output "lambda_role_arn" {
  description = "ARN of the IAM role for the Lambda functions"
  value       = aws_iam_role.lambda_role.arn
}