import streamlit as st

# Set page configuration FIRST - before any other Streamlit commands
st.set_page_config(page_title="Vending Machine Verification", layout="wide")

import logging
# from dotenv import load_dotenv # Removed for containerized deployment
from pages import home, verifications
from pages import health_check, verification_lookup, image_upload
from clients.api_client import APIClient

# Load environment variables from .env file - Removed as env vars will be set by Docker/ECS
# load_dotenv() # Removed

# Configuration validation will be handled by the APIClient and ConfigLoader
# This allows for both CONFIG_SECRET and individual environment variable approaches

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# Initialize API client
# The APIClient will load configuration from AWS Secrets Manager (CONFIG_SECRET)
# or fall back to individual environment variables for backward compatibility.
try:
    api_client = APIClient()
except ValueError as e:
    st.error(f"Error initializing API Client: {e}")
    st.stop() # Stop the app if API client can't be initialized
except Exception as e:
    st.error(f"An unexpected error occurred during API Client initialization: {e}")
    logger.error(f"Unexpected API Client initialization error: {e}", exc_info=True)
    st.stop()

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
    "Verification Results": {"module": verifications, "icon": "📋", "category": "Verification"},
    "Verification Lookup": {"module": verification_lookup, "icon": "🔍", "category": "Verification"},
    "Image Upload": {"module": image_upload, "icon": "📤", "category": "Tools"},
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
    logger.info(f"🔄 Loading page: {selection}")

    if selection not in pages:
        logger.warning(f"Page '{selection}' not found in pages dict")
        st.error(f"Page '{selection}' not found. Defaulting to Home.")
        st.session_state['page'] = 'Home'
        selection = 'Home'

    page_module = pages[selection]["module"]
    logger.info(f"📦 Page module loaded: {page_module.__name__}")

    # Ensure api_client was successfully initialized before passing it
    if 'api_client' in locals() and api_client is not None:
        logger.info(f"🔗 Calling {selection} page with API client")
        page_module.app(api_client)
        logger.info(f"✅ {selection} page loaded successfully")
    else:
        # This case should ideally be caught by the st.stop() earlier
        logger.error("API Client is not available")
        st.error("API Client is not available. Cannot load page.")
except Exception as e:
    logger.error(f"Error loading page {selection}: {str(e)}", exc_info=True)
    st.error(f"An error occurred while loading the page '{selection}'. Please try again or contact support.")

    # Show additional debug information
    with st.expander("🔧 Debug Information", expanded=False):
        st.error(f"**Error Type**: {type(e).__name__}")
        st.error(f"**Error Message**: {str(e)}")
        st.code(f"Page: {selection}\nModule: {pages.get(selection, {}).get('module', 'Unknown')}", language="text")

    # Optionally, redirect to a safe page like Home on error
    # st.session_state['page'] = 'Home'
    # st.rerun()
