from aws_cdk import (
    aws_s3 as s3,
    aws_dynamodb as dynamodb,
    aws_sns as sns,
)
from constructs import Construct

from network_resources import NetworkResources
from container_resources import ContainerResources

class ComputeResources:
    def __init__(
        self, 
        scope: Construct, 
        resource_prefix: str,
        random_suffix: str,
        reference_bucket: s3.Bucket,
        checking_bucket: s3.Bucket,
        results_bucket: s3.Bucket,
        verification_table: dynamodb.Table,
        layout_table: dynamodb.Table,
        notification_topic: sns.Topic
    ) -> None:
        self.scope = scope
        self.resource_prefix = resource_prefix
        self.random_suffix = random_suffix
        
        # Store references to storage resources
        self.reference_bucket = reference_bucket
        self.checking_bucket = checking_bucket
        self.results_bucket = results_bucket
        self.verification_table = verification_table
        self.layout_table = layout_table
        self.notification_topic = notification_topic

        # Create network resources
        network = NetworkResources(
            self.scope,
            self.resource_prefix,
            self.random_suffix
        )
        self.vpc = network.vpc
        self.load_balancer = network.load_balancer
        
        # Create container resources
        container = ContainerResources(
            self.scope,
            self.resource_prefix,
            self.random_suffix,
            network.vpc,
            network.service_security_group,
            network.target_group,
            self.reference_bucket,
            self.checking_bucket,
            self.results_bucket,
            self.verification_table,
            self.layout_table,
            self.notification_topic
        )
        self.ecs_cluster = container.ecs_cluster
        self.execution_role = container.execution_role
        self.task_role = container.task_role
        self.ecr_repository = container.ecr_repository
        self.log_group = container.log_group
        self.task_definition = container.task_definition
        self.service = container.service