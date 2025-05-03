import streamlit as st
import logging

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Image Browser")
    path = st.text_input("Enter path (leave blank for root)", "")
    if st.button("Browse"):
        try:
            images = api_client.browse_images(path)
            if not images:
                st.write("No images found.")
                return
            for img in images:
                key = img.get('key', '')
                try:
                    url_response = api_client.get_image_url(key)
                    image_url = url_response.get('url', '')
                    st.image(image_url, caption=key, width=200)
                except Exception as e:
                    logger.warning(f"Failed to load image {key}: {str(e)}")
                    st.write(f"Image: {key} (Failed to load)")
        except Exception as e:
            logger.error(f"Browse images failed: {str(e)}")
            st.error(str(e))