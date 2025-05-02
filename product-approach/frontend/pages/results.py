"""
View Results Page for Vending Machine Verification System

This module contains the view results page implementation.
"""

import streamlit as st
import logging
from typing import Dict, Any
from pages import BasePage
from api import APIClient
from utils import format_timestamp, generate_color_for_status, safe_get

logger = logging.getLogger("streamlit-app")

class ViewResultsPage(BasePage):
    """Page for viewing verification results."""
    
    def render(self) -> None:
        """Render the view results page."""
        st.header("View Verification Results")
        
        # Get comparison ID from URL or input
        comparison_id = st.text_input(
            "Enter Comparison ID",
            value=st.session_state.get("comparison_id", ""),
            help="ID of the comparison to view"
        )
        
        if st.button("Fetch Results") and comparison_id:
            self._fetch_and_display_results(comparison_id)
    
    def _fetch_and_display_results(self, comparison_id: str) -> None:
        """Fetch and display results for a comparison."""
        with st.spinner("Fetching results..."):
            results = self.api_client.get_comparison(comparison_id)
            if results:
                st.success("Results retrieved successfully!")
                
                # Display comparison details
                st.subheader("Comparison Details")
                if isinstance(results, dict):
                    self._render_comparison_header(results)
                    self._render_verification_results(results)
                    
                    # Full JSON response (expandable)
                    with st.expander("View Raw JSON Response"):
                        st.json(results)
                else:
                    st.write(results)
            else:
                st.warning("No results found for this comparison ID")
    
    def _render_comparison_header(self, results: Dict[str, Any]) -> None:
        """Render the comparison header with key metrics."""
        cols = st.columns(3)
        
        # Extract values with safe_get to handle missing keys
        machine_id = safe_get(results, "vendingMachineId", default="N/A")
        status = safe_get(results, "status", default="N/A")
        timestamp = format_timestamp(safe_get(results, "verificationAt", default="N/A"))
        
        cols[0].metric("Machine ID", machine_id)
        cols[1].metric("Status", status)
        cols[2].metric("Timestamp", timestamp)
        
        # Add visual status indicator
        status_color = generate_color_for_status(status)
        st.markdown(
            f"<div style='background-color: {status_color}; padding: 10px; border-radius: 5px; "
            f"color: white; margin-bottom: 20px;'>Status: {status}</div>",
            unsafe_allow_html=True
        )
    
    def _render_verification_results(self, results: Dict[str, Any]) -> None:
        """Render the verification results section."""
        # Display comparison results
        st.subheader("Verification Results")
        
        # If there are detailed results
        if "verificationResults" in results:
            verification = results["verificationResults"]
            
            # Show summary statistics
            if "verificationSummary" in verification:
                summary = verification["verificationSummary"]
                
                col1, col2, col3, col4 = st.columns(4)
                col1.metric("Total Positions", safe_get(summary, "totalPositionsChecked", default=0))
                col2.metric("Correct Positions", safe_get(summary, "correctPositions", default=0))
                col3.metric("Discrepancies", safe_get(summary, "discrepantPositions", default=0))
                col4.metric("Accuracy", f"{safe_get(summary, 'overallAccuracy', default=0)}%")
                
                st.info(safe_get(summary, "verificationOutcome", default="No summary available"))
            
            # Show discrepancies if any
            if "discrepancies" in verification and verification["discrepancies"]:
                st.subheader("Discrepancies Found")
                
                for idx, disc in enumerate(verification["discrepancies"], 1):
                    st.markdown(
                        f"**{idx}. Position {disc.get('position', 'Unknown')}:** "
                        f"{disc.get('expected', 'Unknown')} → {disc.get('found', 'Unknown')} "
                        f"({disc.get('issue', 'Unknown issue')})"
                    )
            
            # Show result image if available
            if "resultImageUrl" in results:
                st.subheader("Verification Image")
                st.image(results["resultImageUrl"], caption="Verification Result", use_column_width=True)
        else:
            st.write("No detailed verification results available")