# infrastructure/modules/multi_lambda/outputs.tf
output "function_names" {
  description = "Map of function keys to their names"
  value = {
    for key, lambda in aws_lambda_function.function : key => lambda.function_name
  }
}

output "function_arns" {
  description = "Map of function keys to their ARNs"
  value = {
    for key, lambda in aws_lambda_function.function : key => lambda.arn
  }
}

output "invoke_arns" {
  description = "Map of function keys to their invoke ARNs"
  value = {
    for key, lambda in aws_lambda_function.function : key => lambda.invoke_arn
  }
}

output "initialize_function_name" {
  description = "Name of the Initialize Lambda function"
  value       = aws_lambda_function.function["initialize"].function_name
}

output "initialize_function_arn" {
  description = "ARN of the Initialize Lambda function"
  value       = aws_lambda_function.function["initialize"].arn
}

output "fetch_images_function_name" {
  description = "Name of the FetchImages Lambda function"
  value       = aws_lambda_function.function["fetch_images"].function_name
}

output "fetch_images_function_arn" {
  description = "ARN of the FetchImages Lambda function"
  value       = aws_lambda_function.function["fetch_images"].arn
}

output "prepare_prompt_function_name" {
  description = "Name of the PreparePrompt Lambda function"
  value       = aws_lambda_function.function["prepare_prompt"].function_name
}

output "prepare_prompt_function_arn" {
  description = "ARN of the PreparePrompt Lambda function"
  value       = aws_lambda_function.function["prepare_prompt"].arn
}

output "invoke_bedrock_function_name" {
  description = "Name of the InvokeBedrock Lambda function"
  value       = aws_lambda_function.function["invoke_bedrock"].function_name
}

output "invoke_bedrock_function_arn" {
  description = "ARN of the InvokeBedrock Lambda function"
  value       = aws_lambda_function.function["invoke_bedrock"].arn
}

output "process_results_function_name" {
  description = "Name of the ProcessResults Lambda function"
  value       = aws_lambda_function.function["process_results"].function_name
}

output "process_results_function_arn" {
  description = "ARN of the ProcessResults Lambda function"
  value       = aws_lambda_function.function["process_results"].arn
}

output "store_results_function_name" {
  description = "Name of the StoreResults Lambda function"
  value       = aws_lambda_function.function["store_results"].function_name
}

output "store_results_function_arn" {
  description = "ARN of the StoreResults Lambda function"
  value       = aws_lambda_function.function["store_results"].arn
}

output "notify_function_name" {
  description = "Name of the Notify Lambda function"
  value       = aws_lambda_function.function["notify"].function_name
}

output "notify_function_arn" {
  description = "ARN of the Notify Lambda function"
  value       = aws_lambda_function.function["notify"].arn
}

output "get_comparison_function_name" {
  description = "Name of the GetComparison Lambda function"
  value       = aws_lambda_function.function["get_comparison"].function_name
}

output "get_comparison_function_arn" {
  description = "ARN of the GetComparison Lambda function"
  value       = aws_lambda_function.function["get_comparison"].arn
}

output "get_comparison_invoke_arn" {
  description = "Invoke ARN of the GetComparison Lambda function"
  value       = aws_lambda_function.function["get_comparison"].invoke_arn
}

output "get_images_function_name" {
  description = "Name of the GetImages Lambda function"
  value       = aws_lambda_function.function["get_images"].function_name
}

output "get_images_function_arn" {
  description = "ARN of the GetImages Lambda function"
  value       = aws_lambda_function.function["get_images"].arn
}

output "get_images_invoke_arn" {
  description = "Invoke ARN of the GetImages Lambda function"
  value       = aws_lambda_function.function["get_images"].invoke_arn
}

output "lambda_role_name" {
  description = "Name of the IAM role for the Lambda functions"
  value       = aws_iam_role.lambda_role.name
}

output "lambda_role_arn" {
  description = "ARN of the IAM role for the Lambda functions"
  value       = aws_iam_role.lambda_role.arn
}