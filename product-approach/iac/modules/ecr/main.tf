resource "aws_ecr_repository" "this" {
  for_each = var.repositories

  name                 = each.key
  image_tag_mutability = each.value.image_tag_mutability
  force_delete         = each.value.force_delete

  image_scanning_configuration {
    scan_on_push = each.value.scan_on_push
  }

  encryption_configuration {
    encryption_type = each.value.encryption_type
    kms_key         = each.value.kms_key
  }

  tags = merge(
    var.common_tags,
    {
      Name = each.key
    }
  )
}

resource "aws_ecr_lifecycle_policy" "this" {
  for_each   = { for k, v in var.repositories : k => v if v.lifecycle_policy != null }
  repository = aws_ecr_repository.this[each.key].name
  policy     = each.value.lifecycle_policy
}

resource "aws_ecr_repository_policy" "this" {
  for_each   = { for k, v in var.repositories : k => v if v.repository_policy != null }
  repository = aws_ecr_repository.this[each.key].name
  policy     = each.value.repository_policy
}