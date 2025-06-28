output "reference_bucket_name" {
  description = "Name of the reference layout S3 bucket"
  value       = aws_s3_bucket.reference.id
}

output "reference_bucket_arn" {
  description = "ARN of the reference layout S3 bucket"
  value       = aws_s3_bucket.reference.arn
}

output "reference_bucket_domain_name" {
  description = "Domain name of the reference layout S3 bucket"
  value       = aws_s3_bucket.reference.bucket_domain_name
}

output "reference_bucket_regional_domain_name" {
  description = "Regional domain name of the reference layout S3 bucket"
  value       = aws_s3_bucket.reference.bucket_regional_domain_name
}

output "checking_bucket_name" {
  description = "Name of the checking images S3 bucket"
  value       = aws_s3_bucket.checking.id
}

output "checking_bucket_arn" {
  description = "ARN of the checking images S3 bucket"
  value       = aws_s3_bucket.checking.arn
}

output "checking_bucket_domain_name" {
  description = "Domain name of the checking images S3 bucket"
  value       = aws_s3_bucket.checking.bucket_domain_name
}

output "checking_bucket_regional_domain_name" {
  description = "Regional domain name of the checking images S3 bucket"
  value       = aws_s3_bucket.checking.bucket_regional_domain_name
}




output "state_bucket_name" {
  description = "Name of the temporary Base64 S3 bucket"
  value       = aws_s3_bucket.state.id
}

output "state_bucket_arn" {
  description = "ARN of the temporary Base64 S3 bucket"
  value       = aws_s3_bucket.state.arn
}