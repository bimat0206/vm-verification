from aws_cdk import (
    aws_ecs as ecs,
    aws_ecr as ecr,
    aws_iam as iam,
    aws_ec2 as ec2,
    aws_elasticloadbalancingv2 as elbv2,
    aws_logs as logs,
    RemovalPolicy,
    Duration,
    Stack,
    CfnOutput,
)
from constructs import Construct

class StreamlitFrontendResources:
    def __init__(
        self, 
        scope: Construct, 
        resource_prefix: str,
        random_suffix: str,
        vpc: ec2.Vpc,
        ecs_cluster: ecs.Cluster,
        load_balancer: elbv2.ApplicationLoadBalancer,
        backend_service_url: str,
    ) -> None:
        self.scope = scope
        self.resource_prefix = resource_prefix
        self.random_suffix = random_suffix
        self.vpc = vpc
        self.ecs_cluster = ecs_cluster
        self.load_balancer = load_balancer
        self.backend_service_url = backend_service_url
        
        # Create ECR Repository
        self.ecr_repository = self._create_ecr_repository()
        
        # Create IAM roles
        self.execution_role = self._create_execution_role()
        self.task_role = self._create_task_role()
        
        # Create log group
        self.log_group = self._create_log_group()
        
        # Create security group
        self.security_group = self._create_security_group()
        
        # Create target group and listener rule
        self.target_group = self._create_target_group()
        self.listener_rule = self._create_listener_rule()
        
        # Create ECS Task Definition
        self.task_definition = self._create_task_definition()
        
        # Create ECS Service
        self.service = self._create_service()
        
        # Output the ECR repository URI
        self._create_outputs()

    def _create_ecr_repository(self):
        """Create ECR Repository for the Streamlit frontend"""
        repository_name = f"ecr-{self.resource_prefix}-streamlit-{self.random_suffix}"
        return ecr.Repository(
            self.scope, 
            f"{self.resource_prefix}-streamlit-repository",
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

    def _create_execution_role(self):
        """Create IAM role for ECS task execution"""
        role_name = f"iam-{self.resource_prefix}-streamlit-exec-{self.random_suffix}"
        role = iam.Role(
            self.scope, 
            f"{self.resource_prefix}-streamlit-execution-role",
            role_name=role_name,
            assumed_by=iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
            managed_policies=[
                iam.ManagedPolicy.from_aws_managed_policy_name("service-role/AmazonECSTaskExecutionRolePolicy")
            ]
        )
        return role

    def _create_task_role(self):
        """Create IAM role for ECS task"""
        role_name = f"iam-{self.resource_prefix}-streamlit-task-{self.random_suffix}"
        role = iam.Role(
            self.scope, 
            f"{self.resource_prefix}-streamlit-task-role",
            role_name=role_name,
            assumed_by=iam.ServicePrincipal("ecs-tasks.amazonaws.com")
        )
        return role

    def _create_log_group(self):
        """Create CloudWatch log group"""
        log_group_name = f"/ecs/logs-{self.resource_prefix}-streamlit-{self.random_suffix}"
        return logs.LogGroup(
            self.scope, 
            f"{self.resource_prefix}-streamlit-logs",
            log_group_name=log_group_name,
            removal_policy=RemovalPolicy.DESTROY,
            retention=logs.RetentionDays.ONE_MONTH
        )
    
    def _create_security_group(self):
        """Create security group for the Streamlit service"""
        sg_name = f"secgrp-{self.resource_prefix}-streamlit-{self.random_suffix}"
        security_group = ec2.SecurityGroup(
            self.scope, 
            f"{self.resource_prefix}-streamlit-sg",
            vpc=self.vpc,
            security_group_name=sg_name,
            description="Security group for Streamlit frontend service",
            allow_all_outbound=True
        )
        
        # Add ingress rule to allow traffic from the ALB
        security_group.add_ingress_rule(
            ec2.Peer.any_ipv4(),
            ec2.Port.tcp(8501),
            "Allow traffic to Streamlit port"
        )
        
        return security_group
    
    def _create_target_group(self):
        """Create target group for the Streamlit service"""
        target_group_name = f"tg-{self.resource_prefix}-streamlit-{self.random_suffix}"
        # Ensure target group name is not longer than 32 characters (AWS limit)
        if len(target_group_name) > 32:
            target_group_name = target_group_name[:32-len(self.random_suffix)-1] + "-" + self.random_suffix
            
        return elbv2.ApplicationTargetGroup(
            self.scope,
            f"{self.resource_prefix}-streamlit-target-group",
            port=8501,
            protocol=elbv2.ApplicationProtocol.HTTP,
            target_type=elbv2.TargetType.IP,
            target_group_name=target_group_name,
            vpc=self.vpc,
            health_check=elbv2.HealthCheck(
                path="/",
                interval=Duration.seconds(60),
                timeout=Duration.seconds(30),
                healthy_http_codes="200",
                healthy_threshold_count=2,
                unhealthy_threshold_count=5
            )
        )
    
    def _create_listener_rule(self):
        """Create listener rule to route /ui path to Streamlit"""
        # Find the HTTP listener in the load balancer
        http_listener = None
        for listener in self.load_balancer.listeners:
            if listener.connections.default_port == 80:
                http_listener = listener
                break
        
        if not http_listener:
            # If no HTTP listener found, create one
            http_listener = self.load_balancer.add_listener(
                f"{self.resource_prefix}-http-listener",
                port=80,
                open=True
            )
        
        # Create listener rule for '/ui' path
        return http_listener.add_action(
            f"{self.resource_prefix}-streamlit-action",
            action=elbv2.ListenerAction.forward([self.target_group]),
            conditions=[
                elbv2.ListenerCondition.path_patterns(["/ui/*"])
            ],
            priority=10  # Lower priority numbers are evaluated first
        )
    
    def _create_task_definition(self):
        """Create ECS Fargate task definition for Streamlit"""
        container_name = f"container-{self.resource_prefix}-streamlit-{self.random_suffix}"
        family_name = f"task-{self.resource_prefix}-streamlit-{self.random_suffix}"
        
        task_definition = ecs.FargateTaskDefinition(
            self.scope,
            f"{self.resource_prefix}-streamlit-task-definition",
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
                "STREAMLIT_SERVER_PORT": "8501",
                "STREAMLIT_SERVER_HEADLESS": "true",
                "STREAMLIT_SERVER_ENABLE_CORS": "true",
                "STREAMLIT_SERVER_ENABLE_XSRF_PROTECTION": "true",
                "STREAMLIT_BROWSER_GATHER_USAGE_STATS": "false",
                "STREAMLIT_SERVER_BASE_URL_PATH": "/ui",
                "API_URL": self.backend_service_url
            },
            logging=ecs.LogDrivers.aws_logs(
                stream_prefix=self.resource_prefix,
                log_group=self.log_group
            ),
            port_mappings=[ecs.PortMapping(container_port=8501)]
        )
        
        return task_definition
    
    def _create_service(self):
        """Create ECS Service for Streamlit"""
        service_name = f"svc-{self.resource_prefix}-streamlit-{self.random_suffix}"
        service = ecs.FargateService(
            self.scope,
            f"{self.resource_prefix}-streamlit-service",
            cluster=self.ecs_cluster,
            task_definition=self.task_definition,
            desired_count=0,
            security_groups=[self.security_group],
            vpc_subnets=ec2.SubnetSelection(
                subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS
            ),
            service_name=service_name,
            assign_public_ip=False,
            health_check_grace_period=Duration.seconds(120)
        )
        
        # Attach service to target group
        service.attach_to_application_target_group(self.target_group)
        
        return service
    
    def _create_outputs(self):
        """Create outputs"""
        CfnOutput(
            self.scope, 
            "StreamlitRepositoryUri",
            value=self.ecr_repository.repository_uri,
            description="ECR repository URI for Streamlit frontend"
        )
        
        streamlit_url = f"http://{self.load_balancer.load_balancer_dns_name}/ui"
        CfnOutput(
            self.scope, 
            "StreamlitUrl",
            value=streamlit_url,
            description="URL for accessing the Streamlit frontend"
        )