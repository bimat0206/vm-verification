"""
Vending Machine Verification System - Streamlit Frontend (Main Application)

This is the main entry point for the Streamlit application that provides the
frontend interface for the vending machine verification system.
"""

import logging
import streamlit as st
from datetime import datetime

from config import ConfigManager
from api import APIClient
# Update imports to use separate page modules
from pages.dashboard import DashboardPage
from pages.comparison import NewComparisonPage
from pages.results import ViewResultsPage
from pages.images import ImagesPage
from pages.system_info import SystemInfoPage
from utils import setup_logging

# Configure logging
logger = setup_logging()

class App:
    """Main Streamlit application class."""
    
    def __init__(self):
        self.config_manager = ConfigManager()
        self.config = {}
        self.api_client = None
    
    def initialize(self) -> bool:
        """
        Initialize the application.
        
        Returns:
            bool: True if initialization was successful, False otherwise
        """
        try:
            # Load configuration
            logger.info("Application startup: loading configuration")
            self.config = self.config_manager.load_config()
            
            # Initialize API client
            api_endpoint = self.config.get("api_endpoint", "")
            logger.info(f"Using API endpoint: {api_endpoint}")
            
            if not api_endpoint:
                logger.error("No API endpoint configured")
                return False
                
            self.api_client = APIClient(api_endpoint)
            return True
            
        except Exception as e:
            logger.error(f"Error initializing application: {str(e)}", exc_info=True)
            return False
    
    def setup_ui(self) -> None:
        """Set up the Streamlit UI components."""
        # Set page config
        st.set_page_config(
            page_title="Vending Machine Verification",
            page_icon="🏪",
            layout="wide",
            initial_sidebar_state="expanded"
        )

        # Header and description
        st.title("Vending Machine Verification System")
        st.markdown("""
        This application helps you verify the stock and state of vending machines using image comparisons.
        Upload images, request comparisons, and view results.
        """)

        # Display API connection status
        try:
            success, _, status_code = self.api_client.check_health()
            if success:
                st.success(f"✅ Connected to API at {self.config.get('api_endpoint')}")
            else:
                st.warning(f"⚠️ API responded with status code: {status_code}")
        except Exception as e:
            st.error(f"❌ Cannot connect to API: {str(e)}")
            st.info("The application will continue to work, but some features may be limited.")
    
    def run(self) -> None:
        """Run the Streamlit application."""
        # Initialize the application
        if not self.initialize():
            st.error("Failed to initialize application. Please check the logs for details.")
            return
        
        # Set up the UI
        self.setup_ui()
        
        # Sidebar for navigation
        st.sidebar.title("Navigation")
        page = st.sidebar.radio("Go to", ["Dashboard", "New Comparison", "View Results", "Images", "System Info"])
        
        # Create page instances
        dashboard_page = DashboardPage(self.api_client)
        new_comparison_page = NewComparisonPage(self.api_client)
        view_results_page = ViewResultsPage(self.api_client)
        images_page = ImagesPage(self.api_client)
        system_info_page = SystemInfoPage(self.api_client, self.config)
        
        # Render the selected page
        if page == "Dashboard":
            dashboard_page.render()
        elif page == "New Comparison":
            new_comparison_page.render()
        elif page == "View Results":
            view_results_page.render()
        elif page == "Images":
            images_page.render()
        elif page == "System Info":
            system_info_page.render()
        
        # Add footer
        st.markdown("---")
        st.markdown(f"Vending Machine Verification System | {datetime.now().year}")


def main():
    """Entry point for the Streamlit application."""
    try:
        app = App()
        app.run()
    except Exception as e:
        logger.error(f"Unhandled exception in main app: {str(e)}", exc_info=True)
        st.error(f"An unexpected error occurred: {str(e)}")
        st.info("Please check the system logs for more information.")
        
        # Display traceback in development environments
        import traceback
        with st.expander("Error Details"):
            st.code(traceback.format_exc())


if __name__ == "__main__":
    main()