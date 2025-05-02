"""
Images Page for Vending Machine Verification System

This module contains the improved images page implementation with a visual browser.
"""

import streamlit as st
import pandas as pd
import logging
from typing import Optional, Dict, Any, List
from pages import BasePage
from api import APIClient
from utils import format_timestamp, format_file_size

logger = logging.getLogger("streamlit-app")

class ImagesPage(BasePage):
    """Page for browsing and managing images with improved UX."""
    
    def render(self) -> None:
        """Render the images page."""
        st.header("Image Browser")
        
        # Add bucket type selector with tabs
        bucket_tab, _ = st.tabs(["Browse Images", "Upload Images (Coming Soon)"])
        
        with bucket_tab:
            # Add bucket type selector
            st.markdown("### Select Bucket Type")
            bucket_type = st.radio(
                "",
                ["Reference", "Checking"],
                horizontal=True,
                key="bucket_type_images"
            )
            
            # Display image browser
            self._render_image_browser(bucket_type.lower())
    
    def _render_image_browser(self, bucket_type: str) -> None:
        """
        Render improved image browser with visual grid layout.
        
        Args:
            bucket_type: Type of bucket to browse ("reference" or "checking")
        """
        # Current path state for navigation
        current_path_key = f"{bucket_type}_images_current_path"
        if current_path_key not in st.session_state:
            st.session_state[current_path_key] = ""
        
        # Search and filters in a sidebar
        with st.sidebar:
            st.subheader(f"{bucket_type.capitalize()} Images")
            
            # Add search
            search_term = st.text_input("Search images", key=f"{bucket_type}_search_sidebar", 
                                      placeholder="Enter filename or filter")
            
            # Add filters with expander
            with st.expander("Advanced Filters"):
                st.date_input("Date Range (From)", value=None, key=f"{bucket_type}_date_from")
                st.date_input("Date Range (To)", value=None, key=f"{bucket_type}_date_to")
                st.multiselect("File Type", ["PNG", "JPG", "JPEG"], key=f"{bucket_type}_file_type")
                st.slider("Size (MB)", 0, 10, (0, 10), key=f"{bucket_type}_size_range")
        
        # Breadcrumb navigation
        col1, col2 = st.columns([3, 1])
        
        with col1:
            path_parts = ["root"] + [p for p in st.session_state[current_path_key].split("/") if p]
            breadcrumb_html = " / ".join([f"<a href='#' id='{i}'>{part}</a>" for i, part in enumerate(path_parts)])
            st.markdown(f"**Location:** {breadcrumb_html}", unsafe_allow_html=True)
        
        with col2:
            if st.session_state[current_path_key]:
                if st.button("⬆️ Up", key=f"{bucket_type}_up_button_images"):
                    # Go up one level
                    parts = st.session_state[current_path_key].split("/")
                    st.session_state[current_path_key] = "/".join(parts[:-1])
                    st.experimental_rerun()
        
        # Browse the bucket
        try:
            # Get files and folders in current path
            items = self.api_client.browse_images(st.session_state[current_path_key], bucket_type)
            
            if not items or "items" not in items:
                st.info(f"No items found in {bucket_type} bucket at this path")
                return
            
            # Apply search filter if provided
            filtered_items = items["items"]
            if search_term:
                filtered_items = [item for item in items["items"] 
                                 if search_term.lower() in item.get("name", "").lower()]
                
                if not filtered_items:
                    st.info(f"No items matching '{search_term}'")
                    return
            
            # Separate folders and images
            folders = [item for item in filtered_items if item.get("type") == "folder"]
            images = [item for item in filtered_items if item.get("type") == "image"]
            
            # Show summary statistics
            stats_col1, stats_col2, stats_col3 = st.columns(3)
            stats_col1.metric("Folders", len(folders))
            stats_col2.metric("Images", len(images))
            
            total_size = sum(img.get("size", 0) for img in images)
            stats_col3.metric("Total Size", format_file_size(total_size))
            
            # Display folders with a grid layout
            if folders:
                st.markdown("### Folders")
                
                # Create a grid layout for folders
                folder_cols = st.columns(4)
                for i, folder in enumerate(folders):
                    with folder_cols[i % 4]:
                        # Create a card-like display for each folder
                        with st.container():
                            st.markdown(f"""
                            <div style="padding: 10px; border: 1px solid #ddd; border-radius: 5px; text-align: center;">
                                <h4>📁 {folder['name']}</h4>
                            </div>
                            """, unsafe_allow_html=True)
                            
                            if st.button(f"Open", key=f"{bucket_type}_folder_open_{i}"):
                                # Navigate into folder
                                st.session_state[current_path_key] = folder["path"]
                                st.experimental_rerun()
            
            # Display images with a visually appealing grid
            if images:
                st.markdown("### Images")
                
                # Add view mode selector
                view_mode = st.radio("View Mode", ["Grid", "List"], horizontal=True, index=0, key=f"{bucket_type}_view_mode")
                
                if view_mode == "Grid":
                    # Grid view with thumbnails
                    image_cols = st.columns(3)
                    for i, image in enumerate(images):
                        with image_cols[i % 3]:
                            # Get a presigned URL for the image thumbnail
                            presigned_url_data = self.api_client.get_image_presigned_url(image["path"])
                            if presigned_url_data and "presignedUrl" in presigned_url_data:
                                # Show thumbnail with metadata in a card
                                with st.container():
                                    st.image(presigned_url_data["presignedUrl"], caption=image["name"], use_column_width=True)
                                    st.markdown(f"""
                                    <div style="font-size:0.8em;">
                                        Size: {format_file_size(image.get("size", 0))}<br>
                                        Last Modified: {format_timestamp(image.get("lastModified", ""))}
                                    </div>
                                    """, unsafe_allow_html=True)
                                    
                                    # Add action buttons
                                    btn_col1, btn_col2 = st.columns(2)
                                    with btn_col1:
                                        if st.button("View", key=f"{bucket_type}_view_{i}"):
                                            self._view_image_detail(image, presigned_url_data["presignedUrl"])
                                    with btn_col2:
                                        st.download_button(
                                            "Download", 
                                            data="Placeholder", 
                                            file_name=image["name"],
                                            key=f"{bucket_type}_download_{i}"
                                        )
                            else:
                                # Fallback if we can't get the image
                                st.info(f"📷 {image['name']} (Preview not available)")
                else:
                    # List view with details
                    for i, image in enumerate(images):
                        with st.expander(f"{image['name']} ({format_file_size(image.get('size', 0))})"):
                            col1, col2 = st.columns([1, 2])
                            
                            with col1:
                                # Get a presigned URL for the image thumbnail
                                presigned_url_data = self.api_client.get_image_presigned_url(image["path"])
                                if presigned_url_data and "presignedUrl" in presigned_url_data:
                                    st.image(presigned_url_data["presignedUrl"], width=200)
                                else:
                                    st.info("Preview not available")
                            
                            with col2:
                                st.markdown(f"""
                                **Path:** {image['path']}  
                                **Size:** {format_file_size(image.get('size', 0))}  
                                **Last Modified:** {format_timestamp(image.get('lastModified', ''))}  
                                **Bucket:** {image.get('bucket', '')}
                                """)
                                
                                # Add action buttons
                                btn_col1, btn_col2 = st.columns(2)
                                with btn_col1:
                                    if st.button("View", key=f"{bucket_type}_list_view_{i}"):
                                        if presigned_url_data and "presignedUrl" in presigned_url_data:
                                            self._view_image_detail(image, presigned_url_data["presignedUrl"])
                                with btn_col2:
                                    st.download_button(
                                        "Download", 
                                        data="Placeholder", 
                                        file_name=image["name"],
                                        key=f"{bucket_type}_list_download_{i}"
                                    )
            
            if not folders and not images:
                st.info("No folders or images found at this location")
                
        except Exception as e:
            st.error(f"Error browsing {bucket_type} bucket: {str(e)}")
            logger.error(f"Error in image_browser: {str(e)}")
    
    def _view_image_detail(self, image: Dict[str, Any], image_url: str) -> None:
        """
        Display detailed view of an image.
        
        Args:
            image: Image metadata
            image_url: URL to the image
        """
        st.session_state["viewed_image"] = {
            "metadata": image,
            "url": image_url
        }
        
        # Create a modal-like experience using an expander that starts expanded
        with st.expander("Image Details", expanded=True):
            col1, col2 = st.columns([2, 1])
            
            with col1:
                st.image(image_url, use_column_width=True)
            
            with col2:
                st.markdown(f"### {image['name']}")
                st.markdown(f"""
                **Path:** {image['path']}  
                **Size:** {format_file_size(image.get('size', 0))}  
                **Last Modified:** {format_timestamp(image.get('lastModified', ''))}  
                **Bucket:** {image.get('bucket', '')}
                """)
                
                # Add action buttons
                st.download_button(
                    "Download Image", 
                    data="Placeholder", 
                    file_name=image["name"],
                    use_container_width=True
                )
                
                # Add options for verification
                st.markdown("### Use for Verification")
                if st.button("Use as Reference Image", use_container_width=True):
                    st.session_state["reference_image"] = image
                    st.session_state["reference_image"]["presignedUrl"] = image_url
                    st.success(f"Set as reference image: {image['name']}")
                
                if st.button("Use as Checking Image", use_container_width=True):
                    st.session_state["checking_image"] = image
                    st.session_state["checking_image"]["presignedUrl"] = image_url
                    st.success(f"Set as checking image: {image['name']}")
                
                if st.button("Start New Comparison", use_container_width=True):
                    # Navigate to comparison page
                    st.session_state["page"] = "New Comparison"
                    st.experimental_rerun()