# Changelog

## [1.5.0] - 2025-05-19

### Fixed
- Fixed IAM policy error in Step Functions module: "MalformedPolicyDocument: Policy statement must contain resources"
- Removed S3 state management policy when s3_state_bucket_arn is null
- Added state lifecycle rules to terraform.tfvars for proper S3 bucket configuration
- Removed default values from variables.tf for s3_buckets to ensure values are defined in terraform.tfvars

### Changed
- Updated main.tf to properly handle Step Functions module integration with S3 state bucket
- Improved error handling in Step Functions module for S3 state management policy

## [1.4.0] - 2025-05-18

### Added
- Added S3 state management bucket configuration
- Added lifecycle rules for state bucket data management
- Added integration between Step Functions and S3 state bucket

### Changed
- Updated S3 module to support state management bucket
- Updated Step Functions module to access S3 state bucket