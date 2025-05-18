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

output "results_bucket_name" {
  description = "Name of the verification results S3 bucket"
  value       = aws_s3_bucket.results.id
}

output "results_bucket_arn" {
  description = "ARN of the verification results S3 bucket"
  value       = aws_s3_bucket.results.arn
}

output "results_bucket_domain_name" {
  description = "Domain name of the verification results S3 bucket"
  value       = aws_s3_bucket.results.bucket_domain_name
}

output "results_bucket_regional_domain_name" {
  description = "Regional domain name of the verification results S3 bucket"
  value       = aws_s3_bucket.results.bucket_regional_domain_name
}
output "temp_base64_bucket_name" {
  description = "Name of the temporary Base64 S3 bucket"
  value       = aws_s3_bucket.temp_base64.id
}

output "temp_base64_bucket_arn" {
  description = "ARN of the temporary Base64 S3 bucket"
  value       = aws_s3_bucket.temp_base64.arn
}