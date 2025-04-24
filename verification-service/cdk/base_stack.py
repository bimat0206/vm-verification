from aws_cdk import (
    Stack,
    RemovalPolicy,
    Duration,
    CfnOutput,
)
from constructs import Construct
import string
import random

from storage_resources import StorageResources
from compute_resources import ComputeResources

class VerificationServiceStack(Stack):
    def __init__(
        self, 
        scope: Construct, 
        id: str, 
        resource_prefix: str = "verification",
        random_suffix: str = None,
        **kwargs
    ) -> None:
        super().__init__(scope, id, **kwargs)

        # Store the resource prefix for naming consistency
        self.resource_prefix = resource_prefix
        
        # Use provided random suffix or generate one if not provided
        self.random_suffix = random_suffix if random_suffix else self._generate_random_suffix()

        # Create storage resources (S3, DynamoDB, SNS)
        storage = StorageResources(self, self.resource_prefix, self.random_suffix)
        self.reference_bucket = storage.reference_bucket
        self.checking_bucket = storage.checking_bucket
        self.results_bucket = storage.results_bucket
        self.verification_table = storage.verification_table
        self.layout_table = storage.layout_table
        self.notification_topic = storage.notification_topic

        # Create compute resources (VPC, ECS, IAM roles, etc.)
        compute = ComputeResources(
            self, 
            self.resource_prefix, 
            self.random_suffix,
            storage.reference_bucket,
            storage.checking_bucket,
            storage.results_bucket,
            storage.verification_table,
            storage.layout_table,
            storage.notification_topic
        )
        self.vpc = compute.vpc
        self.ecs_cluster = compute.ecs_cluster
        self.execution_role = compute.execution_role
        self.task_role = compute.task_role
        self.ecr_repository = compute.ecr_repository
        self.log_group = compute.log_group
        self.task_definition = compute.task_definition
        self.service = compute.service
        self.load_balancer = compute.load_balancer

        # Output the resource names and ARNs
        self._create_outputs()
    
    def _generate_random_suffix(self, length=8):
        """Generate a random alphanumeric suffix for resource names"""
        # Create a random string of lowercase letters and numbers
        chars = string.ascii_lowercase + string.digits
        return ''.join(random.choice(chars) for _ in range(length))

    def _create_outputs(self):
        """Create stack outputs"""
        CfnOutput(
            self, "RandomSuffix",
            value=self.random_suffix,
            description="Random suffix used for all resources"
        )
        
        CfnOutput(
            self, "VerificationTableName",
            value=self.verification_table.table_name,
            description="DynamoDB table for verification results"
        )

        CfnOutput(
            self, "LayoutTableName",
            value=self.layout_table.table_name,
            description="DynamoDB table for layout metadata"
        )

        CfnOutput(
            self, "ReferenceBucketName",
            value=self.reference_bucket.bucket_name,
            description="S3 bucket for reference images"
        )

        CfnOutput(
            self, "CheckingBucketName",
            value=self.checking_bucket.bucket_name,
            description="S3 bucket for checking images"
        )

        CfnOutput(
            self, "ResultsBucketName",
            value=self.results_bucket.bucket_name,
            description="S3 bucket for results"
        )

        CfnOutput(
            self, "NotificationTopicArn",
            value=self.notification_topic.topic_arn,
            description="SNS topic for notifications"
        )

        CfnOutput(
            self, "ECRRepositoryUri",
            value=self.ecr_repository.repository_uri,
            description="ECR repository for service images"
        )

        CfnOutput(
            self, "LoadBalancerDnsName",
            value=self.load_balancer.load_balancer_dns_name,
            description="ALB DNS name"
        )