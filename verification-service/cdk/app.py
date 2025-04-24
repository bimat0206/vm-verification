#!/usr/bin/env python3
import os
import string
import random
from aws_cdk import App, Environment

from verification_service_stack import VerificationServiceStack

# Function to generate a random suffix (same as in VerificationServiceStack)
def generate_random_suffix(length=8):
    """Generate a random alphanumeric suffix for resource names"""
    chars = string.ascii_lowercase + string.digits
    return ''.join(random.choice(chars) for _ in range(length))

app = App()

# Get environment variables for deployment configuration
env = Environment(
    account=os.environ.get('CDK_DEPLOY_ACCOUNT', os.environ['CDK_DEFAULT_ACCOUNT']),
    region=os.environ.get('CDK_DEPLOY_REGION', os.environ['CDK_DEFAULT_REGION'])
)

# Get resource prefix from environment or use default
resource_prefix = os.environ.get('RESOURCE_PREFIX', 'verification')

# Generate random suffix
random_suffix = generate_random_suffix()

# Create the stack with the specified resource prefix and random suffix
VerificationServiceStack(
    app, 
    f"{resource_prefix}-service-stack-{random_suffix}",
    resource_prefix=resource_prefix,
    random_suffix=random_suffix,  # Pass the random suffix to the stack
    env=env,
    description="Verification Service Infrastructure"
)

app.synth()