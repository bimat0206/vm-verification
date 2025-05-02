"""
Dashboard Page for Vending Machine Verification System

This module contains the dashboard page implementation.
"""

import streamlit as st
import logging
from pages import BasePage
from api import APIClient

logger = logging.getLogger("streamlit-app")

class DashboardPage(BasePage):
    """Dashboard page showing system overview."""
    
    def render(self) -> None:
        """Render the dashboard page."""
        st.header("Dashboard")
        
        col1, col2 = st.columns(2)
        
        with col1:
            self._render_recent_comparisons()
        
        with col2:
            self._render_machines_overview()
    
    def _render_recent_comparisons(self) -> None:
        """Render the recent comparisons section."""
        st.subheader("Recent Comparisons")
        try:
            # This would need a specific API endpoint to list recent comparisons
            # For now, we'll just show a placeholder
            st.info("API endpoint for listing recent comparisons not yet implemented")
            st.markdown("""
            | Comparison ID | Machine ID | Status | Date |
            | ------------- | ---------- | ------ | ---- |
            | comp-001 | vm-123 | Completed | 2025-05-12 |
            | comp-002 | vm-456 | In Progress | 2025-05-11 |
            | comp-003 | vm-123 | Failed | 2025-05-10 |
            """)
        except Exception as e:
            st.error(f"Error fetching recent comparisons: {str(e)}")
    
    def _render_machines_overview(self) -> None:
        """Render the machines overview section."""
        st.subheader("Machines Overview")
        try:
            # This would need a specific API endpoint to get machines overview
            # For now, we'll just show a placeholder
            st.info("API endpoint for machines overview not yet implemented")
            st.markdown("""
            | Machine ID | Location | Last Verified | Status |
            | ---------- | -------- | ------------- | ------ |
            | vm-123 | Building A | 2025-05-12 | OK |
            | vm-456 | Building B | 2025-05-11 | Pending |
            | vm-789 | Building C | 2025-05-09 | Issue Detected |
            """)
        except Exception as e:
            st.error(f"Error fetching machines overview: {str(e)}")