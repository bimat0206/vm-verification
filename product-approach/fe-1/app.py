import streamlit as st
import os
import logging
from pages import home, initiate_verification, verifications, verification_details, image_browser, health_check
from clients.api_client import APIClient

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Initialize API client
api_client = APIClient()

# Set page configuration
st.set_page_config(page_title="Vending Machine Verification", layout="wide")

# Define pages
pages = {
    "Home": home,
    "Initiate Verification": initiate_verification,
    "Verifications": verifications,
    "Verification Details": verification_details,
    "Image Browser": image_browser,
    "Health Check": health_check,
}

# Sidebar navigation
st.sidebar.title("Navigation")
selection = st.sidebar.selectbox("Go to", list(pages.keys()))

# Load selected page
try:
    page = pages[selection]
    page.app(api_client)
except Exception as e:
    logger.error(f"Error loading page {selection}: {str(e)}")
    st.error("An error occurred. Please try again or contact support.")