# Changelog

All notable changes to the API Images Upload Lambda Function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-07

### Added
- Initial implementation of file upload Lambda function
- Support for uploading files to S3 buckets (REFERENCE_BUCKET and CHECKING_BUCKET)
- File type validation for images and documents
- File size validation (10MB limit)
- Path sanitization and organization within buckets
- CORS support for cross-origin requests
- Structured logging with logrus
- Comprehensive error handling and responses
- Docker containerization with multi-stage build
- Automated deployment script with ECR integration
- Support for both reference and checking bucket uploads
- Query parameter support for bucket type, file name, and path
- Content type detection based on file extensions

### Security
- File type whitelist validation
- Path traversal protection
- File size limits to prevent abuse
- Proper CORS configuration

### Documentation
- Complete README with API documentation
- Usage examples and development guide
- Architecture overview
- Security considerations

### Infrastructure
- Go 1.20 runtime
- AWS SDK v2 integration
- Lambda function handler
- ECR repository support
- Terraform-compatible deployment
