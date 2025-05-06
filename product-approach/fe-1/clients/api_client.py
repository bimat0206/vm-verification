import requests
import os
import logging
import streamlit as st
from .aws_client import AWSClient

logger = logging.getLogger(__name__)

class APIClient:
    def __init__(self):
        # Try to get API endpoint from environment variables first, then from Streamlit secrets
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
        # Try environment variables first
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
        
        # Then try Streamlit secrets
        elif hasattr(st, 'secrets'):
            try:
                api_key = st.secrets.get("API_KEY", "")
                if api_key:
                    return api_key
            except Exception as e:
                logger.error(f"Failed to get API_KEY from Streamlit secrets: {str(e)}")
        
        # If we get here, we couldn't find the API key
        raise ValueError("API key not found in environment variables or Streamlit secrets")

    def make_request(self, method, endpoint, params=None, data=None, debug=False):
        headers = {
            'X-Api-Key': self.api_key,  # Using capital X to match API Gateway expectations
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        }
        
        # Remove any trailing slash from base_url and leading slash from endpoint
        base = self.base_url.rstrip('/')
        path = endpoint.lstrip('/')
        url = f"{base}/{path}"
        
        if debug:
            # Return debug info instead of making actual request
            return {
                "debug_info": {
                    "url": url,
                    "method": method,
                    "headers": {**headers, 'X-Api-Key': '*****'},  # Mask API key
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
            
            # Log response details for debugging
            logger.info(f"Response status: {response.status_code}")
            logger.info(f"Response headers: {response.headers}")
            
            # Try to parse JSON response
            try:
                json_response = response.json()
                logger.debug(f"Response body: {json_response}")
            except ValueError:
                logger.debug(f"Response body (not JSON): {response.text[:200]}...")
            
            # Raise exception for bad status codes
            response.raise_for_status()
            
            # Return response data
            try:
                return response.json()
            except ValueError:
                return {"text": response.text}
                
        except requests.RequestException as e:
            logger.error(f"API request failed: {method} {url} - {str(e)}")
            raise Exception(f"Failed to {method.lower()} {endpoint}: {str(e)}")

    def lookup_verification(self, checking_image_url, vending_machine_id=None, limit=1):
        params = {
            "checkingImageUrl": checking_image_url,
            "limit": limit
        }
        if vending_machine_id:
            params["vendingMachineId"] = vending_machine_id
        return self.make_request('GET', 'api/v1/verifications/lookup', params=params)

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
        
        # Add type-specific properties
        if verification_type == "LAYOUT_VS_CHECKING":
            if not layout_id or not layout_prefix:
                raise ValueError("layout_id and layout_prefix are required for LAYOUT_VS_CHECKING verification")
            data["layoutId"] = layout_id
            data["layoutPrefix"] = layout_prefix
        elif verification_type == "PREVIOUS_VS_CURRENT":
            if previous_verification_id:
                data["previousVerificationId"] = previous_verification_id
                
        return self.make_request('POST', 'api/v1/verifications', data=data)

    def list_verifications(self, params=None):
        return self.make_request('GET', 'api/v1/verifications', params=params)

    def get_verification_details(self, verification_id):
        return self.make_request('GET', f'api/v1/verifications/{verification_id}')

    def get_verification_conversation(self, verification_id):
        return self.make_request('GET', f'api/v1/verifications/{verification_id}/conversation')

    def browse_images(self, path='', bucket_type='reference'):
        params = {"bucketType": bucket_type}
        return self.make_request('GET', f'api/v1/images/browser/{path}', params=params)

    def get_image_url(self, key):
        return self.make_request('GET', f'api/v1/images/{key}/view')

    def health_check(self, debug=False):
        # Use the standard API v1 health endpoint
        endpoint = 'api/v1/health'
        
        if debug:
            # Return debug info instead of making actual requests
            return {
                "debug_info": {
                    "base_url": self.base_url,
                    "endpoint": endpoint,
                    "method": "GET",
                    "headers": {
                        'X-Api-Key': '*****',  # Masked for security
                        'Content-Type': 'application/json',
                        'Accept': 'application/json'
                    }
                }
            }
        
        try:
            logger.info(f"Making health check request to {endpoint}...")
            
            # For direct requests to handle different methods
            url = f"{self.base_url.rstrip('/')}/{endpoint.lstrip('/')}"
            headers = {
                'X-Api-Key': self.api_key,  # Using capital X to match API Gateway expectations
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            }
            
            response = requests.get(
                url=url,
                headers=headers,
                timeout=10
            )
            
            # Log response details
            logger.info(f"Response status: {response.status_code}")
            logger.info(f"Response headers: {response.headers}")
            
            # Raise for status to catch HTTP errors
            response.raise_for_status()
            
            # Return the JSON response
            try:
                return response.json()
            except ValueError:
                return {"status": "healthy", "message": response.text}
                
        except requests.exceptions.HTTPError as e:
            error_msg = f"Health check failed with HTTP error: {e}"
            logger.error(error_msg)
            if response.status_code == 403:
                error_msg += "\nPossible API key authentication issue. Please verify your API key."
            raise Exception(error_msg)
            
        except Exception as e:
            error_msg = f"Health check failed: {str(e)}"
            logger.error(error_msg)
            raise Exception(error_msg)
