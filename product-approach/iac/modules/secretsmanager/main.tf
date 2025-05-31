# Create a consistent naming convention with random suffix
locals {
  # Standard naming convention for resources
  name_prefix = var.environment != "" ? "${var.project_name}-${var.environment}" : var.project_name

  # Format the secret name using the same pattern as other resources
  secret_name = lower(join("-", compact([local.name_prefix, "secret", var.secret_base_name, var.name_suffix])))
}

resource "aws_secretsmanager_secret" "secret" {
  name        = local.secret_name
  description = var.secret_description

  tags = merge(
    var.common_tags,
    {
      Name = local.secret_name
    }
  )
}

resource "aws_secretsmanager_secret_version" "secret_version" {
  secret_id = aws_secretsmanager_secret.secret.id
  # Check if secret_value is already JSON, if so use it directly, otherwise wrap it as api_key
  secret_string = can(jsondecode(var.secret_value)) ? var.secret_value : jsonencode({
    api_key = var.secret_value
  })
}
