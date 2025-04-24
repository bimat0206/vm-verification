# infrastructure/modules/ecr/main.tf

locals {
  common_tags = merge(
    var.tags,
    {
      Environment = var.environment
      Name        = "${var.name_prefix}-ecr-repository"
    }
  )

  # Define the function names that need repositories
  function_names = [
    "initialize",
    "fetch-images",
    "prepare-prompt",
    "invoke-bedrock",
    "process-results",
    "store-results",
    "notify",
    "get-comparison",
    "get-images"
  ]
}

# Create ECR repositories for each function
resource "aws_ecr_repository" "function_repos" {
  for_each             = toset(local.function_names)
  name                 = "${var.repository_prefix}-${each.value}"
  image_tag_mutability = var.image_tag_mutability

  image_scanning_configuration {
    scan_on_push = var.enable_scan_on_push
  }

  encryption_configuration {
    encryption_type = var.kms_key_arn != null ? "KMS" : "AES256"
    kms_key         = var.kms_key_arn
  }

  tags = merge(
    local.common_tags,
    {
      Name = "${var.repository_prefix}-${each.value}"
    }
  )
}

# Add lifecycle policies to all repositories
resource "aws_ecr_lifecycle_policy" "policy" {
  for_each   = aws_ecr_repository.function_repos
  repository = each.value.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1,
        action = {
          type = "expire"
        }
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = var.max_image_count
        }
        description = "Keep only the latest ${var.max_image_count} images"
      }
    ]
  })
}

# IAM policy for ECR access
resource "aws_iam_policy" "ecr_policy" {
  name = "${var.repository_prefix}-ecr-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:GetRepositoryPolicy",
          "ecr:DescribeRepositories",
          "ecr:ListImages",
          "ecr:DescribeImages",
          "ecr:BatchGetImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:PutImage"
        ]
        Resource = [for repo in aws_ecr_repository.function_repos : repo.arn]
      }
    ]
  })

  tags = local.common_tags
}

# Null resource to pull and push nginx images to repositories
# Null resource to pull and push nginx images to repositories
resource "null_resource" "push_placeholder_images" {
  for_each = var.push_placeholder_images ? aws_ecr_repository.function_repos : {}

  triggers = {
    repository_url = each.value.repository_url
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command = <<EOF
      # Continue even if login fails (keychain errors can happen but still work)
      aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${split("/", each.value.repository_url)[0]} || true
      docker pull nginx:latest
      docker tag nginx:latest ${each.value.repository_url}:latest
      docker push ${each.value.repository_url}:latest
      echo "Successfully pushed image to ${each.value.repository_url}:latest"
    EOF
  }

  depends_on = [aws_ecr_repository.function_repos]
}