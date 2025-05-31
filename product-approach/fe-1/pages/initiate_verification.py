import streamlit as st
import logging
from .improved_image_selector import render_improved_s3_image_selector

logger = logging.getLogger(__name__)

def get_bucket_name(api_client, bucket_type):
    """Get the actual bucket name from the CONFIG_SECRET in AWS Secrets Manager"""
    try:
        # Get bucket name from the API client's config loader (which loads from CONFIG_SECRET)
        if bucket_type == "reference":
            bucket_name = api_client.config_loader.get('REFERENCE_BUCKET', '')
        elif bucket_type == "checking":
            bucket_name = api_client.config_loader.get('CHECKING_BUCKET', '')
        else:
            logger.warning(f"Unknown bucket type: {bucket_type}")
            return f"{bucket_type}-bucket"  # fallback

        # If we don't have the bucket name from CONFIG_SECRET, this is an error
        if not bucket_name:
            error_msg = f"Bucket name for {bucket_type} not found in CONFIG_SECRET. Please ensure REFERENCE_BUCKET and CHECKING_BUCKET are properly configured in AWS Secrets Manager."
            logger.error(error_msg)
            raise ValueError(error_msg)

        logger.info(f"Retrieved {bucket_type} bucket name from CONFIG_SECRET: {bucket_name}")
        return bucket_name

    except Exception as e:
        logger.error(f"Failed to get bucket name for {bucket_type} from CONFIG_SECRET: {str(e)}")
        # Re-raise the exception so the UI can show the proper error
        raise Exception(f"Configuration error: Could not retrieve {bucket_type} bucket name from CONFIG_SECRET. Please check your AWS Secrets Manager configuration.")

def render_s3_image_selector(api_client, bucket_type, label, session_key):
    """Render an S3 image selector component"""
    st.subheader(label)

    # Initialize session state for this selector
    if f"{session_key}_path" not in st.session_state:
        st.session_state[f"{session_key}_path"] = ""
    if f"{session_key}_selected_image" not in st.session_state:
        st.session_state[f"{session_key}_selected_image"] = None
    if f"{session_key}_selected_url" not in st.session_state:
        st.session_state[f"{session_key}_selected_url"] = ""

    # Path input and browse button
    col_path, col_browse = st.columns([3, 1])
    with col_path:
        current_path = st.text_input(
            f"Path in {bucket_type} bucket",
            value=st.session_state[f"{session_key}_path"],
            key=f"{session_key}_path_input",
            help="Leave blank for root directory"
        )
    with col_browse:
        browse_button = st.button("Browse", key=f"{session_key}_browse")

    # Update path if user typed in the input
    if current_path != st.session_state[f"{session_key}_path"]:
        st.session_state[f"{session_key}_path"] = current_path
        # Clear items to force refresh when path changes
        if f"{session_key}_items" in st.session_state:
            del st.session_state[f"{session_key}_items"]

    # Load items if browse button clicked or no items cached
    should_load = browse_button or f"{session_key}_items" not in st.session_state
    if should_load:
        try:
            with st.spinner("Loading images..."):
                browser_response = api_client.browse_images(current_path, bucket_type)
                st.session_state[f"{session_key}_items"] = browser_response.get('items', [])
                st.session_state[f"{session_key}_current_path"] = browser_response.get('currentPath', '')
                st.session_state[f"{session_key}_parent_path"] = browser_response.get('parentPath', '')
        except Exception as e:
            st.error(f"Failed to browse {bucket_type} bucket: {str(e)}")
            return None

    # Display current path and navigation
    if f"{session_key}_current_path" in st.session_state:
        current_display_path = st.session_state[f"{session_key}_current_path"]
        st.write(f"**Current Path:** {current_display_path if current_display_path else '(root)'}")

        # Go up button
        if st.session_state.get(f"{session_key}_parent_path") is not None:
            if st.button("⬆️ Go Up", key=f"{session_key}_go_up"):
                new_path = st.session_state[f"{session_key}_parent_path"]
                st.session_state[f"{session_key}_path"] = new_path
                # Clear items to force refresh
                if f"{session_key}_items" in st.session_state:
                    del st.session_state[f"{session_key}_items"]
                st.rerun()

    # Display items
    if f"{session_key}_items" in st.session_state:
        items = st.session_state[f"{session_key}_items"]
        if not items:
            st.info("No items found in this location.")
            return st.session_state[f"{session_key}_selected_url"]

        # Create grid layout for items
        cols = st.columns(3)
        for idx, item in enumerate(items):
            col = cols[idx % 3]
            with col:
                name = item.get('name', 'Unknown')
                item_type = item.get('type', 'unknown')
                item_path = item.get('path', '')

                if item_type == 'folder':
                    if st.button(f"📁 {name}", key=f"{session_key}_folder_{idx}"):
                        st.session_state[f"{session_key}_path"] = item_path
                        # Clear items to force refresh when navigating to new folder
                        if f"{session_key}_items" in st.session_state:
                            del st.session_state[f"{session_key}_items"]
                        st.rerun()
                elif item_type == 'image':
                    # Try to show image thumbnail
                    try:
                        logger.info(f"Attempting to get image URL for path: '{item_path}' in bucket: {bucket_type}")
                        url_response = api_client.get_image_url(item_path, bucket_type)
                        logger.info(f"URL response: {url_response}")

                        image_url = url_response.get('presignedUrl', '')
                        if image_url:
                            st.image(image_url, caption=name, width=150)
                        else:
                            logger.warning(f"No presignedUrl in response for {name}: {url_response}")
                            st.write(f"📷 {name}")
                            st.caption("Preview unavailable")
                    except Exception as e:
                        logger.warning(f"Failed to load image {name} (path: {item_path}): {str(e)}")
                        st.write(f"📷 {name}")
                        st.caption("Preview unavailable")
                        # Show debug info in expander
                        with st.expander(f"Debug info for {name}"):
                            st.write(f"**Path:** `{item_path}`")
                            st.write(f"**Type:** {item_type}")
                            st.write(f"**Error:** {str(e)}")

                    # Select button (always show, regardless of image load status)
                    if st.button(f"Select {name}", key=f"{session_key}_select_{idx}"):
                        try:
                            st.session_state[f"{session_key}_selected_image"] = name
                            # Generate proper S3 URL using the bucket configuration
                            bucket_name = get_bucket_name(api_client, bucket_type)
                            st.session_state[f"{session_key}_selected_url"] = f"s3://{bucket_name}/{item_path}"
                            # Don't rerun immediately, let user see the selection
                        except Exception as config_error:
                            st.error(f"Configuration error: {str(config_error)}")
                            return None

    # Display selected image
    if st.session_state[f"{session_key}_selected_image"]:
        st.success(f"**Selected:** {st.session_state[f'{session_key}_selected_image']}")
        st.write(f"**URL:** {st.session_state[f'{session_key}_selected_url']}")

    return st.session_state[f"{session_key}_selected_url"]

def app(api_client):
    st.title("Initiate Verification")

    # Debug section (collapsible)
    with st.expander("🔧 Debug Tools"):
        st.write("**Test Image URL Generation**")
        col1, col2 = st.columns(2)
        with col1:
            test_path = st.text_input("Test image path:", placeholder="e.g., AACZ_3.png")
        with col2:
            test_bucket = st.selectbox("Bucket type:", ["reference", "checking"])

        if st.button("Test Image URL") and test_path:
            try:
                url_response = api_client.get_image_url(test_path, test_bucket)
                st.success("✅ API call successful!")
                st.json(url_response)
            except Exception as e:
                st.error(f"❌ API call failed: {str(e)}")

    # Step 1: Select verification type
    verification_type = st.selectbox(
        "Verification Type",
        ["LAYOUT_VS_CHECKING", "PREVIOUS_VS_CURRENT"],
        format_func=lambda x: "Layout vs Checking" if x == "LAYOUT_VS_CHECKING" else "Previous vs Current"
    )

    # Step 2: Select images (outside of form to allow interactive browsing)
    col1, col2 = st.columns(2)

    with col1:
        # Reference Image S3 Selector
        reference_image_url = render_improved_s3_image_selector(
            api_client, "reference", "Reference Image", "ref_img"
        )

    with col2:
        # Checking Image S3 Selector
        checking_image_url = render_improved_s3_image_selector(
            api_client, "checking", "Checking Image", "check_img"
        )

    # Step 3: Additional metadata in form
    with st.form(key='init_verification_form'):
        # Only show Vending Machine ID for PREVIOUS_VS_CURRENT
        vending_machine_id = None
        if verification_type == "PREVIOUS_VS_CURRENT":
            vending_machine_id = st.text_input("Vending Machine ID", help="Required - ID of the vending machine")

            # Type-specific fields for PREVIOUS_VS_CURRENT only
            previous_verification_id = st.text_input("Previous Verification ID", help="Optional - ID of previous verification")
        else:
            previous_verification_id = None

        submit_button = st.form_submit_button(label='Start Verification')

    if submit_button:
        # Validation based on verification type
        if verification_type == "LAYOUT_VS_CHECKING":
            # For Layout vs Checking, only images are required
            required_fields = [reference_image_url, checking_image_url]
            if not all(required_fields):
                st.error("Both Reference Image and Checking Image must be selected.")
                return
        else:  # PREVIOUS_VS_CURRENT
            # For Previous vs Current, vending machine ID and images are required
            required_fields = [vending_machine_id, reference_image_url, checking_image_url]
            if not all(required_fields):
                st.error("Vending Machine ID, Reference Image, and Checking Image are all required.")
                return

        try:
            # Different parameters based on verification type
            if verification_type == "LAYOUT_VS_CHECKING":
                # For Layout vs Checking, we don't need layout_id, layout_prefix, or vending_machine_id
                response = api_client.initiate_verification(
                    verification_type=verification_type,
                    reference_image_url=reference_image_url,
                    checking_image_url=checking_image_url,
                    vending_machine_id="default",  # Use a default value since API might require it
                    layout_id=1,  # Use default values since these are removed from UI
                    layout_prefix="default"
                )
            else:  # PREVIOUS_VS_CURRENT
                response = api_client.initiate_verification(
                    verification_type=verification_type,
                    reference_image_url=reference_image_url,
                    checking_image_url=checking_image_url,
                    vending_machine_id=vending_machine_id,
                    previous_verification_id=previous_verification_id if previous_verification_id else None
                )

            st.success("Verification initiated successfully!")
            verification_id = response.get('verificationId', 'N/A')
            st.write(f"Verification ID: {verification_id}")
            st.write(f"Status: {response.get('status', 'N/A')}")
            st.write(f"Initiated at: {response.get('verificationAt', 'N/A')}")

            # Store the verification ID in session state
            if verification_id != 'N/A':
                st.session_state['selected_verification'] = verification_id
                st.info("✅ Verification initiated successfully! You can view the status in the Verifications page.")

        except Exception as e:
            logger.error(f"Initiate verification failed: {str(e)}")
            st.error(str(e))