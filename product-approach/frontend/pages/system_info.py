"""
System Info Page for Vending Machine Verification System

This module contains the system information page implementation.
"""

import os
import boto3
import streamlit as st
import logging
from typing import Dict
from pages import BasePage
from api import APIClient

logger = logging.getLogger("streamlit-app")

class SystemInfoPage(BasePage):
    """Page for system information and diagnostics."""
    
    def __init__(self, api_client: APIClient, config: Dict[str, str]):
        super().__init__(api_client)
        self.config = config
    
    def render(self) -> None:
        """Render the system info page."""
        st.header("System Information")
        
        self._display_environment_info()
        self._display_api_connection_test()
        self._display_secrets_manager_test()
        self._display_s3_test()
    
    def _display_environment_info(self) -> None:
        """Display environment information."""
        st.subheader("Environment")
        env_info = {
            "API Endpoint": self.config.get("api_endpoint", "Not configured"),
            "DynamoDB Table": self.config.get("dynamodb_table_name", "Not configured"),
            "S3 Bucket": self.config.get("s3_bucket_name", "Not configured"),
            "Region": os.environ.get("REGION", "Not set"),
            "Secret ARN": os.environ.get("SECRET_ARN", "Not set"),
            "Deployment Type": "App Runner",
            "Instance ID": os.environ.get("AWS_APP_RUNNER_INSTANCE_ID", "Not available")
        }
        
        for key, value in env_info.items():
            st.text(f"{key}: {value}")
    
    def _display_api_connection_test(self) -> None:
        """Display API connection test section."""
        st.subheader("API Connection Test")
        if st.button("Test API Connection"):
            success, data, status_code = self.api_client.check_health()
            if success:
                st.success(f"API connection successful! Status code: {status_code}")
                st.json(data)
            else:
                st.error(f"API connection failed with status code: {status_code}")
                st.json(data)
    
    def _display_secrets_manager_test(self) -> None:
        """Display Secrets Manager test section if configured."""
        if "SECRET_ARN" in os.environ:
            st.subheader("Secrets Manager Test")
            if st.button("Test Secrets Manager Connection"):
                try:
                    session = boto3.session.Session()
                    client = session.client(
                        service_name='secretsmanager', 
                        region_name=self.config.get("region", "us-east-1")
                    )
                    response = client.describe_secret(SecretId=os.environ.get("SECRET_ARN"))
                    st.success("Secrets Manager connection successful!")
                    
                    # Display non-sensitive metadata
                    st.json({
                        "Name": response.get("Name"),
                        "LastChangedDate": str(response.get("LastChangedDate")),
                        "LastAccessedDate": str(response.get("LastAccessedDate")),
                        "ARN": response.get("ARN")
                    })
                except Exception as e:
                    st.error(f"Secrets Manager connection error: {str(e)}")
    
    def _display_s3_test(self) -> None:
        """Display S3 test section."""
        st.subheader("S3 Bucket Test")
        if st.button("Test S3 Connection"):
            try:
                session = boto3.session.Session()
                s3_client = session.client('s3', region_name=self.config.get("region", "us-east-1"))
                bucket = self.config.get("s3_bucket_name")
                
                if bucket:
                    response = s3_client.list_objects_v2(Bucket=bucket, MaxKeys=5)
                    st.success(f"Successfully connected to S3 bucket: {bucket}")
                    
                    # Show a few objects in the bucket
                    if "Contents" in response and len(response["Contents"]) > 0:
                        st.write("Sample objects in bucket:")
                        for obj in response["Contents"]:
                            st.write(f"- {obj['Key']} ({obj['Size']} bytes)")
                    else:
                        st.info("Bucket exists but appears to be empty")
                else:
                    st.warning("No S3 bucket configured")
            except Exception as e:
                st.error(f"S3 connection error: {str(e)}")