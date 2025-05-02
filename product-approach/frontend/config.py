"""
Configuration Management for Vending Machine Verification System

This module handles configuration loading from environment variables and AWS Secrets Manager,
providing a centralized configuration for the application.
"""

import os
import json
import time
import logging
import boto3
from typing import Dict, Optional, Tuple

logger = logging.getLogger("streamlit-app")

class ConfigManager:
    """Manages application configuration loading and validation."""
    
    def __init__(self):
        self.api_endpoint = ""
        self.dynamodb_table = ""
        self.s3_bucket = ""
        self.region = ""
        self.secret_arn = ""
        self.loaded = False
    
    def load_from_environment(self) -> None:
        """Load basic configuration from environment variables."""
        self.api_endpoint = os.environ.get("API_ENDPOINT", "")
        self.dynamodb_table = os.environ.get("DYNAMODB_TABLE", "")
        self.s3_bucket = os.environ.get("S3_BUCKET", "")
        self.region = os.environ.get("REGION", "us-east-1")
        self.secret_arn = os.environ.get("SECRET_ARN", "")
        
        logger.info(
            f"Environment variables: API_ENDPOINT={self.api_endpoint}, "
            f"DYNAMODB_TABLE={self.dynamodb_table}, S3_BUCKET={self.s3_bucket}, "
            f"REGION={self.region}, SECRET_ARN present: {bool(self.secret_arn)}"
        )
    
    def load_from_secrets_manager(self) -> bool:
        """
        Load configuration from AWS Secrets Manager if SECRET_ARN is defined.
        
        Returns:
            bool: True if secrets were successfully loaded, False otherwise.
        """
        if not self.secret_arn:
            logger.info("No SECRET_ARN defined, skipping Secrets Manager")
            return False
            
        logger.info(f"Attempting to load config from Secrets Manager: {self.secret_arn} in {self.region}")
        
        try:
            # Create a Secrets Manager client
            session = boto3.session.Session()
            client = session.client(service_name='secretsmanager', region_name=self.region)
            
            # Try multiple times with backoff
            max_retries = 3
            for attempt in range(max_retries):
                try:
                    logger.info(f"Attempt {attempt+1} to get secret")
                    response = client.get_secret_value(SecretId=self.secret_arn)
                    
                    if 'SecretString' in response:
                        logger.info("Successfully retrieved secret")
                        config = json.loads(response['SecretString'])
                        
                        # Override environment variables if secrets are available
                        if config.get("api_endpoint"):
                            self.api_endpoint = config.get("api_endpoint")
                        if config.get("dynamodb_table_name"):
                            self.dynamodb_table = config.get("dynamodb_table_name")
                        if config.get("s3_bucket_name"):
                            self.s3_bucket = config.get("s3_bucket_name")
                        
                        logger.info("Configuration loaded from Secrets Manager")
                        return True
                    else:
                        logger.warning("Secret has no SecretString field")
                except Exception as e:
                    logger.warning(f"Attempt {attempt+1} failed: {str(e)}")
                    if attempt < max_retries - 1:  # Don't sleep on the last attempt
                        time.sleep(2 ** attempt)  # Exponential backoff
        except Exception as e:
            logger.error(f"Error retrieving configuration from Secrets Manager: {str(e)}")
        
        return False
    
    def validate_and_set_defaults(self) -> None:
        """Validate configuration and set defaults for missing values."""
        if not self.api_endpoint:
            self.api_endpoint = "https://am0ncga8rk.execute-api.us-east-1.amazonaws.com/v1"
            logger.info(f"Using fallback API endpoint: {self.api_endpoint}")
        
        if not self.dynamodb_table:
            self.dynamodb_table = "VerificationResults"
            logger.info(f"Using fallback DynamoDB table: {self.dynamodb_table}")
        
        if not self.s3_bucket:
            self.s3_bucket = "vending-verification-images"
            logger.info(f"Using fallback S3 bucket: {self.s3_bucket}")
    
    def test_api_connection(self) -> Tuple[bool, str]:
        """
        Test connection to the API endpoint.
        
        Returns:
            Tuple[bool, str]: Success status and a message describing the result
        """
        try:
            import requests
            logger.info(f"Testing API connectivity to {self.api_endpoint}/api/v1/health")
            response = requests.get(f"{self.api_endpoint}/api/v1/health", timeout=5)
            if response.status_code == 200:
                logger.info("API connection test successful")
                return True, "Connection successful"
            else:
                logger.warning(f"API connection test returned status code: {response.status_code}")
                return False, f"API returned status code {response.status_code}"
        except Exception as e:
            logger.warning(f"API connection test failed: {str(e)}")
            return False, f"Connection failed: {str(e)}"
    
    def load_config(self) -> Dict[str, str]:
        """
        Load, validate and return the configuration.
        
        Returns:
            Dict[str, str]: Configuration dictionary
        """
        self.load_from_environment()
        self.load_from_secrets_manager()
        self.validate_and_set_defaults()
        self.loaded = True
        
        return {
            "api_endpoint": self.api_endpoint,
            "dynamodb_table_name": self.dynamodb_table,
            "s3_bucket_name": self.s3_bucket,
            "region": self.region
        }