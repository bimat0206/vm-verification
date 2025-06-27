# ECS React Module Changelog

## [1.0.0] - 2025-01-07

### Added
- Initial creation of ECS React module based on ecs-streamlit module
- Support for React/Next.js applications running on ECS Fargate
- Application Load Balancer (ALB) with separate domain naming: `vm-hub-{prefix}-alb`
- Auto-scaling configuration for ECS service
- IAM roles and policies for ECS tasks with access to:
  - API Gateway
  - S3 buckets (reference, checking, results)
  - DynamoDB tables (verification_results, layout_metadata, conversation_history)
  - Secrets Manager
  - CloudWatch Logs
  - ECR repositories
- Health check configuration for Next.js applications (`/api/health` endpoint)
- Environment variables support with secrets integration
- HTTPS support with SSL certificate configuration
- CloudWatch logging with configurable retention
- Security groups for ALB and ECS tasks
- Target group configuration for load balancing

### Configuration
- Default container port: 3000 (Next.js standard)
- Default CPU: 1024 units
- Default memory: 2048 MiB
- Default health check path: `/api/health`
- Environment variables:
  - `PORT`: Container port
  - `NODE_ENV`: Set to "production"
  - `NEXT_TELEMETRY_DISABLED`: Set to "1"
  - `CONFIG_SECRET`: Reference to configuration secret in Secrets Manager
  - `API_KEY_SECRET_NAME`: Reference to API key secret in Secrets Manager

### Dependencies
- VPC module (shared with Streamlit)
- API Gateway module
- S3 buckets module
- DynamoDB tables module
- Secrets Manager modules (separate config secret for React)

### Files Created
- `main.tf`: Core ECS cluster, service, and task definition
- `alb.tf`: Application Load Balancer configuration
- `iam.tf`: IAM roles and policies
- `variables.tf`: Input variables
- `outputs.tf`: Module outputs
- `CHANGELOG.md`: This changelog

### Integration
- Added to main Terraform configuration in `iac/main.tf`
- Added React frontend variables to `iac/variables.tf`
- Created separate secrets manager module for React configuration
- Updated VPC module condition to include React frontend
- Added dedicated ECR repository `react_frontend` in `iac/locals.tf`
- Added ECR repository for Streamlit frontend as well for consistency

### ECR Repository
- Repository name: `{project}-{environment}-ecr-react-frontend-{suffix}`
- Image tag mutability: MUTABLE
- Scan on push: Enabled
- Encryption: AES256
- Force delete: Disabled (for safety)
