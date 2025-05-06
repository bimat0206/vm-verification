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

# Hide the default Streamlit navigation elements
hide_streamlit_style = """
<style>
    /* Hide the top navigation links in sidebar */
    div[data-testid="stSidebarNav"] {
        display: none !important;
    }
    
    /* Hide the app name in sidebar */
    div.sidebar-content div:first-child {
        display: none !important;
    }
    
    /* Hide the hamburger menu */
    section[data-testid="stSidebarUserContent"] > div:first-child {
        display: none !important;
    }
</style>
"""
st.markdown(hide_streamlit_style, unsafe_allow_html=True)

# Define pages with categories and icons
pages = {
    "Home": {"module": home, "icon": "🏠", "category": "Main"},
    "Initiate Verification": {"module": initiate_verification, "icon": "▶️", "category": "Verification"},
    "Verifications": {"module": verifications, "icon": "📋", "category": "Verification"},
    "Verification Details": {"module": verification_details, "icon": "🔍", "category": "Verification"},
    "Verification Lookup": {"module": verification_lookup, "icon": "🔎", "category": "Verification"},
    "Image Browser": {"module": image_browser, "icon": "🖼️", "category": "Tools"},
    "Health Check": {"module": health_check, "icon": "❤️", "category": "Tools"},
}

# Sidebar navigation
st.sidebar.title("Vending Machine Verification")
st.sidebar.markdown("---")

# Group pages by category
categories = {}
for page_name, page_info in pages.items():
    category = page_info["category"]
    if category not in categories:
        categories[category] = []
    categories[category].append(page_name)

# Use session state to keep track of current page
current_page = st.session_state.get('page', 'Home')

# Create a more compact sidebar layout
with st.sidebar:
    # Display navigation by category
    for category, category_pages in categories.items():
        st.subheader(category)
        
        # Create columns for more compact layout
        cols = st.columns(1)
        with cols[0]:
            for page_name in category_pages:
                page_info = pages[page_name]
                if st.button(f"{page_info['icon']} {page_name}", key=page_name, 
                            help=f"Navigate to {page_name}",
                            use_container_width=True,
                            type="primary" if page_name == current_page else "secondary"):
                    st.session_state['page'] = page_name
                    st.rerun()
        
        st.markdown("---")

# Get the current page from session state
selection = st.session_state.get('page', 'Home')

# Load selected page
try:
    page_module = pages[selection]["module"]
    page_module.app(api_client)
except Exception as e:
    logger.error(f"Error loading page {selection}: {str(e)}")
    st.error("An error occurred. Please try again or contact support.")
