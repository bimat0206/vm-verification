# infrastructure/modules/multi_ecr/outputs.tf

output "repository_urls" {
  description = "Map of function names to their ECR repository URLs"
  value = {
    for name, repo in aws_ecr_repository.function_repos : name => repo.repository_url
  }
}

output "repository_arns" {
  description = "Map of function names to their ECR repository ARNs"
  value = {
    for name, repo in aws_ecr_repository.function_repos : name => repo.arn
  }
}

output "ecr_policy_arn" {
  description = "The ARN of the ECR IAM policy"
  value       = aws_iam_policy.ecr_policy.arn
}

# Individual repository outputs for specific functions
output "initialize_repository_url" {
  description = "URL of the Initialize function ECR repository"
  value       = aws_ecr_repository.function_repos["initialize"].repository_url
}

output "fetch_images_repository_url" {
  description = "URL of the fetch-images function ECR repository"
  value       = aws_ecr_repository.function_repos["fetch-images"].repository_url
}

output "prepare_prompt_repository_url" {
  description = "URL of the prepare-prompt function ECR repository"
  value       = aws_ecr_repository.function_repos["prepare-prompt"].repository_url
}

output "invoke_bedrock_repository_url" {
  description = "URL of the invoke-bedrock function ECR repository"
  value       = aws_ecr_repository.function_repos["invoke-bedrock"].repository_url
}

output "process_results_repository_url" {
  description = "URL of the process-results function ECR repository"
  value       = aws_ecr_repository.function_repos["process-results"].repository_url
}

output "store_results_repository_url" {
  description = "URL of the store-results function ECR repository"
  value       = aws_ecr_repository.function_repos["store-results"].repository_url
}

output "notify_repository_url" {
  description = "URL of the notify function ECR repository"
  value       = aws_ecr_repository.function_repos["notify"].repository_url
}

output "get_comparison_repository_url" {
  description = "URL of the get-comparison function ECR repository"
  value       = aws_ecr_repository.function_repos["get-comparison"].repository_url
}

output "get_images_repository_url" {
  description = "URL of the get-images function ECR repository"
  value       = aws_ecr_repository.function_repos["get-images"].repository_url
}