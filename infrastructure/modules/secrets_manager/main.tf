# infrastructure/modules/secrets_manager/main.tf
resource "aws_secretsmanager_secret" "secret" {
  name                    = var.secret_name
  description             = var.description
  recovery_window_in_days = var.recovery_window_in_days
  kms_key_id              = var.kms_key_id

  tags = merge(
    {
      Name        = var.secret_name
      Environment = var.environment
    },
    var.tags
  )
}

resource "aws_secretsmanager_secret_version" "secret_version" {
  count         = var.create_secret_version ? 1 : 0
  secret_id     = aws_secretsmanager_secret.secret.id
  secret_string = var.secret_string_value != null ? var.secret_string_value : jsonencode(var.secret_string_map)
}

# IAM policy for Bedrock access
resource "aws_iam_policy" "bedrock_policy" {
  count = var.create_bedrock_policy ? 1 : 0
  name  = "${var.secret_name}-bedrock-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel"
        ]
        Resource = [
          "arn:aws:bedrock:${var.aws_region}::foundation-model/anthropic.claude-3-5-sonnet-*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          aws_secretsmanager_secret.secret.arn
        ]
      }
    ]
  })
  
  tags = merge(
    {
      Name        = "${var.secret_name}-bedrock-policy"
      Environment = var.environment
    },
    var.tags
  )
}