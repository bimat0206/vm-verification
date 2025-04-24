from aws_cdk import (
    Stack,
    aws_ecs as ecs,
    aws_iam as iam,
    aws_ecr as ecr,
    aws_logs as logs,
    aws_s3 as s3,
    aws_dynamodb as dynamodb,
    aws_sns as sns,
    aws_ec2 as ec2,
    aws_elasticloadbalancingv2 as elbv2,
    RemovalPolicy,
    Duration,
)
from constructs import Construct

class ContainerResources:
    def __init__(
        self, 
        scope: Construct, 
        resource_prefix: str,
        random_suffix: str,
        vpc: ec2.Vpc,
        service_security_group: ec2.SecurityGroup,
        target_group: elbv2.ApplicationTargetGroup,
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
        
        # Network resources
        self.vpc = vpc
        self.service_security_group = service_security_group
        self.target_group = target_group
        
        # Storage resources
        self.reference_bucket = reference_bucket
        self.checking_bucket = checking_bucket
        self.results_bucket = results_bucket
        self.verification_table = verification_table
        self.layout_table = layout_table
        self.notification_topic = notification_topic

        # Create ECS Cluster
        self.ecs_cluster = self._create_ecs_cluster()

        # Create IAM roles
        self.execution_role = self._create_execution_role()
        self.task_role = self._create_task_role()

        # Create ECR Repository
        self.ecr_repository = self._create_ecr_repository()

        # Create log group
        self.log_group = self._create_log_group()

        # Create ECS Task Definition
        self.task_definition = self._create_task_definition()

        # Create ECS Service
        self.service = self._create_service()

    def _create_ecs_cluster(self):
        """Create ECS cluster"""
        cluster_name = f"ecs-{self.resource_prefix}-cluster-{self.random_suffix}"
        return ecs.Cluster(
            self.scope, 
            f"{self.resource_prefix}-cluster",
            vpc=self.vpc,
            cluster_name=cluster_name
        )

    def _create_execution_role(self):
        """Create IAM role for ECS task execution"""
        role_name = f"iam-{self.resource_prefix}-execution-role-{self.random_suffix}"
        role = iam.Role(
            self.scope, 
            f"{self.resource_prefix}-execution-role",
            role_name=role_name,
            assumed_by=iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
            managed_policies=[
                iam.ManagedPolicy.from_aws_managed_policy_name("service-role/AmazonECSTaskExecutionRolePolicy")
            ]
        )
        return role

    def _create_task_role(self):
        """Create IAM role for ECS task"""
        role_name = f"iam-{self.resource_prefix}-task-role-{self.random_suffix}"
        role = iam.Role(
            self.scope, 
            f"{self.resource_prefix}-task-role",
            role_name=role_name,
            assumed_by=iam.ServicePrincipal("ecs-tasks.amazonaws.com")
        )

        # Add permissions to S3 buckets
        role.add_to_policy(iam.PolicyStatement(
            actions=[
                "s3:GetObject",
                "s3:PutObject",
                "s3:ListBucket"
            ],
            resources=[
                self.reference_bucket.bucket_arn,
                f"{self.reference_bucket.bucket_arn}/*",
                self.checking_bucket.bucket_arn,
                f"{self.checking_bucket.bucket_arn}/*",
                self.results_bucket.bucket_arn,
                f"{self.results_bucket.bucket_arn}/*"
            ]
        ))

        # Add permissions to DynamoDB tables
        role.add_to_policy(iam.PolicyStatement(
            actions=[
                "dynamodb:GetItem",
                "dynamodb:PutItem",
                "dynamodb:UpdateItem",
                "dynamodb:DeleteItem",
                "dynamodb:Query",
                "dynamodb:Scan"
            ],
            resources=[
                self.verification_table.table_arn,
                f"{self.verification_table.table_arn}/index/*",
                self.layout_table.table_arn,
                f"{self.layout_table.table_arn}/index/*"
            ]
        ))

        # Add permissions to Bedrock
        role.add_to_policy(iam.PolicyStatement(
            actions=[
                "bedrock:InvokeModel",
                "bedrock:InvokeModelWithResponseStream"
            ],
            resources=["*"]  # Consider restricting to specific model ARNs
        ))

        # Add permissions to SNS
        role.add_to_policy(iam.PolicyStatement(
            actions=[
                "sns:Publish"
            ],
            resources=[self.notification_topic.topic_arn]
        ))

        return role

    def _create_ecr_repository(self):
        """Create ECR Repository for the service"""
        repository_name = f"ecr-{self.resource_prefix}-service-{self.random_suffix}"
        return ecr.Repository(
            self.scope, 
            f"{self.resource_prefix}-repository",
            repository_name=repository_name,
            removal_policy=RemovalPolicy.RETAIN,
            lifecycle_rules=[
                ecr.LifecycleRule(
                    description="Keep only the last 5 images",
                    max_image_count=5,
                    rule_priority=1
                )
            ]
        )

    def _create_log_group(self):
        """Create CloudWatch log group"""
        log_group_name = f"/ecs/logs-{self.resource_prefix}-service-{self.random_suffix}"
        return logs.LogGroup(
            self.scope, 
            f"{self.resource_prefix}-logs",
            log_group_name=log_group_name,
            removal_policy=RemovalPolicy.DESTROY,
            retention=logs.RetentionDays.ONE_MONTH
        )

    def _create_task_definition(self):
        """Create ECS Fargate task definition"""
        container_name = f"container-{self.resource_prefix}-{self.random_suffix}"
        family_name = f"task-{self.resource_prefix}-service-{self.random_suffix}"
        
        task_definition = ecs.FargateTaskDefinition(
            self.scope, 
            f"{self.resource_prefix}-task-definition",
            family=family_name,
            memory_limit_mib=2048,
            cpu=1024,
            execution_role=self.execution_role,
            task_role=self.task_role
        )

        container = task_definition.add_container(
            container_name,
            image=ecs.ContainerImage.from_ecr_repository(self.ecr_repository),
            essential=True,
            environment={
                "SERVER_PORT": "3000",
                "SERVER_READ_TIMEOUT_SECS": "30",
                "SERVER_WRITE_TIMEOUT_SECS": "30",
                "SERVER_IDLE_TIMEOUT_SECS": "60",
                "AWS_REGION": Stack.of(self.scope).region,
                "DYNAMODB_VERIFICATION_TABLE": self.verification_table.table_name,
                "DYNAMODB_LAYOUT_TABLE": self.layout_table.table_name,
                "S3_REFERENCE_BUCKET": self.reference_bucket.bucket_name,
                "S3_CHECKING_BUCKET": self.checking_bucket.bucket_name,
                "S3_RESULTS_BUCKET": self.results_bucket.bucket_name,
                "BEDROCK_MODEL_ID": "anthropic.claude-3-7-sonnet-20250219",
                "BEDROCK_MAX_RETRIES": "3",
                "NOTIFICATION_SNS_TOPIC_ARN": self.notification_topic.topic_arn,
                "LOG_LEVEL": "INFO"
            },
            logging=ecs.LogDrivers.aws_logs(
                stream_prefix=self.resource_prefix,
                log_group=self.log_group
            ),
            port_mappings=[ecs.PortMapping(container_port=3000)]
        )

        return task_definition

    def _create_service(self):
        """Create ECS Service"""
        service_name = f"svc-{self.resource_prefix}-{self.random_suffix}"
        service = ecs.FargateService(
            self.scope, 
            f"{self.resource_prefix}-service",
            cluster=self.ecs_cluster,
            task_definition=self.task_definition,
            desired_count=0,  # Set to 0 to not deploy any tasks
            security_groups=[self.service_security_group],
            vpc_subnets=ec2.SubnetSelection(
                subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS
            ),
            service_name=service_name,
            assign_public_ip=False,
            health_check_grace_period=Duration.seconds(60)
        )

        # Attach service to target group
        service.attach_to_application_target_group(self.target_group)

        return service