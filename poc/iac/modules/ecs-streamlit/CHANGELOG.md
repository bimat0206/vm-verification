# Changelog

All notable changes to the ECS Streamlit module will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2024-12-19

### Added
- Support for CONFIG_SECRET environment variable for centralized configuration management
- Enhanced IAM policies to support AWS Secrets Manager access for configuration secrets
- Improved security by moving configuration from individual environment variables to encrypted secrets

### Changed
- Updated IAM role policies to use `config_secret_resource_arn` instead of `secret_resource_arn`
- Modified Secrets Manager policy condition to check for `CONFIG_SECRET` instead of `API_KEY_SECRET_NAME`
- Enhanced configuration management to support both legacy environment variables and new secret-based approach

### Security
- Replaced hardcoded environment variables with secure AWS Secrets Manager integration
- Centralized configuration storage for better security posture
- Reduced environment variable exposure in ECS task definitions

## [1.0.0] - 2024-XX-XX

### Added
- Initial release of the ECS Streamlit module
- Fargate task definition for running Streamlit applications
- Application Load Balancer (ALB) configuration
- Auto-scaling policies
- Security group configurations
- IAM roles and policies for ECS tasks
- CloudWatch logging integration
- Health check configuration
- Tagging support