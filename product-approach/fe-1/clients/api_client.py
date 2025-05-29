import requests
import os
import logging
import streamlit as st
from .aws_client import AWSClient

logger = logging.getLogger(__name__)

class APIClient:
    def __init__(self):
        self.base_url = os.environ.get('API_ENDPOINT', '')
        if not self.base_url and hasattr(st, 'secrets'):
            try:
                self.base_url = st.secrets.get("API_ENDPOINT", "")
            except Exception as e:
                logger.error(f"Failed to get API_ENDPOINT from Streamlit secrets: {str(e)}")
        
        if not self.base_url:
            raise ValueError("API_ENDPOINT not found in environment variables or Streamlit secrets")
        
        self.api_key = self.get_api_key()
        logger.info(f"Initialized API client with base URL: {self.base_url}")

    def get_api_key(self):
        if 'API_KEY' in os.environ:
            return os.environ['API_KEY']
        elif 'API_KEY_SECRET_NAME' in os.environ:
            aws_client = AWSClient()
            secret_name = os.environ['API_KEY_SECRET_NAME']
            try:
                secret = aws_client.get_secret(secret_name)
                return secret['SecretString']
            except Exception as e:
                logger.error(f"Failed to retrieve API key from Secrets Manager: {str(e)}")
                raise
        elif hasattr(st, 'secrets'):
            try:
                api_key = st.secrets.get("API_KEY", "")
                if api_key:
                    return api_key
            except Exception as e:
                logger.error(f"Failed to get API_KEY from Streamlit secrets: {str(e)}")
        
        raise ValueError("API key not found in environment variables or Streamlit secrets")

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
        return self.make_request('GET', 'verifications/lookup', params=params)

    def initiate_verification(self, verification_type, reference_image_url, checking_image_url, 
                              vending_machine_id, layout_id=None, layout_prefix=None,
                              previous_verification_id=None):
        data = {
            "verificationType": verification_type,
            "referenceImageUrl": reference_image_url,
            "checkingImageUrl": checking_image_url,
            "vendingMachineId": vending_machine_id,
            "notificationEnabled": False
        }

        if verification_type == "LAYOUT_VS_CHECKING":
            if not layout_id or not layout_prefix:
                raise ValueError("layout_id and layout_prefix are required for LAYOUT_VS_CHECKING verification")
            data["layoutId"] = layout_id
            data["layoutPrefix"] = layout_prefix
        elif verification_type == "PREVIOUS_VS_CURRENT":
            if previous_verification_id:
                data["previousVerificationId"] = previous_verification_id

        return self.make_request('POST', 'verifications', data=data)

    def list_verifications(self, params=None):
        return self.make_request('GET', 'verifications', params=params)

    def get_verification_details(self, verification_id):
        return self.make_request('GET', f'verifications/{verification_id}')

    def get_verification_conversation(self, verification_id):
        return self.make_request('GET', f'verifications/{verification_id}/conversation')

    def browse_images(self, path='', bucket_type='reference'):
        params = {"bucketType": bucket_type}
        return self.make_request('GET', f'images/browser/{path}', params=params)

    def get_image_url(self, key):
        return self.make_request('GET', f'images/{key}/view')
