from aws_cdk import (
    aws_dynamodb as dynamodb,
    aws_s3 as s3,
    aws_sns as sns,
    RemovalPolicy,
    Duration,
)
from constructs import Construct

class StorageResources:
    def __init__(
        self, 
        scope: Construct, 
        resource_prefix: str,
        random_suffix: str
    ) -> None:
        self.scope = scope
        self.resource_prefix = resource_prefix
        self.random_suffix = random_suffix

        # Create S3 buckets
        self.reference_bucket = self._create_s3_bucket("reference")
        self.checking_bucket = self._create_s3_bucket("checking")
        self.results_bucket = self._create_s3_bucket("results")

        # Create DynamoDB tables
        self.verification_table = self._create_verification_table()
        self.layout_table = self._create_layout_table()

        # Create SNS topic for notifications
        self.notification_topic = self._create_notification_topic()

    def _create_s3_bucket(self, purpose):
        """Create an S3 bucket with proper settings"""
        bucket_name = f"s3-{self.resource_prefix}-{purpose}-{self.random_suffix}"
        return s3.Bucket(
            self.scope, 
            f"{self.resource_prefix}-{purpose}-bucket",
            bucket_name=bucket_name,
            removal_policy=RemovalPolicy.RETAIN,
            encryption=s3.BucketEncryption.S3_MANAGED,
            block_public_access=s3.BlockPublicAccess.BLOCK_ALL,
            versioned=True,
            lifecycle_rules=[
                s3.LifecycleRule(
                    expiration=Duration.days(365),  # Keep data for 1 year
                    transitions=[
                        s3.Transition(
                            storage_class=s3.StorageClass.INFREQUENT_ACCESS,
                            transition_after=Duration.days(30)
                        ),
                        s3.Transition(
                            storage_class=s3.StorageClass.GLACIER,
                            transition_after=Duration.days(90)
                        )
                    ]
                )
            ]
        )

    def _create_verification_table(self):
        """Create the verification results DynamoDB table"""
        table_name = f"dynamodb-{self.resource_prefix}-verification-results-{self.random_suffix}"
        table = dynamodb.Table(
            self.scope, 
            f"{self.resource_prefix}-verification-results-table",
            table_name=table_name,
            partition_key=dynamodb.Attribute(
                name="verificationId",
                type=dynamodb.AttributeType.STRING
            ),
            sort_key=dynamodb.Attribute(
                name="verificationAt",
                type=dynamodb.AttributeType.STRING
            ),
            billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST,
            removal_policy=RemovalPolicy.RETAIN,
            point_in_time_recovery=True,
            time_to_live_attribute="expirationTime"
        )

        # Add GSI for layoutId
        table.add_global_secondary_index(
            index_name="GSI1",
            partition_key=dynamodb.Attribute(
                name="layoutId",
                type=dynamodb.AttributeType.NUMBER
            ),
            sort_key=dynamodb.Attribute(
                name="verificationAt",
                type=dynamodb.AttributeType.STRING
            ),
            projection_type=dynamodb.ProjectionType.ALL
        )

        # Add GSI for verificationStatus
        table.add_global_secondary_index(
            index_name="GSI2",
            partition_key=dynamodb.Attribute(
                name="status",
                type=dynamodb.AttributeType.STRING
            ),
            sort_key=dynamodb.Attribute(
                name="verificationAt",
                type=dynamodb.AttributeType.STRING
            ),
            projection_type=dynamodb.ProjectionType.INCLUDE,
            non_key_attributes=[
                "vendingMachineId", 
                "layoutId", 
                "resultImageUrl", 
                "discrepancies"
            ]
        )

        return table

    def _create_layout_table(self):
        """Create the layout metadata DynamoDB table"""
        table_name = f"dynamodb-{self.resource_prefix}-layout-metadata-{self.random_suffix}"
        table = dynamodb.Table(
            self.scope, 
            f"{self.resource_prefix}-layout-metadata-table",
            table_name=table_name,
            partition_key=dynamodb.Attribute(
                name="layoutId",
                type=dynamodb.AttributeType.NUMBER
            ),
            sort_key=dynamodb.Attribute(
                name="layoutPrefix",
                type=dynamodb.AttributeType.STRING
            ),
            billing_mode=dynamodb.BillingMode.PAY_PER_REQUEST,
            removal_policy=RemovalPolicy.RETAIN,
            point_in_time_recovery=True
        )

        # Add GSI for createdAt
        table.add_global_secondary_index(
            index_name="GSI1",
            partition_key=dynamodb.Attribute(
                name="createdAt",
                type=dynamodb.AttributeType.STRING
            ),
            sort_key=dynamodb.Attribute(
                name="layoutId",
                type=dynamodb.AttributeType.NUMBER
            ),
            projection_type=dynamodb.ProjectionType.KEYS_ONLY
        )

        # Add GSI for vendingMachineId
        table.add_global_secondary_index(
            index_name="GSI2",
            partition_key=dynamodb.Attribute(
                name="vendingMachineId",
                type=dynamodb.AttributeType.STRING
            ),
            sort_key=dynamodb.Attribute(
                name="createdAt",
                type=dynamodb.AttributeType.STRING
            ),
            projection_type=dynamodb.ProjectionType.ALL
        )

        return table

    def _create_notification_topic(self):
        """Create SNS topic for notifications"""
        topic_name = f"sns-{self.resource_prefix}-notifications-{self.random_suffix}"
        return sns.Topic(
            self.scope, 
            f"{self.resource_prefix}-notifications-topic",
            topic_name=topic_name,
            display_name="Verification Service Notifications"
        )