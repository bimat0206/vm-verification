"""
New Comparison Page for Vending Machine Verification System

This module contains the new comparison page implementation with improved UI/UX
and S3 image browser functionality.
"""

import streamlit as st
import logging
from typing import Optional, Dict, Any, List
from pages import BasePage
from api import APIClient
from utils import format_timestamp, format_file_size, parse_s3_url

logger = logging.getLogger("streamlit-app")

class NewComparisonPage(BasePage):
    """Page for creating new verification comparisons with improved UX."""
    
    def __init__(self, api_client: APIClient):
        super().__init__(api_client)
        # Initialize state for selected images
        if "reference_image" not in st.session_state:
            st.session_state.reference_image = None
        if "checking_image" not in st.session_state:
            st.session_state.checking_image = None
    
    def render(self) -> None:
        """Render the new comparison page."""
        st.header("Create New Comparison")
        
        # Add verification type selector with improved styling
        st.markdown("### Select Verification Type")
        verification_type = st.radio(
            "",
            ["Layout vs Checking", "Previous vs Current"],
            captions=["Compare against reference layout", "Compare against previous state"],
            horizontal=True,
            key="verification_type"
        )
        
        # Use tabs for a cleaner layout
        ref_tab, check_tab, details_tab = st.tabs(["Reference Image", "Checking Image", "Verification Details"])
        
        with ref_tab:
            self._render_reference_image_selector(verification_type)
        
        with check_tab:
            self._render_checking_image_selector()
        
        with details_tab:
            self._render_verification_details()
        
        # Display selected images summary
        self._render_image_summary()
        
        # Submit button with improved styling
        if st.button("Start Comparison", type="primary", use_container_width=True, 
                    disabled=not (st.session_state.reference_image and 
                                st.session_state.checking_image and 
                                "vending_machine_id" in st.session_state and 
                                st.session_state.vending_machine_id)):
            self._handle_form_submission()
    
    def _render_reference_image_selector(self, verification_type: str) -> None:
        """Render the reference image selector tab."""
        if verification_type == "Layout vs Checking":
            st.subheader("Select Reference Layout Image")
            bucket_type = "reference"
            help_text = "Select a reference layout image from the reference bucket"
        else:
            st.subheader("Select Previous Checking Image")
            bucket_type = "checking"
            help_text = "Select a previous checking image from the checking bucket"
        
        st.markdown(help_text)
        
        # Image browser for selecting reference image
        selected_image = self._image_browser(bucket_type)
        
        if selected_image:
            st.session_state.reference_image = selected_image
            st.success(f"Reference image selected: {selected_image['path']}")
            
            # Preview the selected image
            if "presignedUrl" in selected_image:
                st.image(selected_image["presignedUrl"], caption="Selected Reference Image", use_column_width=True)
    
    def _render_checking_image_selector(self) -> None:
        """Render the checking image selector tab."""
        st.subheader("Select Checking Image")
        st.markdown("Select the current checking image from the checking bucket")
        
        # Image browser for selecting checking image
        selected_image = self._image_browser("checking")
        
        if selected_image:
            st.session_state.checking_image = selected_image
            st.success(f"Checking image selected: {selected_image['path']}")
            
            # Preview the selected image
            if "presignedUrl" in selected_image:
                st.image(selected_image["presignedUrl"], caption="Selected Checking Image", use_column_width=True)
    
    def _render_verification_details(self) -> None:
        """Render the verification details tab."""
        st.subheader("Verification Details")
        
        # Vending Machine ID
        vending_machine_id = st.text_input(
            "Vending Machine ID", 
            value=st.session_state.get("vending_machine_id", ""),
            help="Unique identifier for the vending machine",
            key="vending_machine_id"
        )
        
        # Location (optional)
        location = st.text_input(
            "Location (Optional)",
            value=st.session_state.get("location", ""),
            help="Physical location of the vending machine",
            key="location"
        )
        
        # Additional options with expander
        with st.expander("Advanced Options"):
            st.checkbox("Enable Notifications", value=False, key="notifications_enabled", 
                       help="Send notifications when verification completes")
            st.select_slider("Confidence Threshold", options=["Low", "Medium", "High"], 
                           value="Medium", key="confidence_threshold",
                           help="Set the confidence threshold for discrepancy detection")
    
    def _render_image_summary(self) -> None:
        """Render a summary of the selected images for verification."""
        st.markdown("---")
        st.markdown("### Verification Summary")
        
        col1, col2 = st.columns(2)
        
        with col1:
            st.markdown("**Reference Image:**")
            if st.session_state.reference_image:
                st.code(st.session_state.reference_image["path"], language=None)
            else:
                st.info("No reference image selected")
        
        with col2:
            st.markdown("**Checking Image:**")
            if st.session_state.checking_image:
                st.code(st.session_state.checking_image["path"], language=None)
            else:
                st.info("No checking image selected")
        
        # Display validation messages if needed
        if not (st.session_state.reference_image and st.session_state.checking_image):
            st.warning("Please select both reference and checking images")
        
        if not st.session_state.get("vending_machine_id"):
            st.warning("Please enter a Vending Machine ID")
    
    def _image_browser(self, bucket_type: str) -> Optional[Dict[str, Any]]:
        """
        Enhanced S3 image browser component.
        
        Args:
            bucket_type: Type of bucket to browse ("reference" or "checking")
            
        Returns:
            Optional[Dict[str, Any]]: Selected image metadata or None
        """
        # Current path state for navigation
        current_path_key = f"{bucket_type}_current_path"
        if current_path_key not in st.session_state:
            st.session_state[current_path_key] = ""
        
        # Breadcrumb navigation
        path_parts = ["root"] + [p for p in st.session_state[current_path_key].split("/") if p]
        breadcrumb_html = " / ".join([f"<a href='#' id='{i}'>{part}</a>" for i, part in enumerate(path_parts)])
        st.markdown(f"**Location:** {breadcrumb_html}", unsafe_allow_html=True)
        
        # Up button if not at root
        col1, col2 = st.columns([1, 5])
        with col1:
            if st.session_state[current_path_key]:
                if st.button("⬆️ Up", key=f"{bucket_type}_up_button"):
                    # Go up one level
                    parts = st.session_state[current_path_key].split("/")
                    st.session_state[current_path_key] = "/".join(parts[:-1])
                    st.experimental_rerun()
        
        # Search or filter
        with col2:
            search_term = st.text_input("Search or filter", key=f"{bucket_type}_search", 
                                      placeholder="Enter filename or filter")
        
        # Browse the bucket
        try:
            # Get files and folders in current path
            items = self.api_client.browse_images(st.session_state[current_path_key], bucket_type)
            
            if not items or "items" not in items:
                st.info(f"No items found in {bucket_type} bucket at this path")
                return None
            
            # Apply search filter if provided
            filtered_items = items["items"]
            if search_term:
                filtered_items = [item for item in items["items"] 
                                 if search_term.lower() in item.get("name", "").lower()]
                
                if not filtered_items:
                    st.info(f"No items matching '{search_term}'")
                    return None
            
            # Separate folders and images
            folders = [item for item in filtered_items if item.get("type") == "folder"]
            images = [item for item in filtered_items if item.get("type") == "image"]
            
            # Display folders first with thumbnails
            if folders:
                st.markdown("#### Folders")
                folder_cols = st.columns(3)
                for i, folder in enumerate(folders):
                    with folder_cols[i % 3]:
                        if st.button(f"📁 {folder['name']}", key=f"{bucket_type}_folder_{i}"):
                            # Navigate into folder
                            st.session_state[current_path_key] = folder["path"]
                            st.experimental_rerun()
            
            # Display images with thumbnails and selection
            if images:
                st.markdown("#### Images")
                selected_image = None
                
                # Create a grid layout for images
                image_cols = st.columns(3)
                for i, image in enumerate(images):
                    with image_cols[i % 3]:
                        # Get a presigned URL for the image thumbnail
                        presigned_url_data = self.api_client.get_image_presigned_url(image["path"])
                        if presigned_url_data and "presignedUrl" in presigned_url_data:
                            # Show thumbnail
                            st.image(presigned_url_data["presignedUrl"], caption=image["name"], width=150)
                            # Add selection button
                            if st.button(f"Select", key=f"{bucket_type}_select_{i}"):
                                # Add the presigned URL to the image data
                                image["presignedUrl"] = presigned_url_data["presignedUrl"]
                                return image
                        else:
                            # Fallback if we can't get the image
                            st.info(f"📷 {image['name']}")
                            if st.button(f"Select (no preview)", key=f"{bucket_type}_select_np_{i}"):
                                return image
            
            if not folders and not images:
                st.info("No folders or images found at this location")
            
            return None
            
        except Exception as e:
            st.error(f"Error browsing {bucket_type} bucket: {str(e)}")
            logger.error(f"Error in image_browser: {str(e)}")
            return None
    
    def _handle_form_submission(self) -> None:
        """Handle form submission for new comparison."""
        verification_type = st.session_state.get("verification_type", "Layout vs Checking")
        reference_image = st.session_state.reference_image
        checking_image = st.session_state.checking_image
        vending_machine_id = st.session_state.vending_machine_id
        location = st.session_state.get("location", "")
        
        if not (reference_image and checking_image and vending_machine_id):
            st.error("Please fill in all required fields")
            return
        
        # Construct S3 URIs for images
        reference_img_url = f"s3://{reference_image['bucket']}/{reference_image['path']}"
        checking_img_url = f"s3://{checking_image['bucket']}/{checking_image['path']}"
        
        # Map UI selection to API verification type
        api_verification_type = "LAYOUT_VS_CHECKING" if verification_type == "Layout vs Checking" else "PREVIOUS_VS_CURRENT"
        
        # Additional options
        notification_enabled = st.session_state.get("notifications_enabled", False)
            
        with st.spinner("Starting comparison..."):
            # Add notification flag to API call
            result = self.api_client.create_comparison(
                reference_img_url, 
                checking_img_url, 
                vending_machine_id, 
                location, 
                api_verification_type,
                notification_enabled
            )
            
            if result:
                st.success("Comparison started successfully!")
                
                # Show verification details with nice formatting
                st.markdown("### Verification Details")
                
                # Create a nicer result display
                if "executionArn" in result:
                    comparison_id = result["executionArn"].split(":")[-1]
                    st.session_state["last_comparison_id"] = comparison_id
                    
                    # Display result info in a better format
                    col1, col2 = st.columns(2)
                    with col1:
                        st.markdown("**Verification ID:**")
                        st.code(comparison_id)
                        
                        st.markdown("**Status:**")
                        st.code(result.get("status", "PROCESSING"))
                    
                    with col2:
                        st.markdown("**Started At:**")
                        st.code(format_timestamp(result.get("verificationAt", "")))
                        
                        st.markdown("**Machine ID:**")
                        st.code(vending_machine_id)
                    
                    # Add a nice view results button
                    if st.button("View Results", type="primary", use_container_width=True):
                        st.session_state["page"] = "View Results"
                        st.session_state["comparison_id"] = comparison_id
                        st.experimental_rerun()
            else:
                st.error("Failed to start comparison. Please check the logs for details.")