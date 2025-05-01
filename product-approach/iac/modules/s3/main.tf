# Reference Bucket for layout images
resource "aws_s3_bucket" "reference" {
  bucket        = var.reference_bucket_name
  force_destroy = var.force_destroy

  tags = merge(
    var.common_tags,
    {
      Name = var.reference_bucket_name
    }
  )
}

# Checking Bucket for uploaded checking images
resource "aws_s3_bucket" "checking" {
  bucket        = var.checking_bucket_name
  force_destroy = var.force_destroy

  tags = merge(
    var.common_tags,
    {
      Name = var.checking_bucket_name
    }
  )
}

# Results Bucket for verification results and visualizations
resource "aws_s3_bucket" "results" {
  bucket        = var.results_bucket_name
  force_destroy = var.force_destroy

  tags = merge(
    var.common_tags,
    {
      Name = var.results_bucket_name
    }
  )
}

# Server-side encryption for all buckets
resource "aws_s3_bucket_server_side_encryption_configuration" "reference" {
  bucket = aws_s3_bucket.reference.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "checking" {
  bucket = aws_s3_bucket.checking.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "results" {
  bucket = aws_s3_bucket.results.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Versioning configuration
resource "aws_s3_bucket_versioning" "reference" {
  bucket = aws_s3_bucket.reference.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_versioning" "checking" {
  bucket = aws_s3_bucket.checking.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_versioning" "results" {
  bucket = aws_s3_bucket.results.id
  versioning_configuration {
    status = "Enabled"
  }
}

# Public access block for all buckets
resource "aws_s3_bucket_public_access_block" "reference" {
  bucket                  = aws_s3_bucket.reference.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_public_access_block" "checking" {
  bucket                  = aws_s3_bucket.checking.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_public_access_block" "results" {
  bucket                  = aws_s3_bucket.results.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# CORS configuration for reference bucket
resource "aws_s3_bucket_cors_configuration" "reference" {
  bucket = aws_s3_bucket.reference.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = ["*"] # In production, this should be restricted to specific domains
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

# CORS configuration for checking bucket
resource "aws_s3_bucket_cors_configuration" "checking" {
  bucket = aws_s3_bucket.checking.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD", "PUT", "POST"]
    allowed_origins = ["*"] # In production, this should be restricted to specific domains
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

# CORS configuration for results bucket
resource "aws_s3_bucket_cors_configuration" "results" {
  bucket = aws_s3_bucket.results.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = ["*"] # In production, this should be restricted to specific domains
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

# Reference bucket lifecycle configuration
resource "aws_s3_bucket_lifecycle_configuration" "reference" {
  bucket = aws_s3_bucket.reference.id

  dynamic "rule" {
    for_each = var.reference_lifecycle_rules

    content {
      id     = rule.value.id
      status = rule.value.enabled ? "Enabled" : "Disabled"

      dynamic "filter" {
        for_each = rule.value.prefix != null ? [rule.value.prefix] : []
        content {
          prefix = filter.value
        }
      }

      dynamic "expiration" {
        for_each = rule.value.expiration_days != null ? [rule.value.expiration_days] : []
        content {
          days = expiration.value
        }
      }

      dynamic "noncurrent_version_expiration" {
        for_each = rule.value.noncurrent_version_expiration_days != null ? [rule.value.noncurrent_version_expiration_days] : []
        content {
          noncurrent_days = noncurrent_version_expiration.value
        }
      }

      dynamic "abort_incomplete_multipart_upload" {
        for_each = rule.value.abort_incomplete_multipart_upload_days != null ? [rule.value.abort_incomplete_multipart_upload_days] : []
        content {
          days_after_initiation = abort_incomplete_multipart_upload.value
        }
      }
    }
  }
}

# Checking bucket lifecycle configuration
resource "aws_s3_bucket_lifecycle_configuration" "checking" {
  bucket = aws_s3_bucket.checking.id

  dynamic "rule" {
    for_each = var.checking_lifecycle_rules

    content {
      id     = rule.value.id
      status = rule.value.enabled ? "Enabled" : "Disabled"

      dynamic "filter" {
        for_each = rule.value.prefix != null ? [rule.value.prefix] : []
        content {
          prefix = filter.value
        }
      }

      dynamic "expiration" {
        for_each = rule.value.expiration_days != null ? [rule.value.expiration_days] : []
        content {
          days = expiration.value
        }
      }

      dynamic "noncurrent_version_expiration" {
        for_each = rule.value.noncurrent_version_expiration_days != null ? [rule.value.noncurrent_version_expiration_days] : []
        content {
          noncurrent_days = noncurrent_version_expiration.value
        }
      }

      dynamic "abort_incomplete_multipart_upload" {
        for_each = rule.value.abort_incomplete_multipart_upload_days != null ? [rule.value.abort_incomplete_multipart_upload_days] : []
        content {
          days_after_initiation = abort_incomplete_multipart_upload.value
        }
      }
    }
  }
}

# Results bucket lifecycle configuration
resource "aws_s3_bucket_lifecycle_configuration" "results" {
  bucket = aws_s3_bucket.results.id

  dynamic "rule" {
    for_each = var.results_lifecycle_rules

    content {
      id     = rule.value.id
      status = rule.value.enabled ? "Enabled" : "Disabled"

      dynamic "filter" {
        for_each = rule.value.prefix != null ? [rule.value.prefix] : []
        content {
          prefix = filter.value
        }
      }

      dynamic "expiration" {
        for_each = rule.value.expiration_days != null ? [rule.value.expiration_days] : []
        content {
          days = expiration.value
        }
      }

      dynamic "noncurrent_version_expiration" {
        for_each = rule.value.noncurrent_version_expiration_days != null ? [rule.value.noncurrent_version_expiration_days] : []
        content {
          noncurrent_days = noncurrent_version_expiration.value
        }
      }

      dynamic "abort_incomplete_multipart_upload" {
        for_each = rule.value.abort_incomplete_multipart_upload_days != null ? [rule.value.abort_incomplete_multipart_upload_days] : []
        content {
          days_after_initiation = abort_incomplete_multipart_upload.value
        }
      }
    }
  }
}

# Create folder structure in reference bucket
resource "aws_s3_object" "reference_raw_folder" {
  bucket       = aws_s3_bucket.reference.id
  key          = "raw/"
  content_type = "application/x-directory"
  content      = ""
}

resource "aws_s3_object" "reference_processed_folder" {
  bucket       = aws_s3_bucket.reference.id
  key          = "processed/"
  content_type = "application/x-directory"
  content      = ""
}

resource "aws_s3_object" "reference_logs_folder" {
  bucket       = aws_s3_bucket.reference.id
  key          = "logs/"
  content_type = "application/x-directory"
  content      = ""
}

# Create folder structure in checking bucket based on current date
resource "aws_s3_object" "checking_current_date_folder" {
  bucket       = aws_s3_bucket.checking.id
  key          = "${formatdate("YYYY-MM-DD", timestamp())}/"
  content_type = "application/x-directory"
  content      = ""
}

# Create folder structure in results bucket based on current date
resource "aws_s3_object" "results_current_date_folder" {
  bucket       = aws_s3_bucket.results.id
  key          = "${formatdate("YYYY-MM-DD", timestamp())}/"
  content_type = "application/x-directory"
  content      = ""
}