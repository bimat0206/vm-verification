"""
API Client for Vending Machine Verification System

This module provides a client for interacting with the verification API endpoints.
Enhanced with better S3 browsing capabilities.
"""

import logging
import requests
from typing import Dict, List, Optional, Any, Tuple

logger = logging.getLogger("streamlit-app")

class APIClient:
    """Client for interacting with the verification API."""
    
    def __init__(self, api_endpoint: str):
        self.api_endpoint = api_endpoint
        logger.info(f"API client initialized with endpoint: {api_endpoint}")
    
    def create_comparison(self, 
                        reference_img: str, 
                        checking_img: str, 
                        machine_id: str, 
                        location: Optional[str] = None,
                        verification_type: str = "LAYOUT_VS_CHECKING",
                        notification_enabled: bool = False) -> Optional[Dict[str, Any]]:
        """
        Start a new verification comparison.
        
        Args:
            reference_img: S3 key for the reference image
            checking_img: S3 key for the checking image
            machine_id: ID of the vending machine
            location: Optional location of the vending machine
            verification_type: Type of verification ("LAYOUT_VS_CHECKING" or "PREVIOUS_VS_CURRENT")
            notification_enabled: Whether to send notifications on completion
            
        Returns:
            Optional[Dict[str, Any]]: Response data or None if request failed
        """
        data = {
            "verificationType": verification_type,
            "referenceImageUrl": reference_img,
            "checkingImageUrl": checking_img,
            "vendingMachineId": machine_id,
            "notificationEnabled": notification_enabled
        }
        
        # Add optional fields if provided
        if location:
            data["location"] = location
            
        # Add layoutId and layoutPrefix if it's a layout-based verification
        if verification_type == "LAYOUT_VS_CHECKING" and "_" in reference_img:
            try:
                # Attempt to extract layoutId and layoutPrefix from the referenceImageUrl
                # Expected format: s3://bucket/path/layoutId_layoutPrefix/image.png
                parts = reference_img.split('/')[-2].split('_')
                if len(parts) >= 2:
                    data["layoutId"] = int(parts[0])
                    data["layoutPrefix"] = parts[1]
            except (IndexError, ValueError) as e:
                logger.warning(f"Could not parse layoutId and layoutPrefix from {reference_img}: {e}")
        
        logger.info(f"Creating comparison with data: {data}")
        try:
            response = requests.post(f"{self.api_endpoint}/api/v1/verifications", json=data)
            if response.status_code == 200 or response.status_code == 202:
                logger.info(f"Comparison created successfully: {response.status_code}")
                return response.json()
            else:
                logger.error(f"API Error: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            logger.error(f"Error creating comparison: {str(e)}")
            return None
    
    def get_comparison(self, comparison_id: str) -> Optional[Dict[str, Any]]:
        """
        Get details of a verification comparison.
        
        Args:
            comparison_id: ID of the comparison to retrieve
            
        Returns:
            Optional[Dict[str, Any]]: Comparison data or None if request failed
        """
        logger.info(f"Fetching comparison: {comparison_id}")
        try:
            response = requests.get(f"{self.api_endpoint}/api/v1/verifications/{comparison_id}")
            if response.status_code == 200:
                logger.info(f"Comparison fetched successfully: {response.status_code}")
                return response.json()
            else:
                logger.error(f"API Error: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            logger.error(f"Error fetching comparison: {str(e)}")
            return None
    
    def lookup_verification(self, checking_img_url: str, machine_id: Optional[str] = None, 
                           limit: int = 1) -> Optional[Dict[str, Any]]:
        """
        Lookup previous verification by checking image.
        
        Args:
            checking_img_url: S3 key for checking image
            machine_id: Optional machine ID filter
            limit: Maximum number of results to return
            
        Returns:
            Optional[Dict[str, Any]]: Lookup results or None if request failed
        """
        params = {
            "checkingImageUrl": checking_img_url,
            "limit": limit
        }
        
        if machine_id:
            params["vendingMachineId"] = machine_id
            
        logger.info(f"Looking up verification with params: {params}")
        try:
            response = requests.get(f"{self.api_endpoint}/api/v1/verifications/lookup", params=params)
            if response.status_code == 200:
                logger.info(f"Verification lookup successful: {response.status_code}")
                return response.json()
            else:
                logger.error(f"API Error: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            logger.error(f"Error looking up verification: {str(e)}")
            return None
    
    def browse_images(self, path: str, bucket_type: str) -> Optional[Dict[str, Any]]:
        """
        Browse images in S3 bucket with enhanced functionality.
        
        Args:
            path: Path prefix in S3
            bucket_type: Type of bucket ("reference" or "checking")
            
        Returns:
            Optional[Dict[str, Any]]: Browse results or None if request failed
        """
        params = {
            "bucketType": bucket_type
        }
        
        logger.info(f"Browsing images at path: {path}, bucket type: {bucket_type}")
        
        # We'll simulate the API response for now to provide better browsing UX
        # In production, this would be replaced with the actual API call
        
        # Simulated API response for testing UI functionality
        # In production, uncomment the API call and remove this simulation
        """
        try:
            response = requests.get(f"{self.api_endpoint}/api/v1/images/browser/{path}", params=params)
            if response.status_code == 200:
                logger.info("Images browsed successfully")
                return response.json()
            else:
                logger.error(f"API Error: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            logger.error(f"Error browsing images: {str(e)}")
            return None
        """
        
        # Simulation for development - REMOVE in production
        import random
        
        # Generate simulated bucket and path info
        bucket = f"{bucket_type}-bucket-{random.randint(100, 999)}"
        
        # Generate simulated folders and images based on bucket type and path
        items = []
        
        # Add folders
        if path == "":
            # Root level
            folders = ["2025", "machines", "templates"] if bucket_type == "reference" else ["2025", "daily", "special"]
            for f in folders:
                items.append({
                    "name": f,
                    "type": "folder",
                    "path": f,
                    "lastModified": "2025-05-01T10:00:00Z",
                    "size": 0,
                    "bucket": bucket
                })
        elif path == "2025":
            # Year folders
            months = ["01-January", "02-February", "03-March", "04-April", "05-May"]
            for m in months:
                items.append({
                    "name": m,
                    "type": "folder",
                    "path": f"2025/{m}",
                    "lastModified": "2025-05-01T10:00:00Z",
                    "size": 0,
                    "bucket": bucket
                })
        elif "machines" in path:
            # Machine folders
            machines = ["VM-3245", "VM-3246", "VM-3247"]
            for m in machines:
                items.append({
                    "name": m,
                    "type": "folder",
                    "path": f"machines/{m}",
                    "lastModified": "2025-05-01T10:00:00Z",
                    "size": 0,
                    "bucket": bucket
                })
        
        # Add some images based on path
        if "VM-3245" in path or "05-May" in path:
            # Add simulated images
            image_types = [".jpg", ".png"]
            prefixes = ["layout", "reference", "checking"] if bucket_type == "reference" else ["check", "daily", "weekly"]
            for i in range(1, 6):
                img_type = image_types[i % len(image_types)]
                prefix = prefixes[i % len(prefixes)]
                img_name = f"{prefix}_{i}{img_type}"
                items.append({
                    "name": img_name,
                    "type": "image",
                    "path": f"{path}/{img_name}" if path else img_name,
                    "lastModified": f"2025-05-{i:02d}T10:00:00Z",
                    "size": random.randint(100000, 5000000),
                    "bucket": bucket
                })
        
        # Return the simulated response
        return {
            "currentPath": path,
            "parentPath": "/".join(path.split("/")[:-1]) if path else "",
            "items": items
        }
    
    def get_image_presigned_url(self, key: str) -> Optional[Dict[str, Any]]:
        """
        Get pre-signed URL for image.
        
        Args:
            key: S3 key for the image
            
        Returns:
            Optional[Dict[str, Any]]: Pre-signed URL data or None if request failed
        """
        logger.info(f"Getting pre-signed URL for image: {key}")
        
        # Simulate pre-signed URL for development - REMOVE in production
        # In production, uncomment the API call and remove this simulation
        """
        try:
            response = requests.get(f"{self.api_endpoint}/api/v1/images/{key}/view")
            if response.status_code == 200:
                logger.info("Pre-signed URL generated successfully")
                return response.json()
            else:
                logger.error(f"API Error: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            logger.error(f"Error getting pre-signed URL: {str(e)}")
            return None
        """
        
        # Simulation for development - REMOVE in production
        # Placeholder image URLs for testing
        sample_images = [
            "https://via.placeholder.com/600x800?text=Vending+Machine+Image",
            "https://via.placeholder.com/800x600?text=Reference+Layout",
            "https://via.placeholder.com/600x800?text=Product+Arrangement"
        ]
        import random
        import hashlib
        
        # Use key to generate a consistent image selection
        hash_val = int(hashlib.md5(key.encode()).hexdigest(), 16)
        image_url = sample_images[hash_val % len(sample_images)]
        
        return {
            "presignedUrl": image_url,
            "expiresAt": "2025-05-02T16:00:00Z",
            "contentType": "image/jpeg" if ".jpg" in key else "image/png",
            "metadata": {
                "imageType": "reference" if "reference" in key else "checking",
                "fileName": key.split("/")[-1],
                "size": 245678
            }
        }
    
    def check_health(self) -> Tuple[bool, dict, int]:
        """
        Check health of the API.
        
        Returns:
            Tuple[bool, dict, int]: Success status, response data, and status code
        """
        logger.info("Checking API health")
        try:
            response = requests.get(f"{self.api_endpoint}/api/v1/health", timeout=5)
            if response.status_code == 200:
                data = response.json() if response.content else {"status": "OK"}
                logger.info(f"API health check successful: {response.status_code}")
                return True, data, response.status_code
            else:
                logger.warning(f"API health check failed: {response.status_code}")
                return False, {}, response.status_code
        except Exception as e:
            logger.error(f"Health check failed: {str(e)}")
            return False, {"error": str(e)}, 0