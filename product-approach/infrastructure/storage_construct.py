from aws_cdk import (
    aws_s3 as s3,
    aws_dynamodb as dynamodb,
    RemovalPolicy,
)
from constructs import Construct

class StorageConstruct(Construct):
    def __init__(self, scope: Construct, id: str, project_prefix: str, stage: str, resource_suffix: str, **kwargs):
        super().__init__(scope, id, **kwargs)

        # Reference Bucket: Stores reference layout images and JSON configuration files
        self.reference_bucket = s3.Bucket(
            self,
            f"{project_prefix}-reference-bucket-{stage}-{resource_suffix}",
            bucket_name=f"{project_prefix}-reference-{stage}-{resource_suffix}",
            encryption=s3.BucketEncryption.S3_MANAGED,
            versioned=True,
            removal_policy=RemovalPolicy.RETAIN,
        )

        # Checking Bucket: Stores checking images uploaded by field employees
        self.checking_bucket = s3.Bucket(
            self,
            f"{project_prefix}-checking-bucket-{stage}-{resource_suffix}",
            bucket_name=f"{project_prefix}-checking-{stage}-{resource_suffix}",
            encryption=s3.BucketEncryption.S3_MANAGED,
            versioned=True,
            removal_policy=RemovalPolicy.RETAIN,
        )

        # Results Bucket: Stores processed results and visualizations
        self.results_bucket = s3.Bucket(
            self,
            f"{project_prefix}-results-bucket-{stage}-{resource_suffix}",
            bucket_name=f"{project_prefix}-results-{stage}-{resource_suffix}",
            encryption=s3.BucketEncryption.S3_MANAGED,
            versioned=True,
            removal_policy=RemovalPolicy.RETAIN,
        )

        # DynamoDB Tables
        self.verification_results_table = dynamodb.Table(
            self,
            f"{project_prefix}-verification-results-table-{stage}-{resource_suffix}",
            table_name=f"{project_prefix}-verification-results-{stage}-{resource_suffix}",
            partition_key=dynamodb.Attribute(name="verificationId", type=dynamodb.AttributeType.STRING),
            sort_key=dynamodb.Attribute(name="verificationAt", type=dynamodb.AttributeType.STRING),
            billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST,
            removal_policy=RemovalPolicy.RETAIN,
        )

        self.conversation_history_table = dynamodb.Table(
            self,
            f"{project_prefix}-conversation-history-table-{stage}-{resource_suffix}",
            table_name=f"{project_prefix}-conversation-history-{stage}-{resource_suffix}",
            partition_key=dynamodb.Attribute(name="verificationId", type=dynamodb.AttributeType.STRING),
            sort_key=dynamodb.Attribute(name="conversationAt", type=dynamodb.AttributeType.STRING),
            billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST,
            removal_policy=RemovalPolicy.RETAIN,
        )

        self.layout_metadata_table = dynamodb.Table(
            self,
            f"{project_prefix}-layout-metadata-table-{stage}-{resource_suffix}",
            table_name=f"{project_prefix}-layout-metadata-{stage}-{resource_suffix}",
            partition_key=dynamodb.Attribute(name="layoutId", type=dynamodb.AttributeType.NUMBER),
            sort_key=dynamodb.Attribute(name="layoutPrefix", type=dynamodb.AttributeType.STRING),
            billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST,
            removal_policy=RemovalPolicy.RETAIN,
        )