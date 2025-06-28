# Changelog

All notable changes to the API Images Browser Lambda function will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-20

### Added
- Initial release of the API Images Browser Lambda function
- S3 bucket browsing functionality for reference and checking buckets
- Support for hierarchical folder navigation
- Automatic image file detection and classification
- RESTful API endpoint `/api/images/browser/{path+}`
- Query parameter support for bucket type selection (`reference` or `checking`)
- Comprehensive error handling with structured error responses
- CORS support for web application integration
- Structured logging with configurable log levels
- Docker containerization for AWS Lambda deployment
- Path sanitization to prevent directory traversal attacks
- Support for multiple image formats (jpg, jpeg, png, gif, bmp, webp, tiff, tif)
- Parent path navigation for folder hierarchy
- File metadata including size and last modified timestamp
- Alphabetical sorting of items (folders first, then files)
- URL path decoding for special characters
- Environment variable configuration for bucket names
- IAM permission documentation
- Comprehensive README with API documentation
- Local development support with Docker

### Security
- Input path validation and sanitization
- Bucket access restricted to configured reference and checking buckets only
- No direct file content access (metadata only)
- CORS headers properly configured for web security

### Performance
- Efficient S3 ListObjectsV2 API usage with delimiter for folder structure
- Limited result sets (max 1000 items per request) to prevent timeouts
- Minimal memory footprint with streaming JSON responses

### Dependencies
- Go 1.20
- AWS Lambda Go SDK v1.41.0
- AWS SDK for Go v2 (S3 service)
- Logrus for structured logging
- Alpine Linux base image for minimal container size

### Configuration
- `REFERENCE_BUCKET`: S3 bucket name for reference images
- `CHECKING_BUCKET`: S3 bucket name for checking images  
- `LOG_LEVEL`: Configurable logging level (DEBUG, INFO, WARN, ERROR)

### API Features
- GET `/api/images/browser/{path+}` endpoint
- Query parameter: `bucketType` (reference|checking)
- JSON response format with items array
- Support for folder and image item types
- Parent path calculation for navigation
- Total item count in response
- HTTP status codes: 200 (success), 400 (bad request), 405 (method not allowed), 500 (server error)

### Documentation
- Complete API documentation in README.md
- IAM permission requirements
- Deployment instructions for AWS Lambda
- Local development setup guide
- Docker build and deployment instructions
- Integration guide for API Gateway
- Troubleshooting section
- Security considerations
- Monitoring and logging guidance

## [Unreleased]

### Planned Features
- Pagination support for large directories
- File content type detection
- Image thumbnail generation
- Search functionality within buckets
- Batch operations support
- Caching for improved performance
- Rate limiting protection
- Enhanced error messages with suggestion
- Support for additional file types
- Metadata filtering options

### Potential Improvements
- Add unit tests and integration tests
- Implement request/response caching
- Add metrics collection for monitoring
- Support for custom S3 endpoints
- Enhanced logging with request tracing
- Performance optimizations for large buckets
- Support for S3 versioning
- Add health check endpoint
- Implement request validation middleware
- Add support for S3 object tagging

---

## Version History

- **1.0.0** (2025-01-20): Initial release with core S3 browsing functionality
