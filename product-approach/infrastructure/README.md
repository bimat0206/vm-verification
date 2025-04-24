# Vending Verification Infrastructure

This directory contains the Terraform configuration for deploying the Vending Verification backend infrastructure on AWS.

## Prerequisites

- Terraform installed (version 1.0.0 or later)
- AWS CLI configured with appropriate credentials
- AWS account with necessary permissions

## Infrastructure Components

The infrastructure includes:

- S3 bucket for storing verification images
- DynamoDB table for storing verification results
- EC2 instance for running the backend application
- IAM roles and policies for secure access
- Security group for network access control
- AWS Bedrock access for image verification

## Usage

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Create a `terraform.tfvars` file with your specific values:
   ```hcl
   aws_region = "us-east-1"
   environment = "dev"
   vpc_id = "vpc-xxxxxxxx"
   subnet_id = "subnet-xxxxxxxx"
   ```

3. Review the planned changes:
   ```bash
   terraform plan
   ```

4. Apply the changes:
   ```bash
   terraform apply
   ```

5. To destroy the infrastructure:
   ```bash
   terraform destroy
   ```

## Variables

| Name | Description | Type | Default |
|------|-------------|------|---------|
| aws_region | AWS region to deploy resources | string | "us-east-1" |
| environment | Environment name (e.g., dev, prod) | string | "dev" |
| s3_bucket_name | Name of the S3 bucket for storing images | string | "vending-verification-images" |
| dynamodb_table_name | Name of the DynamoDB table for verification results | string | "vending-verification-results" |
| vpc_id | ID of the VPC where resources will be deployed | string | - |
| subnet_id | ID of the subnet where the EC2 instance will be deployed | string | - |
| ami_id | AMI ID for the EC2 instance | string | "ami-0c7217cdde317cfec" |
| instance_type | EC2 instance type | string | "t2.micro" |

## Outputs

| Name | Description |
|------|-------------|
| s3_bucket_name | Name of the S3 bucket for storing images |
| dynamodb_table_name | Name of the DynamoDB table for verification results |
| app_instance_id | ID of the EC2 instance running the application |
| app_instance_public_ip | Public IP address of the EC2 instance |
| app_instance_private_ip | Private IP address of the EC2 instance |
| app_security_group_id | ID of the security group for the application |
| app_role_arn | ARN of the IAM role for the application |

## Security Considerations

- The EC2 instance is deployed in a VPC with a security group that allows inbound traffic on port 5000
- S3 bucket and DynamoDB table are created with default encryption
- IAM roles follow the principle of least privilege
- AWS Bedrock access is restricted to the specific model required

## Maintenance

- Regularly update the AMI ID to use the latest Amazon Linux 2023 AMI
- Monitor AWS Bedrock usage and costs
- Review and update security group rules as needed
- Consider implementing backup strategies for the DynamoDB table 