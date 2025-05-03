import streamlit as st
import os
import logging
from dotenv import load_dotenv
from pages import home, initiate_verification, verifications, verification_details
from pages import image_browser, health_check, verification_lookup
from clients.api_client import APIClient

# Load environment variables from .env file
load_dotenv()

# Verify required environment variables
if not os.getenv('API_ENDPOINT'):
    raise ValueError("API_ENDPOINT environment variable is required. Please set it in your .env file.")

if not (os.getenv('API_KEY') or os.getenv('API_KEY_SECRET_NAME')):
    raise ValueError("Either API_KEY or API_KEY_SECRET_NAME must be set in your .env file.")

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
    "Verification Lookup": verification_lookup,
    "Image Browser": image_browser,
    "Health Check": health_check,
}

# Sidebar navigation
st.sidebar.title("Navigation")
# Use session state to keep track of current page
current_page = st.session_state.get('page', 'Home')
selection = st.sidebar.selectbox("Go to", list(pages.keys()), index=list(pages.keys()).index(current_page))

# Update session state when selection changes
if selection != current_page:
    st.session_state['page'] = selection

# Load selected page
try:
    page = pages[selection]
    page.app(api_client)
except Exception as e:
    logger.error(f"Error loading page {selection}: {str(e)}")
    st.error("An error occurred. Please try again or contact support.")