import requests
import os
import logging
from .aws_client import AWSClient

logger = logging.getLogger(__name__)

class APIClient:
    def __init__(self):
        self.base_url = os.environ.get('API_ENDPOINT', '')
        if not self.base_url:
            raise ValueError("API_ENDPOINT environment variable not set")
        self.api_key = self.get_api_key()

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
        else:
            raise ValueError("API key not found in environment or secrets")

    def make_request(self, method, endpoint, params=None, data=None):
        headers = {'x-api-key': self.api_key}
        url = f"{self.base_url}/{endpoint}"
        try:
            response = requests.request(method, url, headers=headers, params=params, json=data, timeout=10)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            logger.error(f"API request failed: {method} {endpoint} - {str(e)}")
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

    def health_check(self):
        return self.make_request('GET', 'api/v1/health')