from aws_cdk import (
    aws_apigateway as apigw,
    aws_iam as iam,
    Duration,
)
from constructs import Construct
import os

class ApiConstruct(Construct):
    """
    Creates the API Gateway for the Kootoro verification system.
    Implements all the endpoints described in the design document.
    """
    
    def __init__(
        self,
        scope: Construct,
        id: str,
        project_prefix: str,
        stage: str,
        lambda_functions,
        state_machine,
        resource_suffix: str,
        **kwargs
    ) -> None:
        super().__init__(scope, id, **kwargs)
        self.resource_suffix = resource_suffix
        
        # Create API Gateway
        self.api = apigw.RestApi(
            self,
            f"{project_prefix}-api-{self.resource_suffix}",
            rest_api_name=f"{project_prefix}-verification-api-{stage}-{self.resource_suffix}",
            description="API for Kootoro vending machine verification",
            default_cors_preflight_options=apigw.CorsOptions(
                allow_origins=apigw.Cors.ALL_ORIGINS,
                allow_methods=apigw.Cors.ALL_METHODS,
                allow_headers=[
                    "Content-Type",
                    "X-Amz-Date",
                    "Authorization",
                    "X-Api-Key",
                ],
                max_age=Duration.days(1)
            ),
            deploy_options=apigw.StageOptions(
                stage_name=stage,
                throttling_rate_limit=int(os.getenv("API_RATE_LIMIT", "100")),
                throttling_burst_limit=int(os.getenv("API_BURST_LIMIT", "50")),
                logging_level=apigw.MethodLoggingLevel.INFO,
                metrics_enabled=True,
                data_trace_enabled=os.getenv("API_DATA_TRACE", "false").lower() == "true"
            ),
            endpoint_types=[apigw.EndpointType.REGIONAL]
        )
        
        # Add API Key for authorization
        api_key = self.api.add_api_key(
            f"{project_prefix}-api-key-{self.resource_suffix}",
            api_key_name=f"{project_prefix}-verification-{stage}-{self.resource_suffix}"
        )
        
        # Create usage plan
        plan = self.api.add_usage_plan(
            f"{project_prefix}-usage-plan-{self.resource_suffix}",
            name=f"{project_prefix}-verification-plan-{stage}-{self.resource_suffix}",
            throttle=apigw.ThrottleSettings(
                rate_limit=int(os.getenv("API_RATE_LIMIT", "100")),
                burst_limit=int(os.getenv("API_BURST_LIMIT", "50"))
            )
        )
        
        plan.add_api_key(api_key)
        plan.add_api_stage(
            stage=self.api.deployment_stage
        )
        
        # Create API resources and methods
        # Base path: /api/v1
        api_v1 = self.api.root.add_resource("api").add_resource("v1")
        
        # ======================
        # Verification endpoints
        # ======================
        verifications = api_v1.add_resource("verifications")
        
        # POST /api/v1/verifications - Create a new verification
        start_verification_integration = apigw.LambdaIntegration(
            lambda_functions["initialize"],
            proxy=True,
            integration_responses=[
                apigw.IntegrationResponse(
                    status_code="202",
                    response_parameters={
                        "method.response.header.Access-Control-Allow-Origin": "'*'",
                        "method.response.header.Location": "integration.response.body.location"
                    }
                )
            ]
        )
        
        verifications.add_method(
            "POST",
            start_verification_integration,
            api_key_required=True,
            method_responses=[
                apigw.MethodResponse(
                    status_code="202",
                    response_parameters={
                        "method.response.header.Access-Control-Allow-Origin": True,
                        "method.response.header.Location": True
                    }
                )
            ]
        )
        
        # GET /api/v1/verifications - List verifications
        # Note: In production, this would use a dedicated Lambda function for listing
        list_verifications_integration = apigw.LambdaIntegration(
            lambda_functions["initialize"],
            proxy=True
        )
        
        verifications.add_method(
            "GET",
            list_verifications_integration,
            api_key_required=True
        )
        
        # GET /api/v1/verifications/{id} - Get verification by ID
        verification_by_id = verifications.add_resource("{id}")
        
        # Note: In production, this would use a dedicated Lambda function for retrieval
        get_verification_integration = apigw.LambdaIntegration(
            lambda_functions["initialize"],
            proxy=True
        )
        
        verification_by_id.add_method(
            "GET",
            get_verification_integration,
            api_key_required=True
        )
        
        # GET /api/v1/verifications/{id}/conversation - Get conversation for verification
        conversation = verification_by_id.add_resource("conversation")
        
        # Note: In production, this would use a dedicated Lambda function for retrieval
        get_conversation_integration = apigw.LambdaIntegration(
            lambda_functions["initialize"],
            proxy=True
        )
        
        conversation.add_method(
            "GET",
            get_conversation_integration,
            api_key_required=True
        )
        
        # GET /api/v1/verifications/lookup - Lookup verifications by checking image
        lookup = verifications.add_resource("lookup")
        
        # Note: In production, this would use a dedicated Lambda function for lookup
        lookup_integration = apigw.LambdaIntegration(
            lambda_functions["initialize"],
            proxy=True
        )
        
        lookup.add_method(
            "GET",
            lookup_integration,
            api_key_required=True
        )
        
        # ======================
        # Health check endpoint
        # ======================
        health = api_v1.add_resource("health")
        
        health_integration = apigw.MockIntegration(
            integration_responses=[
                apigw.IntegrationResponse(
                    status_code="200",
                    response_templates={
                        "application/json": '{"status": "healthy", "version": "1.0.0", "timestamp": "$context.requestTime"}'
                    }
                )
            ],
            request_templates={
                "application/json": '{"statusCode": 200}'
            }
        )
        
        health.add_method(
            "GET",
            health_integration,
            method_responses=[
                apigw.MethodResponse(
                    status_code="200",
                    response_models={
                        "application/json": apigw.Model.EMPTY_MODEL
                    }
                )
            ]
        )
        
        # ======================
        # Image rendering APIs
        # ======================
        images = api_v1.add_resource("images")
        
        # GET /api/v1/images/{key}/view - Get pre-signed URL for an image
        image_view = images.add_resource("{key}").add_resource("view")
        
        # Note: In production, this would use a dedicated Lambda function for image view
        image_view_integration = apigw.LambdaIntegration(
            lambda_functions["initialize"],
            proxy=True
        )
        
        image_view.add_method(
            "GET",
            image_view_integration,
            api_key_required=True
        )
        
        # GET /api/v1/images/browser/{path} - Browse images
        browser = images.add_resource("browser").add_resource("{path+}")
        
        # Note: In production, this would use a dedicated Lambda function for browsing
        browser_integration = apigw.LambdaIntegration(
            lambda_functions["initialize"],
            proxy=True
        )
        
        browser.add_method(
            "GET",
            browser_integration,
            api_key_required=True
        )