"""
Pages Package for Vending Machine Verification System

This package contains the different pages used in the Streamlit application.
"""

# Import for proper package recognition
import logging
from api import APIClient

logger = logging.getLogger("streamlit-app")

# Base Page class that all pages inherit from
class BasePage:
    """Base class for all application pages."""
    
    def __init__(self, api_client: APIClient):
        self.api_client = api_client
    
    def render(self) -> None:
        """
        Render the page content.
        This should be implemented by subclasses.
        """
        raise NotImplementedError("Subclasses must implement render()")