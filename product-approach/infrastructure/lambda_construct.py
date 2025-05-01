from aws_cdk import (
    Duration,
    aws_lambda as lambda_,
    aws_ecr as ecr,
    aws_iam as iam,
    aws_logs as logs,
)
from constructs import Construct
import os

class LambdaConstruct(Construct):
    """
    Creates Lambda functions for the Kootoro verification workflow.
    All functions are deployed using container images for consistency.
    For first deployment, a basic nginx container is used.
    """
    
    def __init__(
        self, 
        scope: Construct, 
        id: str, 
        project_prefix: str, 
        stage: str, 
        storage, 
        ecr_repository,
        bedrock_secret,
        resource_suffix: str,
        **kwargs
    ) -> None:
        super().__init__(scope, id, **kwargs)
        self.resource_suffix = resource_suffix

        self.functions = {}

        # Common Lambda environment variables
        common_env = {
            "PROJECT_PREFIX": project_prefix,
            "STAGE": stage,
            "REFERENCE_BUCKET": storage.reference_bucket.bucket_name,
            "CHECKING_BUCKET": storage.checking_bucket.bucket_name,
            "RESULTS_BUCKET": storage.results_bucket.bucket_name,
            "VERIFICATION_RESULTS_TABLE": storage.verification_results_table.table_name,
            "LAYOUT_METADATA_TABLE": storage.layout_metadata_table.table_name,
            "CONVERSATION_HISTORY_TABLE": storage.conversation_history_table.table_name,
            "VERIFICATION_PREFIX": os.getenv("VERIFICATION_PREFIX", "verif-"),
            "BEDROCK_MODEL": os.getenv("BEDROCK_MODEL", "anthropic.claude-3-7-sonnet-20250219-v1:0"),
            "ANTHROPIC_VERSION": os.getenv("ANTHROPIC_VERSION", "bedrock-2023-05-31"),
            "MAX_TOKENS": os.getenv("MAX_TOKENS", "24000"),
            "BUDGET_TOKENS": os.getenv("BUDGET_TOKENS", "16000"),
            "THINKING_TYPE": os.getenv("THINKING_TYPE", "enable"),
            "PROMPT_VERSION": os.getenv("PROMPT_VERSION", "1.0"),
            "TURN1_PROMPT_VERSION": os.getenv("TURN1_PROMPT_VERSION", "1.1.0"),
            "BEDROCK_SECRET_ARN": bedrock_secret.secret_arn
        }

        # Create a Lambda execution role with basic permissions
        lambda_execution_role = iam.Role(
            self,
            f"{project_prefix}-lambda-execution-role",
            assumed_by=iam.ServicePrincipal("lambda.amazonaws.com"),
            managed_policies=[
                iam.ManagedPolicy.from_aws_managed_policy_name("service-role/AWSLambdaBasicExecutionRole")
            ]
        )

        # S3 read permissions
        lambda_execution_role.add_to_policy(
            iam.PolicyStatement(
                actions=[
                    "s3:GetObject",
                    "s3:ListBucket",
                ],
                resources=[
                    storage.reference_bucket.bucket_arn,
                    f"{storage.reference_bucket.bucket_arn}/*",
                    storage.checking_bucket.bucket_arn,
                    f"{storage.checking_bucket.bucket_arn}/*",
                ]
            )
        )

        # S3 write permissions for results
        lambda_execution_role.add_to_policy(
            iam.PolicyStatement(
                actions=["s3:PutObject"],
                resources=[
                    storage.results_bucket.bucket_arn,
                    f"{storage.results_bucket.bucket_arn}/*",
                ]
            )
        )

        # DynamoDB permissions
        lambda_execution_role.add_to_policy(
            iam.PolicyStatement(
                actions=[
                    "dynamodb:PutItem",
                    "dynamodb:GetItem",
                    "dynamodb:UpdateItem",
                    "dynamodb:Query",
                ],
                resources=[
                    storage.verification_results_table.table_arn,
                    storage.layout_metadata_table.table_arn,
                    storage.conversation_history_table.table_arn,
                    f"{storage.verification_results_table.table_arn}/index/*",
                    f"{storage.layout_metadata_table.table_arn}/index/*",
                    f"{storage.conversation_history_table.table_arn}/index/*",
                ]
            )
        )

        # Bedrock permissions
        lambda_execution_role.add_to_policy(
            iam.PolicyStatement(
                actions=[
                    "bedrock:InvokeModel"
                ],
                resources=[
                    f"arn:aws:bedrock:{os.getenv('CDK_DEFAULT_REGION', 'us-east-1')}::foundation-model/anthropic.claude-3-7-sonnet-20250219-v1:0"
                ]
            )
        )
        
        # Secretsmanager permissions
        bedrock_secret.grant_read(lambda_execution_role)

        # Define the functions based on the design document
        # For initial deployment, these functions will use a placeholder nginx image
        function_definitions = [
            # Workflow initialization functions
            {"name": "initialize", "description": "Validates inputs and prepares workflow", "memory": 512, "timeout": 30},
            {"name": "fetch_historical_verification", "description": "Retrieves historical verification data", "memory": 512, "timeout": 30},
            {"name": "fetch_images", "description": "Retrieves images from S3", "memory": 768, "timeout": 45},
            
            # Prompt preparation functions
            {"name": "prepare_system_prompt", "description": "Creates system prompt for Bedrock", "memory": 512, "timeout": 30},
            {"name": "prepare_turn_prompt", "description": "Creates turn-specific prompt", "memory": 256, "timeout": 15},
            
            # Bedrock interaction functions
            {"name": "invoke_bedrock", "description": "Invokes Bedrock model API", "memory": 1024, "timeout": 120},
            
            # Response processing functions
            {"name": "process_turn1_response", "description": "Processes reference image analysis", "memory": 512, "timeout": 60},
            {"name": "process_turn2_response", "description": "Processes checking image comparison", "memory": 512, "timeout": 60},
            
            # Results handling functions
            {"name": "finalize_results", "description": "Compiles verification results", "memory": 512, "timeout": 60},
            {"name": "store_results", "description": "Saves results to DynamoDB and S3", "memory": 512, "timeout": 30},
            {"name": "notify", "description": "Sends notifications", "memory": 256, "timeout": 30},
            
            # Error handling functions
            {"name": "handle_bedrock_error", "description": "Processes Bedrock API errors", "memory": 256, "timeout": 30},
            {"name": "finalize_with_error", "description": "Creates partial results on error", "memory": 256, "timeout": 30},
            
            # Layout rendering function
            {"name": "render_layout", "description": "Renders layout images from JSON", "memory": 1024, "timeout": 60},
        ]
        
        # Create all the functions
        for func_def in function_definitions:
            function = self._create_lambda_function(
                function_name=func_def["name"],
                description=func_def["description"],
                role=lambda_execution_role,
                env_vars=common_env,
                ecr_repository=ecr_repository,
                memory_size=func_def["memory"],
                timeout=func_def["timeout"]
            )
            self.functions[func_def["name"]] = function
    
    def _create_lambda_function(
        self, 
        function_name: str, 
        description: str, 
        role, 
        env_vars, 
        ecr_repository,
        memory_size=256, 
        timeout=30
    ):
        """Helper method to create a Lambda function with container image"""
        
        # Format the function name for the resource ID
        function_id = function_name.replace("_", "")
        
        # For the first deployment, use a simple nginx image as placeholder
        # This will be replaced with the actual image later
        container_image = lambda_.DockerImageCode.from_ecr_repository(
            repository=ecr_repository,
            tag_or_digest="latest",
        ) if os.getenv('USE_ECR_IMAGES', 'false').lower() == 'true' else lambda_.DockerImageCode.from_image_asset(
            # Use a public nginx image as placeholder until we have our own images
            directory=os.path.join(os.path.dirname(os.path.abspath(__file__)), 'placeholder_image'),
            file='Dockerfile'
        )
        
        function = lambda_.DockerImageFunction(
            self,
            f"{function_id}-function-{self.resource_suffix}",
            function_name=f"{env_vars['PROJECT_PREFIX']}-{function_id}-{env_vars['STAGE']}-{self.resource_suffix}",
            description=description,
            code=container_image,
            memory_size=memory_size,
            timeout=Duration.seconds(timeout),
            environment=env_vars,
            role=role,
            log_retention=logs.RetentionDays.THREE_MONTHS,
            tracing=lambda_.Tracing.ACTIVE
        )
        
        return function