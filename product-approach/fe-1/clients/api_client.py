import requests
import os
import logging
from .aws_client import AWSClient
from .config_loader import ConfigLoader

logger = logging.getLogger(__name__)

class APIClient:
    def __init__(self):
        # Load configuration using the new config loader
        self.config_loader = ConfigLoader()

        self.base_url = self.config_loader.get('API_ENDPOINT', '')
        if not self.base_url:
            raise ValueError("API_ENDPOINT not found in configuration. For cloud deployment, ensure CONFIG_SECRET is properly set in ECS Task Definition. For local development, set API_ENDPOINT in environment variables or .streamlit/secrets.toml")

        self.api_key = self.get_api_key()
        logger.info(f"Initialized API client with base URL: {self.base_url}")
        if self.config_loader.is_loaded_from_secret():
            logger.info("Configuration loaded from AWS Secrets Manager")
        elif self.config_loader.is_loaded_from_streamlit():
            logger.info("Configuration loaded from Streamlit secrets (local development)")
        else:
            logger.info("Configuration loaded from environment variables")

    def get_api_key(self):
        # First, try direct API_KEY from environment (for local development)
        direct_api_key = os.environ.get('API_KEY', '')
        if direct_api_key:
            logger.info("Using API_KEY from environment variable")
            return direct_api_key

        # Try Streamlit secrets for local development (only if Streamlit is available)
        streamlit_api_key = self._get_streamlit_api_key()
        if streamlit_api_key:
            return streamlit_api_key

        # Check for API_KEY_SECRET_NAME from config loader or environment (for cloud deployment)
        api_key_secret_name = self.config_loader.get('API_KEY_SECRET_NAME', '') or os.environ.get('API_KEY_SECRET_NAME', '')

        if not api_key_secret_name:
            raise ValueError("API key not found. For local development, set API_KEY in environment variables or .streamlit/secrets.toml. For cloud deployment, ensure API_KEY_SECRET_NAME is set in ECS Task Definition.")

        # Try to get API key from AWS Secrets Manager
        try:
            aws_client = AWSClient()
            secret = aws_client.get_secret(api_key_secret_name)
            # Parse the secret string as JSON and extract the api_key
            import json
            secret_data = json.loads(secret['SecretString'])
            api_key = secret_data.get('api_key', '')
            if not api_key:
                raise ValueError("api_key not found in the secret data")
            logger.info("Using API_KEY from AWS Secrets Manager")
            return api_key
        except Exception as e:
            logger.error(f"Failed to retrieve API key from Secrets Manager: {str(e)}")
            raise ValueError(f"Failed to retrieve API key from AWS Secrets Manager: {str(e)}. For local development, consider setting API_KEY directly in environment variables or .streamlit/secrets.toml")

    def _get_streamlit_api_key(self):
        """Safely try to get API key from Streamlit secrets without importing at module level"""
        try:
            import streamlit as st
            if hasattr(st, 'secrets') and st.secrets:
                streamlit_api_key = st.secrets.get('API_KEY', '')
                if streamlit_api_key:
                    logger.info("Using API_KEY from Streamlit secrets (local development)")
                    return streamlit_api_key
        except Exception as e:
            logger.debug(f"Could not load API_KEY from Streamlit secrets: {str(e)}")
        return None

    def make_request(self, method, endpoint, params=None, data=None, debug=False):
        headers = {
            'X-Api-Key': self.api_key,
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        }

        base = self.base_url.rstrip('/')
        path = endpoint.lstrip('/')
        url = f"{base}/{path}"

        if debug:
            return {
                "debug_info": {
                    "url": url,
                    "method": method,
                    "headers": {**headers, 'X-Api-Key': '*****'},
                    "params": params,
                    "data": data
                }
            }

        try:
            logger.info(f"Making {method} request to {url}")
            response = requests.request(
                method,
                url,
                headers=headers,
                params=params,
                json=data,
                timeout=10
            )

            logger.info(f"Response status: {response.status_code}")
            logger.debug(f"Response headers: {response.headers}")

            response.raise_for_status()

            try:
                return response.json()
            except ValueError:
                return {"text": response.text}

        except requests.RequestException as e:
            logger.error(f"API request failed: {method} {url} - {str(e)}")
            raise Exception(f"Failed to {method.lower()} {endpoint}: {str(e)}")

    # ✅ Fixed: all endpoint paths are now relative to the base /v1
    def health_check(self, debug=False):
        return self.make_request('GET', 'api/health', debug=debug)

    def lookup_verification(self, checking_image_url, vending_machine_id=None, limit=1):
        params = {
            "checkingImageUrl": checking_image_url,
            "limit": limit
        }
        if vending_machine_id:
            params["vendingMachineId"] = vending_machine_id
        return self.make_request('GET', 'api/verifications/lookup', params=params)

    def initiate_verification(self, verificationType, referenceImageUrl, checkingImageUrl,
                              notificationEnabled=False):
        """
        Simplified verification API call with only required fields.

        Args:
            verificationType: Type of verification (e.g., "LAYOUT_VS_CHECKING", "PREVIOUS_VS_CURRENT")
            referenceImageUrl: URL of the reference image
            checkingImageUrl: URL of the checking image
            notificationEnabled: Whether to enable notifications (default: False)
        """
        # Build the verification context according to simplified API specification
        verification_context = {
            "verificationType": verificationType,
            "referenceImageUrl": referenceImageUrl,
            "checkingImageUrl": checkingImageUrl,
            "notificationEnabled": notificationEnabled
        }

        # Wrap the verification context in the expected API structure
        data = {
            "verificationContext": verification_context
        }

        return self.make_request('POST', 'api/verifications', data=data)

    def list_verifications(self, params=None):
        return self.make_request('GET', 'api/verifications', params=params)

    def get_verification_details(self, verification_id):
        return self.make_request('GET', f'api/verifications/{verification_id}')

    def get_verification_conversation(self, verification_id):
        return self.make_request('GET', f'api/verifications/{verification_id}/conversation')

    def browse_images(self, path='', bucket_type='reference'):
        params = {"bucketType": bucket_type}
        # Use different endpoints based on whether path is provided
        if path and path.strip():
            # Use the path endpoint for non-empty paths
            endpoint = f'api/images/browser/{path.strip()}'
        else:
            # Use the base endpoint for empty/root paths
            endpoint = 'api/images/browser'
        return self.make_request('GET', endpoint, params=params)

    def get_image_url(self, key, bucket_type='reference'):
        import urllib.parse
        # URL encode the key to handle special characters and path separators
        encoded_key = urllib.parse.quote(key, safe='')
        params = {"bucketType": bucket_type}
        return self.make_request('GET', f'api/images/{encoded_key}/view', params=params)
