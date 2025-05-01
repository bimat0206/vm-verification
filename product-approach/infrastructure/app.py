#!/usr/bin/env python3
import os
from dotenv import load_dotenv
from aws_cdk import App, Environment
from infrastructure_stack import KootoroVerificationStack

# Load environment variables
load_dotenv()

app = App()

# Determine environment
account = os.getenv('CDK_DEFAULT_ACCOUNT', os.getenv('AWS_ACCOUNT_ID'))
region = os.getenv('CDK_DEFAULT_REGION', os.getenv('AWS_REGION', 'us-east-1'))

env = Environment(account=account, region=region)

# Create the main infrastructure stack
KootoroVerificationStack(
    app, 
    f"{os.getenv('PROJECT_PREFIX', 'kootoro')}-verification",
    env=env,
    description="Kootoro GenAI Vending Machine Verification Solution"
)

app.synth()