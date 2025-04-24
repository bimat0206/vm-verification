from aws_cdk import (
    Stack,
    aws_dynamodb as dynamodb,
    aws_s3 as s3,
    aws_ecs as ecs,
    aws_ec2 as ec2,
    aws_ecr as ecr,
    aws_iam as iam,
    aws_elasticloadbalancingv2 as elbv2,
    aws_logs as logs,
    aws_sns as sns,
    RemovalPolicy,
    Duration,
    CfnOutput,
)
from constructs import Construct
import string
import random

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

        # Create S3 buckets
        self.reference_bucket = self._create_s3_bucket("reference")
        self.checking_bucket = self._create_s3_bucket("checking")
        self.results_bucket = self._create_s3_bucket("results")

        # Create DynamoDB tables
        self.verification_table = self._create_verification_table()
        self.layout_table = self._create_layout_table()

        # Create SNS topic for notifications
        self.notification_topic = self._create_notification_topic()

        # Create VPC
        self.vpc = self._create_vpc()

        # Create ECS Cluster
        self.ecs_cluster = self._create_ecs_cluster()

        # Create IAM role for ECS task execution
        self.execution_role = self._create_execution_role()
        self.task_role = self._create_task_role()

        # Create ECR Repository
        self.ecr_repository = self._create_ecr_repository()

        # Create log group
        self.log_group = self._create_log_group()

        # Create ECS Task Definition
        self.task_definition = self._create_task_definition()

        # Create ECS Service with Load Balancer
        self.service, self.load_balancer = self._create_service_with_alb()

        # Output the resource names and ARNs
        self._create_outputs()
    
    def _generate_random_suffix(self, length=8):
        """Generate a random alphanumeric suffix for resource names"""
        # Create a random string of lowercase letters and numbers
        chars = string.ascii_lowercase + string.digits
        return ''.join(random.choice(chars) for _ in range(length))

    def _create_s3_bucket(self, purpose):
        """Create an S3 bucket with proper settings"""
        bucket_name = f"s3-{self.resource_prefix}-{purpose}-{self.random_suffix}"
        return s3.Bucket(
            self, 
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
            self, 
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
            self, 
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
            self, 
            f"{self.resource_prefix}-notifications-topic",
            topic_name=topic_name,
            display_name="Verification Service Notifications"
        )

    def _create_vpc(self):
        """Create VPC for the application"""
        return ec2.Vpc(
            self, 
            f"{self.resource_prefix}-vpc",
            max_azs=2,
            nat_gateways=1,
            subnet_configuration=[
                ec2.SubnetConfiguration(
                    name=f"public-{self.resource_prefix}-{self.random_suffix}",
                    subnet_type=ec2.SubnetType.PUBLIC,
                    cidr_mask=24
                ),
                ec2.SubnetConfiguration(
                    name=f"private-{self.resource_prefix}-{self.random_suffix}",
                    subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS,
                    cidr_mask=24
                )
            ]
        )

    def _create_ecs_cluster(self):
        """Create ECS cluster"""
        cluster_name = f"ecs-{self.resource_prefix}-cluster-{self.random_suffix}"
        return ecs.Cluster(
            self, 
            f"{self.resource_prefix}-cluster",
            vpc=self.vpc,
            cluster_name=cluster_name
        )

    def _create_execution_role(self):
        """Create IAM role for ECS task execution"""
        role_name = f"iam-{self.resource_prefix}-execution-role-{self.random_suffix}"
        role = iam.Role(
            self, 
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
            self, 
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
            self, 
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
            self, 
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
            self, 
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
                "AWS_REGION": Stack.of(self).region,
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

# In the _create_service_with_alb method, change the security group naming pattern:

    def _create_service_with_alb(self):
        """Create ECS Service with Application Load Balancer"""
        # Create security group for the load balancer
        lb_sg_name = f"secgrp-{self.resource_prefix}-lb-{self.random_suffix}"
        lb_security_group = ec2.SecurityGroup(
            self, 
            f"{self.resource_prefix}-lb-sg",
            vpc=self.vpc,
            security_group_name=lb_sg_name,
            description="Security group for verification service load balancer",
            allow_all_outbound=True
        )
        lb_security_group.add_ingress_rule(
            ec2.Peer.any_ipv4(),
            ec2.Port.tcp(80),
            "Allow HTTP traffic"
        )
        lb_security_group.add_ingress_rule(
            ec2.Peer.any_ipv4(),
            ec2.Port.tcp(443),
            "Allow HTTPS traffic"
        )

        # Create security group for the service
        service_sg_name = f"secgrp-{self.resource_prefix}-service-{self.random_suffix}"
        service_security_group = ec2.SecurityGroup(
            self, 
            f"{self.resource_prefix}-service-sg",
            vpc=self.vpc,
            security_group_name=service_sg_name,
            description="Security group for verification service",
            allow_all_outbound=True
        )
        service_security_group.add_ingress_rule(
            lb_security_group,
            ec2.Port.tcp(3000),
            "Allow traffic from ALB"
        )
        
        # Create load balancer
        lb_name = f"alb-{self.resource_prefix}-{self.random_suffix}"
        # Ensure the ALB name is not longer than 32 characters (AWS limit)
        if len(lb_name) > 32:
            lb_name = lb_name[:32-len(self.random_suffix)-1] + "-" + self.random_suffix
            
        load_balancer = elbv2.ApplicationLoadBalancer(
            self, 
            f"{self.resource_prefix}-alb",
            vpc=self.vpc,
            internet_facing=True,
            security_group=lb_security_group,
            load_balancer_name=lb_name
        )

        # Create listener
        listener = load_balancer.add_listener(
            f"{self.resource_prefix}-http-listener",
            port=80,
            open=True
        )

        # Create target group
        target_group_name = f"tg-{self.resource_prefix}-{self.random_suffix}"
        # Ensure target group name is not longer than 32 characters (AWS limit)
        if len(target_group_name) > 32:
            target_group_name = target_group_name[:32-len(self.random_suffix)-1] + "-" + self.random_suffix
            
        # Create target group separately
        target_group = elbv2.ApplicationTargetGroup(
            self,
            f"{self.resource_prefix}-target-group",
            port=3000,
            protocol=elbv2.ApplicationProtocol.HTTP,
            target_type=elbv2.TargetType.IP,
            target_group_name=target_group_name,
            vpc=self.vpc,
            health_check=elbv2.HealthCheck(
                path="/health",
                interval=Duration.seconds(30),
                timeout=Duration.seconds(5),
                healthy_http_codes="200"
            )
        )
        
        # Add target group to listener
        listener.add_target_groups(
            f"{self.resource_prefix}-targets",
            target_groups=[target_group]
        )

        # Create ECS Service with desired count 0 (no tasks will be deployed)
        service_name = f"svc-{self.resource_prefix}-{self.random_suffix}"
        service = ecs.FargateService(
            self, 
            f"{self.resource_prefix}-service",
            cluster=self.ecs_cluster,
            task_definition=self.task_definition,
            desired_count=0,  # Set to 0 to not deploy any tasks
            security_groups=[service_security_group],
            vpc_subnets=ec2.SubnetSelection(
                subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS
            ),
            service_name=service_name,
            assign_public_ip=False,
            health_check_grace_period=Duration.seconds(60)
        )

        # Attach service to target group
        service.attach_to_application_target_group(target_group)

        return service, load_balancer

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