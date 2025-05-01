from aws_cdk import (
    Stack,
    RemovalPolicy,
    Duration,
    aws_ecr as ecr,
    aws_secretsmanager as secretsmanager,
    aws_cloudwatch as cloudwatch,
)
from constructs import Construct
import os
import random
import string
from storage_construct import StorageConstruct
from lambda_construct import LambdaConstruct
from api_construct import ApiConstruct
from step_functions_construct import StepFunctionsConstruct

class KootoroVerificationStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, **kwargs) -> None:
        super().__init__(scope, construct_id, **kwargs)

        # Get configuration values from environment
        project_prefix = os.getenv('PROJECT_PREFIX', 'kootoro')
        stage = os.getenv('STAGE', 'dev')
        
        # Generate a random suffix for resource names (6 characters)
        random_suffix = ''.join(random.choices(string.ascii_lowercase + string.digits, k=6))
        self.resource_suffix = random_suffix
        
        # Create ECR repository for Lambda container images
        ecr_repository = ecr.Repository(
            self, 
            f"{project_prefix}-ecr-repo-{self.resource_suffix}",
            repository_name=f"{project_prefix}-verification-{stage}-{self.resource_suffix}",
            removal_policy=RemovalPolicy.RETAIN,
            image_scan_on_push=True,
            lifecycle_rules=[
                ecr.LifecycleRule(
                    description="Keep only the latest 10 images",
                    max_image_count=10,
                    rule_priority=1
                )
            ]
        )
        
        # Create storage resources (S3 buckets, DynamoDB tables)
        storage = StorageConstruct(
            self, 
            f"{project_prefix}-storage-{self.resource_suffix}",
            project_prefix=project_prefix,
            stage=stage,
            resource_suffix=self.resource_suffix
        )
        
        # Create Secrets Manager for Bedrock credentials
        bedrock_secret = secretsmanager.Secret(
            self, 
            f"{project_prefix}-bedrock-secret-{self.resource_suffix}",
            secret_name=f"{project_prefix}/bedrock-api-key/{stage}-{self.resource_suffix}",
            description="API key for Amazon Bedrock"
        )
        
        # Create Lambda functions
        lambda_construct = LambdaConstruct(
            self, 
            f"{project_prefix}-lambda-{self.resource_suffix}",
            project_prefix=project_prefix,
            stage=stage,
            storage=storage,
            ecr_repository=ecr_repository,
            bedrock_secret=bedrock_secret,
            resource_suffix=self.resource_suffix
        )
        
        # Create Step Functions workflow
        step_functions = StepFunctionsConstruct(
            self, 
            f"{project_prefix}-step-functions-{self.resource_suffix}",
            project_prefix=project_prefix,
            stage=stage,
            lambda_functions=lambda_construct.functions,
            resource_suffix=self.resource_suffix
        )
        
        # Create API Gateway
        api = ApiConstruct(
            self, 
            f"{project_prefix}-api-{self.resource_suffix}",
            project_prefix=project_prefix,
            stage=stage,
            lambda_functions=lambda_construct.functions,
            state_machine=step_functions.state_machine,
            resource_suffix=self.resource_suffix
        )
        
        # CloudWatch dashboard for monitoring
        dashboard = cloudwatch.Dashboard(
            self, 
            f"{project_prefix}-dashboard-{self.resource_suffix}",
            dashboard_name=f"{project_prefix}-verification-{stage}-{self.resource_suffix}"
        )
        
        # Add widgets to the dashboard
        dashboard.add_widgets(
            cloudwatch.GraphWidget(
                title="Lambda Execution Duration",
                left=[
                    lambda_function.metric_duration()
                    for lambda_function in lambda_construct.functions.values()
                ][:5]  # Limit to 5 functions for readability
            ),
            cloudwatch.GraphWidget(
                title="Step Functions Execution Time",
                left=[step_functions.state_machine.metric_time()]
            ),
            cloudwatch.GraphWidget(
                title="API Gateway Latency",
                left=[api.api.metric_latency()]
            ),
            cloudwatch.GraphWidget(
                title="Lambda Errors",
                left=[
                    lambda_function.metric_errors()
                    for lambda_function in lambda_construct.functions.values()
                ][:5]  # Limit to 5 functions
            )
        )