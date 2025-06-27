# Changelog

All notable changes to the SecretsManager module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2024-12-19

### Enhanced
- Enhanced secret value handling to support both simple strings and JSON objects
- Added automatic detection of JSON content using `can(jsondecode())` function
- Improved flexibility for storing complex configuration data in secrets
- Maintains backward compatibility with existing API key storage format

### Changed
- Modified `aws_secretsmanager_secret_version` resource to intelligently handle JSON vs simple string values
- Updated secret storage logic to preserve JSON structure when provided

## [1.0.0] - 2024-XX-XX

### Added
- Initial release of the SecretsManager module
- Support for creating and managing AWS Secrets Manager secrets
- Secret rotation configuration
- KMS encryption key integration
- Secret policy management
- Secret version management
- Recovery window configuration
- Tagging support