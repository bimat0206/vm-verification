# modules/secretsmanager/main.tf

resource "aws_secretsmanager_secret" "secret" {
  name        = var.secret_name
  description = var.secret_description

  tags = merge(
    var.common_tags,
    {
      Name = var.secret_name
    }
  )
}

resource "aws_secretsmanager_secret_version" "secret_version" {
  secret_id     = aws_secretsmanager_secret.secret.id
  secret_string = jsonencode({
    api_key = var.secret_value
  })
}