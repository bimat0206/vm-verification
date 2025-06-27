import streamlit as st
import logging

logger = logging.getLogger(__name__)

def render_improved_s3_image_selector(api_client, bucket_type, label, session_key):
    """Render a clean, streamlined S3 image selector focused on the core workflow"""
    st.subheader(label)

    # Initialize session state
    init_keys = [
        f"{session_key}_path", f"{session_key}_selected_image", f"{session_key}_selected_url",
        f"{session_key}_page", f"{session_key}_search", f"{session_key}_show_images",
        f"{session_key}_items_per_page", f"{session_key}_auto_navigate", f"{session_key}_show_advanced",
        f"{session_key}_previewed_image"
    ]

    defaults = ["", None, "", 0, "", False, 12, True, False, None]

    for key, default in zip(init_keys, defaults):
        if key not in st.session_state:
            st.session_state[key] = default

    # Simplified main controls - focus on core workflow
    col_path, col_browse = st.columns([3, 1])
    with col_path:
        current_path = st.text_input(
            f"📁 Browse {bucket_type} bucket",
            value=st.session_state[f"{session_key}_path"],
            key=f"{session_key}_path_input",
            placeholder="Enter folder path or leave blank for root",
            label_visibility="collapsed"
        )
    with col_browse:
        browse_button = st.button("🔍 Browse", key=f"{session_key}_browse", use_container_width=True)

    # Auto-navigate is enabled by default but hidden to reduce clutter
    auto_navigate = st.session_state[f"{session_key}_auto_navigate"]

    # Update path if changed
    if current_path != st.session_state[f"{session_key}_path"]:
        st.session_state[f"{session_key}_path"] = current_path
        st.session_state[f"{session_key}_page"] = 0
        # Clear cached items
        if f"{session_key}_items" in st.session_state:
            del st.session_state[f"{session_key}_items"]
        # Auto-load if auto-navigate is enabled
        if auto_navigate:
            browse_button = True

    # Load items when browse button is clicked, auto-navigate is triggered, or items are not cached
    should_load = browse_button or (auto_navigate and f"{session_key}_items" not in st.session_state)

    if should_load:
        try:
            with st.spinner("Loading..."):
                browser_response = api_client.browse_images(current_path, bucket_type)
                st.session_state[f"{session_key}_items"] = browser_response.get('items', [])
                st.session_state[f"{session_key}_current_path"] = browser_response.get('currentPath', '')
                st.session_state[f"{session_key}_parent_path"] = browser_response.get('parentPath', '')
        except Exception as e:
            st.error(f"❌ Failed to browse {bucket_type} bucket: {str(e)}")
            return st.session_state[f"{session_key}_selected_url"]

    # Navigation and controls
    if f"{session_key}_items" in st.session_state:
        current_display_path = st.session_state.get(f"{session_key}_current_path", "")

        # Compact navigation bar
        nav_col1, nav_col2, nav_col3 = st.columns([1, 2, 1])

        with nav_col1:
            # Go up button
            parent_path = st.session_state.get(f"{session_key}_parent_path")
            if parent_path is not None:
                if st.button("⬆️ Up", key=f"{session_key}_go_up", use_container_width=True):
                    st.session_state[f"{session_key}_path"] = parent_path
                    st.session_state[f"{session_key}_page"] = 0
                    # Clear preview when navigating
                    st.session_state[f"{session_key}_previewed_image"] = None
                    if f"{session_key}_items" in st.session_state:
                        del st.session_state[f"{session_key}_items"]
                    if auto_navigate:
                        st.rerun()

        with nav_col2:
            # Compact breadcrumb display
            if current_display_path:
                path_parts = [part for part in current_display_path.split('/') if part]
                breadcrumb_text = "🏠"
                if path_parts:
                    breadcrumb_text += " / " + " / ".join(path_parts[-2:])  # Show last 2 levels only
                st.write(f"**📍 {breadcrumb_text}**")
            else:
                st.write("**📍 🏠 Root**")

        with nav_col3:
            if st.button("🏠 Root", key=f"{session_key}_go_root", use_container_width=True):
                st.session_state[f"{session_key}_path"] = ""
                st.session_state[f"{session_key}_page"] = 0
                # Clear preview when navigating
                st.session_state[f"{session_key}_previewed_image"] = None
                if f"{session_key}_items" in st.session_state:
                    del st.session_state[f"{session_key}_items"]
                if auto_navigate:
                    st.rerun()

        # Advanced navigation (collapsible)
        with st.expander("🔧 Advanced Navigation", expanded=False):
            # Full breadcrumb navigation
            if current_display_path:
                st.write("**Click to navigate:**")
                breadcrumb_cols = st.columns([1, 1, 1, 1, 1, 1])

                # Root button
                with breadcrumb_cols[0]:
                    if st.button("🏠 Root", key=f"{session_key}_nav_root", use_container_width=True):
                        st.session_state[f"{session_key}_path"] = ""
                        st.session_state[f"{session_key}_page"] = 0
                        # Clear preview when navigating
                        st.session_state[f"{session_key}_previewed_image"] = None
                        if f"{session_key}_items" in st.session_state:
                            del st.session_state[f"{session_key}_items"]
                        if auto_navigate:
                            st.rerun()

                # Path parts as clickable buttons
                path_parts = [part for part in current_display_path.split('/') if part]
                for i, part in enumerate(path_parts[:5]):
                    col_idx = i + 1
                    if col_idx < len(breadcrumb_cols):
                        with breadcrumb_cols[col_idx]:
                            nav_path = '/'.join(path_parts[:i+1])
                            if st.button(f"📁 {part}", key=f"{session_key}_nav_{i}", use_container_width=True):
                                st.session_state[f"{session_key}_path"] = nav_path
                                st.session_state[f"{session_key}_page"] = 0
                                # Clear preview when navigating
                                st.session_state[f"{session_key}_previewed_image"] = None
                                if f"{session_key}_items" in st.session_state:
                                    del st.session_state[f"{session_key}_items"]
                                if auto_navigate:
                                    st.rerun()

            # Settings
            col_auto, col_refresh = st.columns(2)
            with col_auto:
                auto_navigate_new = st.checkbox(
                    "Auto-navigate folders",
                    value=auto_navigate,
                    key=f"{session_key}_auto_nav_advanced",
                    help="Automatically load folder contents when clicked"
                )
                if auto_navigate_new != auto_navigate:
                    st.session_state[f"{session_key}_auto_navigate"] = auto_navigate_new
                    st.rerun()

            with col_refresh:
                if st.button("🔄 Refresh Current Folder", key=f"{session_key}_refresh_advanced", use_container_width=True):
                    if f"{session_key}_items" in st.session_state:
                        del st.session_state[f"{session_key}_items"]
                    if auto_navigate:
                        st.rerun()

        # Content display
        items = st.session_state[f"{session_key}_items"]
        if items:
            # Simple search bar (most commonly used)
            search_term = st.text_input(
                "🔍 Search",
                value=st.session_state[f"{session_key}_search"],
                key=f"{session_key}_search_input",
                placeholder="Type to filter items...",
                label_visibility="collapsed"
            )
            if search_term != st.session_state[f"{session_key}_search"]:
                st.session_state[f"{session_key}_search"] = search_term
                st.session_state[f"{session_key}_page"] = 0
                # Clear preview when search changes
                st.session_state[f"{session_key}_previewed_image"] = None

            # Filter items based on search
            filtered_items = items
            if search_term:
                filtered_items = [
                    item for item in items
                    if search_term.lower() in item.get('name', '').lower()
                ]

            # Separate folders and images
            folders = [item for item in filtered_items if item.get('type') == 'folder']
            images = [item for item in filtered_items if item.get('type') == 'image']
            files = [item for item in filtered_items if item.get('type') not in ['folder', 'image']]

            # Compact summary
            if search_term:
                st.caption(f"Found: {len(folders)} folders, {len(images)} images")

            # Show all items by default (no pagination for simplicity)
            all_items = folders + images + files
            show_images = st.session_state[f"{session_key}_show_images"]

            # View options in collapsible section
            with st.expander("⚙️ View Options", expanded=False):
                # Show preview info
                st.info("💡 **New!** Click '👁️ Preview' on any image to see it. Only one preview loads at a time for better performance.")

                col_preview, col_pagination = st.columns(2)
                with col_preview:
                    show_images_new = st.checkbox(
                        "Show ALL image previews (legacy mode)",
                        value=show_images,
                        key=f"{session_key}_show_images_advanced",
                        help="Enable this to show all image previews at once (may be slower)"
                    )
                    if show_images_new != show_images:
                        st.session_state[f"{session_key}_show_images"] = show_images_new
                        # Clear individual preview when switching to global mode
                        if show_images_new:
                            st.session_state[f"{session_key}_previewed_image"] = None
                        st.rerun()

                with col_pagination:
                    items_per_page = st.selectbox(
                        "Items per page",
                        [12, 24, 48, 100],
                        index=[12, 24, 48, 100].index(st.session_state.get(f"{session_key}_items_per_page", 12)),
                        key=f"{session_key}_items_per_page_advanced"
                    )
                    if items_per_page != st.session_state[f"{session_key}_items_per_page"]:
                        st.session_state[f"{session_key}_items_per_page"] = items_per_page
                        st.session_state[f"{session_key}_page"] = 0

            # Pagination (only show if many items)
            total_items = len(all_items)
            current_page = st.session_state[f"{session_key}_page"]
            items_per_page = st.session_state[f"{session_key}_items_per_page"]

            if total_items > items_per_page:
                total_pages = (total_items - 1) // items_per_page + 1
                start_idx = current_page * items_per_page
                end_idx = min(start_idx + items_per_page, total_items)
                page_items = all_items[start_idx:end_idx]

                # Compact pagination
                page_col1, page_col2, page_col3 = st.columns([1, 2, 1])
                with page_col1:
                    if current_page > 0:
                        if st.button("⬅️ Prev", key=f"{session_key}_prev_page", use_container_width=True):
                            st.session_state[f"{session_key}_page"] = current_page - 1
                            # Clear preview when changing pages
                            st.session_state[f"{session_key}_previewed_image"] = None
                            st.rerun()
                with page_col2:
                    st.write(f"**Page {current_page + 1} of {total_pages}**")
                with page_col3:
                    if current_page < total_pages - 1:
                        if st.button("Next ➡️", key=f"{session_key}_next_page", use_container_width=True):
                            st.session_state[f"{session_key}_page"] = current_page + 1
                            # Clear preview when changing pages
                            st.session_state[f"{session_key}_previewed_image"] = None
                            st.rerun()
            else:
                page_items = all_items

            # Display items
            if page_items:
                render_items_grid(api_client, bucket_type, session_key, page_items, show_images)
            else:
                st.info("No items found matching your search.")
        else:
            st.info("📂 No items found. Try browsing a different folder.")

    # Display selected image (compact)
    if st.session_state[f"{session_key}_selected_image"]:
        st.success(f"✅ **Selected:** {st.session_state[f'{session_key}_selected_image']}")
        # Show URL in collapsible section to reduce clutter
        with st.expander("📋 View S3 URL", expanded=False):
            st.code(st.session_state[f'{session_key}_selected_url'], language='text')

    return st.session_state[f"{session_key}_selected_url"]


def render_items_grid(api_client, bucket_type, session_key, items, show_images):
    """Render items in a clean, focused grid layout with on-demand preview"""
    if not items:
        return

    # Get currently previewed image
    previewed_image = st.session_state.get(f"{session_key}_previewed_image")

    # Create responsive grid
    cols = st.columns(3)

    for idx, item in enumerate(items):
        col = cols[idx % 3]
        with col:
            name = item.get('name', 'Unknown')
            item_type = item.get('type', 'unknown')
            item_path = item.get('path', '')

            # Clean container for each item
            with st.container():
                if item_type == 'folder':
                    # Folder display - clean and simple
                    st.write(f"**📁 {name}**")
                    if st.button(f"Open", key=f"{session_key}_folder_{idx}", use_container_width=True):
                        st.session_state[f"{session_key}_path"] = item_path
                        st.session_state[f"{session_key}_page"] = 0
                        # Clear preview when navigating
                        st.session_state[f"{session_key}_previewed_image"] = None
                        if f"{session_key}_items" in st.session_state:
                            del st.session_state[f"{session_key}_items"]
                        if st.session_state.get(f"{session_key}_auto_navigate", True):
                            st.rerun()

                elif item_type == 'image':
                    # Check if this image is currently previewed
                    is_previewed = previewed_image == item_path
                    is_selected = st.session_state.get(f"{session_key}_selected_image") == name

                    # Image file name display with visual feedback
                    if is_selected:
                        # Selected image - green background
                        st.markdown(f"""
                        <div style="background-color: #d4edda; padding: 8px; border-radius: 4px; border-left: 4px solid #28a745;">
                            <strong>🖼️ {name}</strong> ✅
                        </div>
                        """, unsafe_allow_html=True)
                    elif is_previewed:
                        # Previewed image - blue background
                        st.markdown(f"""
                        <div style="background-color: #d1ecf1; padding: 8px; border-radius: 4px; border-left: 4px solid #17a2b8;">
                            <strong>🖼️ {name}</strong> 👁️
                        </div>
                        """, unsafe_allow_html=True)
                    else:
                        # Regular image - clickable
                        st.write(f"**🖼️ {name}**")

                    # Clickable file name for preview (only if not already previewed)
                    if not is_previewed:
                        if st.button(f"👁️ Preview", key=f"{session_key}_preview_{idx}", use_container_width=True):
                            # Set this image as the previewed one (clear any previous preview)
                            st.session_state[f"{session_key}_previewed_image"] = item_path
                            st.rerun()

                    # Show image preview if this specific image is previewed OR if global show_images is enabled
                    should_show_preview = is_previewed or show_images
                    if should_show_preview:
                        try:
                            url_response = api_client.get_image_url(item_path, bucket_type)
                            image_url = (
                                url_response.get('presignedUrl') or
                                url_response.get('url') or
                                url_response.get('imageUrl') or
                                url_response.get('signedUrl')
                            )

                            if image_url:
                                st.image(image_url, use_column_width=True)
                                # Add a "Hide Preview" button for the previewed image
                                if is_previewed:
                                    if st.button(f"🙈 Hide Preview", key=f"{session_key}_hide_{idx}", use_container_width=True):
                                        st.session_state[f"{session_key}_previewed_image"] = None
                                        st.rerun()
                            else:
                                st.info("Preview unavailable")

                        except Exception as e:
                            # Compact error display
                            with st.expander("⚠️ Preview error", expanded=False):
                                st.error(f"Failed to load: {str(e)}")
                                st.caption(f"Path: {item_path}")

                    # Compact file info
                    size = item.get('size', 0)
                    if size > 0:
                        size_mb = size / (1024 * 1024)
                        st.caption(f"📏 {size_mb:.1f} MB")

                    # Primary action - Select button
                    if st.button(f"✅ Select", key=f"{session_key}_select_{idx}", use_container_width=True, type="primary"):
                        try:
                            # Get bucket name from config
                            if bucket_type == "reference":
                                bucket_name = api_client.config_loader.get('REFERENCE_BUCKET', '')
                            elif bucket_type == "checking":
                                bucket_name = api_client.config_loader.get('CHECKING_BUCKET', '')
                            else:
                                bucket_name = f"{bucket_type}-bucket"

                            st.session_state[f"{session_key}_selected_image"] = name
                            st.session_state[f"{session_key}_selected_url"] = f"s3://{bucket_name}/{item_path}"
                            st.success(f"✅ Selected!")
                        except Exception as e:
                            st.error(f"❌ Error: {str(e)}")

                else:
                    # Other file types - minimal display
                    st.write(f"**📄 {name}**")
                    st.caption(f"Type: {item_type}")

                # Subtle separator
                st.write("")
