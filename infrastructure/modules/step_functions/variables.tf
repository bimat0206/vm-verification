# infrastructure/modules/step_functions/variables.tf
variable "state_machine_name" {
  description = "Name of the Step Functions state machine"
  type        = string
}

variable "environment" {
  description = "Environment name (e.g., dev, prod)"
  type        = string
  default     = "dev"
}

variable "initialize_function_arn" {
  description = "ARN of the Initialize Lambda function"
  type        = string
}

variable "fetch_images_function_arn" {
  description = "ARN of the FetchImages Lambda function"
  type        = string
}

variable "prepare_prompt_function_arn" {
  description = "ARN of the PreparePrompt Lambda function"
  type        = string
}

variable "invoke_bedrock_function_arn" {
  description = "ARN of the InvokeBedrock Lambda function"
  type        = string
}

variable "process_results_function_arn" {
  description = "ARN of the ProcessResults Lambda function"
  type        = string
}

variable "store_results_function_arn" {
  description = "ARN of the StoreResults Lambda function"
  type        = string
}

variable "notify_function_arn" {
  description = "ARN of the Notify Lambda function"
  type        = string
}

variable "tags" {
  description = "Additional tags to apply to the state machine"
  type        = map(string)
  default     = {}
}