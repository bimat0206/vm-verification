import streamlit as st
import os
import logging
# from dotenv import load_dotenv # Removed for containerized deployment
from pages import home, initiate_verification, verifications, verification_details
from pages import image_browser, health_check, verification_lookup
from clients.api_client import APIClient

# Load environment variables from .env file - Removed as env vars will be set by Docker/ECS
# load_dotenv() # Removed

# Verify required environment variables
if not os.getenv('API_ENDPOINT'):
    # Updated error message for clarity in containerized environments
    raise ValueError(
        "API_ENDPOINT environment variable is required. "
        "Ensure it's set in the Docker environment (via ARG/ENV) "
        "or available in Streamlit secrets (secrets.toml)."
    )

if not (os.getenv('API_KEY') or os.getenv('API_KEY_SECRET_NAME')):
    # Updated error message for clarity
    raise ValueError(
        "Either API_KEY or API_KEY_SECRET_NAME must be set as an environment variable "
        "(via Docker ARG/ENV or ECS task definition) or available in Streamlit secrets (secrets.toml)."
    )

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Initialize API client
# The APIClient will attempt to get API_ENDPOINT and API_KEY from os.environ first,
# then fall back to st.secrets if available.
try:
    api_client = APIClient()
except ValueError as e:
    st.error(f"Error initializing API Client: {e}")
    st.stop() # Stop the app if API client can't be initialized
except Exception as e:
    st.error(f"An unexpected error occurred during API Client initialization: {e}")
    logger.error(f"Unexpected API Client initialization error: {e}", exc_info=True)
    st.stop()


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
    /* div.sidebar-content div:first-child {
        display: none !important;
    } */
    
    /* Hide the hamburger menu */
    /* section[data-testid="stSidebarUserContent"] > div:first-child {
        display: none !important;
    } */
</style>
"""
# Note: Hiding sidebar-content first-child and stSidebarUserContent first-child
# can sometimes hide the entire sidebar if not careful with Streamlit versions or custom components.
# If the sidebar title "Vending Machine Verification" is desired, these specific rules might need adjustment.
# Keeping them commented out for now as they are aggressive. The stSidebarNav hide is usually sufficient.
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
        # Using 1 column is effectively no column change, but keeps structure if you want to expand to 2 later
        cols = st.columns(1)
        with cols[0]:
            for page_name in category_pages:
                page_info = pages[page_name]
                # Ensure the key is unique and descriptive for Streamlit's widget management
                button_key = f"nav_button_{page_name.lower().replace(' ', '_')}"
                if st.button(f"{page_info['icon']} {page_name}", key=button_key,
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
    if selection not in pages:
        st.error(f"Page '{selection}' not found. Defaulting to Home.")
        st.session_state['page'] = 'Home'
        selection = 'Home'

    page_module = pages[selection]["module"]
    # Ensure api_client was successfully initialized before passing it
    if 'api_client' in locals() and api_client is not None:
        page_module.app(api_client)
    else:
        # This case should ideally be caught by the st.stop() earlier
        st.error("API Client is not available. Cannot load page.")
except Exception as e:
    logger.error(f"Error loading page {selection}: {str(e)}", exc_info=True)
    st.error(f"An error occurred while loading the page '{selection}'. Please try again or contact support.")
    # Optionally, redirect to a safe page like Home on error
    # st.session_state['page'] = 'Home'
    # st.rerun()
