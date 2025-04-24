from aws_cdk import (
    aws_ec2 as ec2,
    aws_elasticloadbalancingv2 as elbv2,
    Duration,
)
from constructs import Construct

class NetworkResources:
    def __init__(
        self, 
        scope: Construct, 
        resource_prefix: str,
        random_suffix: str
    ) -> None:
        self.scope = scope
        self.resource_prefix = resource_prefix
        self.random_suffix = random_suffix

        # Create VPC
        self.vpc = self._create_vpc()
        
        # Create security groups
        self.lb_security_group = self._create_lb_security_group()
        self.service_security_group = self._create_service_security_group()
        
        # Create load balancer
        self.load_balancer = self._create_load_balancer()
        
        # Create listener
        self.listener = self._create_listener()
        
        # Create target group
        self.target_group = self._create_target_group()

    def _create_vpc(self):
        """Create VPC for the application"""
        return ec2.Vpc(
            self.scope, 
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

    def _create_lb_security_group(self):
        """Create security group for the load balancer"""
        lb_sg_name = f"secgrp-{self.resource_prefix}-lb-{self.random_suffix}"
        security_group = ec2.SecurityGroup(
            self.scope, 
            f"{self.resource_prefix}-lb-sg",
            vpc=self.vpc,
            security_group_name=lb_sg_name,
            description="Security group for verification service load balancer",
            allow_all_outbound=True
        )
        security_group.add_ingress_rule(
            ec2.Peer.any_ipv4(),
            ec2.Port.tcp(80),
            "Allow HTTP traffic"
        )
        security_group.add_ingress_rule(
            ec2.Peer.any_ipv4(),
            ec2.Port.tcp(443),
            "Allow HTTPS traffic"
        )
        return security_group

    def _create_service_security_group(self):
        """Create security group for the service"""
        service_sg_name = f"secgrp-{self.resource_prefix}-service-{self.random_suffix}"
        security_group = ec2.SecurityGroup(
            self.scope, 
            f"{self.resource_prefix}-service-sg",
            vpc=self.vpc,
            security_group_name=service_sg_name,
            description="Security group for verification service",
            allow_all_outbound=True
        )
        security_group.add_ingress_rule(
            self.lb_security_group,
            ec2.Port.tcp(3000),
            "Allow traffic from ALB"
        )
        return security_group

    def _create_load_balancer(self):
        """Create load balancer"""
        lb_name = f"alb-{self.resource_prefix}-{self.random_suffix}"
        # Ensure the ALB name is not longer than 32 characters (AWS limit)
        if len(lb_name) > 32:
            lb_name = lb_name[:32-len(self.random_suffix)-1] + "-" + self.random_suffix
            
        return elbv2.ApplicationLoadBalancer(
            self.scope, 
            f"{self.resource_prefix}-alb",
            vpc=self.vpc,
            internet_facing=True,
            security_group=self.lb_security_group,
            load_balancer_name=lb_name
        )

    def _create_listener(self):
        """Create listener for the load balancer"""
        return self.load_balancer.add_listener(
            f"{self.resource_prefix}-http-listener",
            port=80,
            open=True
        )

    def _create_target_group(self):
        """Create target group for the load balancer"""
        target_group_name = f"tg-{self.resource_prefix}-{self.random_suffix}"
        # Ensure target group name is not longer than 32 characters (AWS limit)
        if len(target_group_name) > 32:
            target_group_name = target_group_name[:32-len(self.random_suffix)-1] + "-" + self.random_suffix
            
        target_group = elbv2.ApplicationTargetGroup(
            self.scope,
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
        self.listener.add_target_groups(
            f"{self.resource_prefix}-targets",
            target_groups=[target_group]
        )
        
        return target_group