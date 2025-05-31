import streamlit as st
import logging

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Image Browser")
    
    col1, col2 = st.columns(2)
    with col1:
        bucket_type = st.selectbox("Bucket Type", ["reference", "checking"])
    with col2:
        path = st.text_input("Enter path (leave blank for root)", "")
    
    if st.button("Browse"):
        try:
            browser_response = api_client.browse_images(path, bucket_type)
            
            # Display current path and navigation
            current_path = browser_response.get('currentPath', '')
            parent_path = browser_response.get('parentPath', '')
            
            st.write(f"**Current Path:** {current_path}")
            if parent_path:
                if st.button("⬆️ Go Up"):
                    st.session_state['browser_path'] = parent_path
                    st.rerun()
            
            # Display items
            items = browser_response.get('items', [])
            if not items:
                st.info("No items found in this location.")
                return
                
            # Create grid layout for items
            cols = st.columns(4)
            for idx, item in enumerate(items):
                col = cols[idx % 4]
                with col:
                    name = item.get('name', 'Unknown')
                    item_type = item.get('type', 'unknown')
                    
                    if item_type == 'folder':
                        st.button(f"📁 {name}", key=f"folder_{idx}", on_click=lambda p=item.get('path', ''): st.session_state.update({'browser_path': p}))
                    elif item_type == 'image':
                        try:
                            url_response = api_client.get_image_url(item.get('path', ''), bucket_type)
                            image_url = url_response.get('presignedUrl', '')
                            if image_url:
                                st.image(image_url, caption=name, width=150)
                                st.text(f"Last modified: {item.get('lastModified', 'unknown')}")
                                st.text(f"Size: {item.get('size', 0)} bytes")
                        except Exception as e:
                            logger.warning(f"Failed to load image {name}: {str(e)}")
                            st.write(f"📷 {name} (Failed to load)")
        except Exception as e:
            logger.error(f"Browse images failed: {str(e)}")
            st.error(str(e))