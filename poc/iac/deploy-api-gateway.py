import boto3

# ---------------------------
# User-defined configuration
# ---------------------------
API_ID = "hpux2uegnd"      # Your API Gateway REST API ID
STAGE_NAME = "v1"          # The stage name to deploy to
AWS_REGION = "us-east-1"   # The AWS region for your API Gateway

# Create boto3 client for API Gateway
client = boto3.client('apigateway', region_name=AWS_REGION)

def create_deployment(api_id, stage_name):
    """
    Create a new deployment for the given API and stage.
    """
    print(f"Creating deployment for API: {api_id}, Stage: {stage_name}")
    response = client.create_deployment(
        restApiId=api_id,
        stageName=stage_name,
        description='Auto deployment by script'
    )
    deployment_id = response['id']
    print(f"Deployment created with ID: {deployment_id}")
    return deployment_id

def update_stage(api_id, stage_name, deployment_id):
    """
    Point the stage to the new deployment.
    """
    print(f"Updating stage '{stage_name}' to deployment '{deployment_id}'")
    client.update_stage(
        restApiId=api_id,
        stageName=stage_name,
        patchOperations=[
            {
                'op': 'replace',
                'path': '/deploymentId',
                'value': deployment_id
            }
        ]
    )
    print("Stage updated to new deployment.")

def main():
    deployment_id = create_deployment(API_ID, STAGE_NAME)
    update_stage(API_ID, STAGE_NAME, deployment_id)
    print(f"API {API_ID} deployed to stage {STAGE_NAME} and made active.")

if __name__ == "__main__":
    main()
