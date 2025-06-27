import streamlit as st
import logging
from datetime import datetime, timedelta

logger = logging.getLogger(__name__)



def parse_s3_uri(s3_uri):
    """Helper function to parse S3 URI into bucket name and key."""
    if not s3_uri or not s3_uri.startswith("s3://"):
        return None, None
    try:
        parts = s3_uri.replace("s3://", "").split("/", 1)
        bucket_name = parts[0]
        key = parts[1] if len(parts) > 1 else ""
        return bucket_name, key
    except Exception:
        return None, None

def determine_bucket_type(bucket_name):
    """Determine bucket type from bucket name."""
    if not bucket_name:
        return None

    bucket_lower = bucket_name.lower()
    if "reference" in bucket_lower:
        return "reference"
    elif "checking" in bucket_lower:
        return "checking"
    else:
        # Default fallback - could be enhanced based on naming conventions
        return None

def get_detailed_verification_data(api_client, verification_id):
    """Fetch detailed verification data by ID."""
    if not api_client or not verification_id:
        return None

    try:
        detailed_data = api_client.get_verification_details(verification_id)
        logger.info(f"Successfully fetched detailed data for verification {verification_id}")
        return detailed_data
    except Exception as e:
        logger.warning(f"Could not fetch detailed verification data for {verification_id}: {str(e)}")
        return None

def extract_llm_analysis(basic_data, detailed_data):
    """Extract LLM analysis from verification data."""
    # Try to find analysis in various possible fields
    analysis_fields = [
        'llmAnalysis', 'analysis', 'llm_analysis',
        'verificationAnalysis', 'aiAnalysis', 'description'
    ]

    # Check detailed data first, then basic data
    for data_source in [detailed_data, basic_data]:
        if not data_source:
            continue

        for field in analysis_fields:
            if field in data_source and data_source[field]:
                return data_source[field]

        # Check nested in result or summary
        for nested_field in ['result', 'verificationSummary']:
            if nested_field in data_source and isinstance(data_source[nested_field], dict):
                for field in analysis_fields:
                    if field in data_source[nested_field] and data_source[nested_field][field]:
                        return data_source[nested_field][field]

    return None

def get_llm_analysis_from_conversation(api_client, verification_id):
    """Fetch LLM analysis from conversation API turn2Content."""
    if not api_client or not verification_id:
        return None, None

    try:
        conversation_data = api_client.get_verification_conversation(verification_id)
        logger.info(f"Successfully fetched conversation data for verification {verification_id}")

        # Extract turn2Content which contains the LLM analysis
        turn2_content = conversation_data.get('turn2Content') if conversation_data else None
        if turn2_content and turn2_content.get('content'):
            return turn2_content.get('content'), turn2_content

        return None, None
    except Exception as e:
        logger.warning(f"Could not fetch conversation data for {verification_id}: {str(e)}")
        return None, None

def cleanup_old_llm_cache():
    """Clean up very old LLM cache entries to prevent memory buildup."""
    from datetime import datetime, timedelta

    # Only clean up cache entries older than 30 minutes to preserve performance
    cutoff_time = datetime.now() - timedelta(minutes=30)

    cache_keys_to_remove = []
    for key in st.session_state.keys():
        if key.startswith('llm_data_'):
            cached_data = st.session_state.get(key, {})
            cached_at_str = cached_data.get('cached_at')

            if cached_at_str:
                try:
                    cached_at = datetime.fromisoformat(cached_at_str)
                    if cached_at < cutoff_time:
                        cache_keys_to_remove.append(key)
                except (ValueError, TypeError):
                    # If we can't parse the date, remove the entry
                    cache_keys_to_remove.append(key)
            else:
                # If no timestamp, remove the entry
                cache_keys_to_remove.append(key)

    # Remove old cache entries
    for key in cache_keys_to_remove:
        del st.session_state[key]

    if cache_keys_to_remove:
        logger.info(f"Cleaned up {len(cache_keys_to_remove)} old LLM cache entries (older than 30 minutes)")



def show_additional_verification_details(detailed_data, basic_data):
    """Show additional details from detailed verification data."""
    if not detailed_data:
        return

    # Fields that might be in detailed data but not in basic list data
    additional_fields = {
        'processingTime': '⏱️ Processing Time',
        'confidence': '🎯 Confidence Score',
        'modelVersion': '🤖 Model Version',
        'processingStatus': '⚙️ Processing Status',
        'errorMessage': '❌ Error Message',
        'metadata': '📋 Metadata'
    }

    details_shown = False
    for field, label in additional_fields.items():
        if field in detailed_data and detailed_data[field] and field not in basic_data:
            if not details_shown:
                details_shown = True

            value = detailed_data[field]
            if isinstance(value, dict):
                st.write(f"**{label}:**")
                st.json(value)
            else:
                st.write(f"**{label}:** {value}")

    if not details_shown:
        st.write("No additional details available.")

def handle_verification_lookup(api_client, verification_id):
    """Handle lookup of a specific verification by ID."""
    try:
        with st.spinner(f"Looking up verification {verification_id}..."):
            verification_data = api_client.get_verification_details(verification_id)

        if verification_data:
            st.success(f"✅ Found verification: {verification_id}")

            # Display the verification in a card format
            status = verification_data.get('verificationStatus', 'UNKNOWN')
            render_verification_card(
                verification_data,
                0,  # idx
                status,
                verification_id,
                f"lookup_{verification_id}",
                True,  # Always expanded for lookup
                api_client
            )
        else:
            st.error(f"❌ No verification found with ID: {verification_id}")

    except Exception as e:
        logger.error(f"Error looking up verification {verification_id}: {str(e)}")
        st.error(f"❌ Error looking up verification: {str(e)}")

        # Show helpful suggestions
        st.info("""
        **Troubleshooting tips:**
        - Verify the verification ID is correct and complete
        - Check that the verification exists in the system
        - Ensure you have proper access permissions
        """)

def app(api_client):
    # Apply custom CSS first
    apply_custom_css()

    # Clean up old LLM cache entries to prevent memory buildup
    cleanup_old_llm_cache()

    # Page Title - Simplified to match typical Streamlit page structure
    st.title("📊 Verification Results")
    st.markdown("View and analyze verification results from completed verifications.")

    # Quick verification lookup by ID
    with st.expander("🔍 Quick Verification Lookup", expanded=False):
        st.markdown("Enter a specific verification ID to view detailed results:")

        lookup_col1, lookup_col2 = st.columns([3, 1])
        with lookup_col1:
            lookup_verification_id = st.text_input(
                "Verification ID",
                placeholder="Enter verification ID (e.g., a041e458-3171-43e9-a149-f63c5916d3a2)",
                help="Enter the exact verification ID to fetch detailed results",
                key="lookup_verification_id"
            )
        with lookup_col2:
            st.markdown("<br>", unsafe_allow_html=True)  # Add spacing
            if st.button("🔍 Lookup", use_container_width=True, type="primary"):
                if lookup_verification_id.strip():
                    handle_verification_lookup(api_client, lookup_verification_id.strip())
                else:
                    st.warning("Please enter a verification ID")

    st.markdown("---")


    # Debug info for troubleshooting
    show_debug = st.checkbox("🔧 Show Debug Info", help="Show debugging information")
    st.session_state['show_debug_info'] = show_debug

    if show_debug:
        st.info(f"✅ Page loaded successfully at {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        if api_client and hasattr(api_client, 'base_url'):
            st.info(f"🔗 API Client Base URL: {api_client.base_url}")
            # Test API connection
            try:
                test_response = api_client.list_verifications({'limit': 1})
                st.success(f"✅ API Connection Working - Found {test_response.get('pagination', {}).get('total', 0)} total verifications")
            except Exception as e:
                st.error(f"❌ API Connection Failed: {str(e)}")
                logger.error(f"API test failed: {str(e)}")
        else:
            st.warning("API client not available for debug info.")


    # Initialize session state for filter management
    if 'verification_filters_applied' not in st.session_state:
        st.session_state['verification_filters_applied'] = False
    if 'verification_last_params' not in st.session_state:
        st.session_state['verification_last_params'] = {}
    if 'verification_loading' not in st.session_state:
        st.session_state['verification_loading'] = False
    if 'verification_last_applied_filters' not in st.session_state:
        st.session_state['verification_last_applied_filters'] = {}
    if 'verification_expanded_results' not in st.session_state:
        st.session_state['verification_expanded_results'] = set()

    # Define sort_options and defaults
    sort_options = {
        "newest": "verificationAt:desc",
        "oldest": "verificationAt:asc",
        "accuracy_high": "overallAccuracy:desc",
        "accuracy_low": "overallAccuracy:asc"
    }

    # Initialize filter values from session state or defaults
    last_filters = st.session_state.get('verification_last_applied_filters', {})
    status_filter_default = last_filters.get('status', "all")
    vending_machine_id_default = last_filters.get('machine_id', "")
    search_verification_id_default = last_filters.get('search_verification_id', "") 
    date_option_default = last_filters.get('date_option', "All time")
    
    from_date_str = last_filters.get('from_date')
    to_date_str = last_filters.get('to_date')

    from_date_default = datetime.fromisoformat(from_date_str) if from_date_str else (datetime.now() - timedelta(days=7))
    to_date_default = datetime.fromisoformat(to_date_str) if to_date_str else datetime.now()
    
    limit_default = st.session_state.get('verification_last_limit', 5)
    sort_by_default = st.session_state.get('verification_last_sort_key', "newest")


    apply_filters_btn = False 

    # Collapsible search and filter section
    with st.expander("🔍 Search & Filter Options", expanded=True):

        st.markdown("##### Quick Search")
        q_col1, q_col2, q_col3 = st.columns(3) 

        with q_col1:
            status_filter = st.selectbox(
                "Status",
                ["all", "CORRECT", "INCORRECT"],
                index=["all", "CORRECT", "INCORRECT"].index(status_filter_default),
                format_func=lambda x: "All Statuses" if x == "all" else x,
                help="Filter by verification status"
            )
        with q_col2:
            vending_machine_id = st.text_input(
                "Machine ID",
                value=vending_machine_id_default,
                placeholder="Filter by Machine ID",
                help="Enter specific machine ID to filter results"
            )
        with q_col3:
            search_verification_id = st.text_input(
                "Verification ID",
                value=search_verification_id_default,
                placeholder="Filter by Verification ID",
                help="Enter specific Verification ID to search"
            )
        
        st.markdown("##### Advanced Date & Display Options")
        col_date1, col_date2 = st.columns(2)
        with col_date1:
            date_option = st.selectbox(
                "Time Period",
                ["All time", "Last 24 hours", "Last 7 days", "Last 30 days", "Custom"],
                index=["All time", "Last 24 hours", "Last 7 days", "Last 30 days", "Custom"].index(date_option_default),
                help="Select time period for results"
            )

        with col_date2:
            if date_option == "Custom":
                custom_from_date = st.date_input("From Date", from_date_default)
                custom_to_date = st.date_input("To Date", to_date_default)
            else:
                st.write("") 

        col_display1, col_display2 = st.columns(2)
        with col_display1:
            limit = st.number_input(
                "Results per page",
                min_value=1,
                max_value=100,
                value=limit_default,
                step=1,
                help="Number of verification results to show per page (minimum 1)"
            )
        with col_display2:
            sort_by = st.selectbox(
                "Sort by",
                list(sort_options.keys()),
                index=list(sort_options.keys()).index(sort_by_default),
                format_func=lambda x: {
                    "newest": "Newest first",
                    "oldest": "Oldest first",
                    "accuracy_high": "Highest accuracy first",
                    "accuracy_low": "Lowest accuracy first"
                }[x],
                help="Choose how to sort the verification results"
            )
        
        st.markdown("<br>", unsafe_allow_html=True)
        action_col1, action_col2 = st.columns(2)
        with action_col1:
            apply_filters_btn = st.button(
                "▶️ Apply Filters & Search",
                type="primary",
                use_container_width=True,
                help="Apply filters and search for results"
            )
        with action_col2:
            if st.button("🔄 Reset All Filters", help="Clear all filters and reset to defaults", use_container_width=True):
                keys_to_clear = [
                    'verification_filters_applied', 
                    'verification_last_params', 
                    'verification_last_applied_filters',
                    'verification_results',
                    'verification_pagination',
                    'verification_page',
                    'verification_last_limit',
                    'verification_last_sort_key'
                ]
                for key in keys_to_clear:
                    if key in st.session_state:
                        del st.session_state[key]
                st.rerun()

        st.markdown("---")

    params = {}
    page = st.session_state.get('verification_page', 0)
    offset = page * limit

    is_pagination_change = False
    if st.session_state.get('verification_filters_applied', False):
        if page != st.session_state.get('verification_last_page', 0) or \
           limit != st.session_state.get('verification_last_limit', 20) or \
           sort_by != st.session_state.get('verification_last_sort_key', "newest"):
            is_pagination_change = True
            logger.info("Pagination change detected.")

    should_fetch_data = apply_filters_btn or is_pagination_change

    if apply_filters_btn:
        st.session_state['verification_page'] = 0
        page = 0
        offset = 0
        logger.info("Apply filters button clicked. Resetting page to 0.")

    if should_fetch_data:
        if status_filter != "all":
            params["verificationStatus"] = status_filter
        if vending_machine_id:
            params["vendingMachineId"] = vending_machine_id.strip()
        if search_verification_id: 
            params["verificationId"] = search_verification_id.strip()


        final_from_date = None
        final_to_date = None
        if date_option == "Custom":
            final_from_date = custom_from_date
            final_to_date = custom_to_date
        elif date_option == "Last 24 hours":
            final_from_date = datetime.now() - timedelta(days=1)
            final_to_date = datetime.now()
        elif date_option == "Last 7 days":
            final_from_date = datetime.now() - timedelta(days=7)
            final_to_date = datetime.now()
        elif date_option == "Last 30 days":
            final_from_date = datetime.now() - timedelta(days=30)
            final_to_date = datetime.now()
        
        if final_from_date and final_to_date:
            params["fromDate"] = datetime.combine(final_from_date, datetime.min.time()).isoformat()
            params["toDate"] = datetime.combine(final_to_date, datetime.max.time()).isoformat()

        params["limit"] = limit
        params["offset"] = offset
        params["sortBy"] = sort_options[sort_by]

        st.session_state['verification_last_params'] = params.copy()
        st.session_state['verification_filters_applied'] = True
        
        st.session_state['verification_last_page'] = page
        st.session_state['verification_last_limit'] = limit
        st.session_state['verification_last_sort_key'] = sort_by
        
        current_applied_filters = {
            'status': status_filter,
            'machine_id': vending_machine_id,
            'search_verification_id': search_verification_id, 
            'date_option': date_option,
            'from_date': final_from_date.isoformat() if final_from_date else None,
            'to_date': final_to_date.isoformat() if final_to_date else None,
        }
        st.session_state['verification_last_applied_filters'] = current_applied_filters.copy()
        
        st.session_state['verification_loading'] = True
        try:
            with st.spinner("⏳ Searching verification results..."):
                response = api_client.list_verifications(params)
            st.session_state['verification_results'] = response.get("results", [])
            st.session_state['verification_pagination'] = response.get("pagination", {})
            st.session_state['verification_loading'] = False
        except Exception as e:
            logger.error(f"List verifications failed: {str(e)}", exc_info=True)
            st.error(f"❌ Failed to fetch verification results: {str(e)}")
            if st.session_state.get('show_debug_info', False):
                st.error(f"🔍 Detailed error: {repr(e)}")
                import traceback
                st.code(traceback.format_exc(), language="python")
            st.session_state['verification_loading'] = False
            st.session_state['verification_results'] = []
            st.session_state['verification_pagination'] = {}
            return

    verifications = st.session_state.get('verification_results', [])
    pagination = st.session_state.get('verification_pagination', {})

    if st.session_state.get('verification_loading', False):
        st.info("⏳ Loading results...")
        return

    if not st.session_state.get('verification_filters_applied', False):
        st.info("ℹ️ Use the filters above and click 'Apply Filters & Search' to find verification results.")
        return

    if not verifications and st.session_state.get('verification_filters_applied', False):
        st.warning("📭 No verification results found matching your criteria. Try adjusting your search filters.")
        return
    
    if verifications:
        total_results = pagination.get("total", 0)
        current_limit = st.session_state.get('verification_last_limit', limit)
        
        start_idx = (page * current_limit) + 1
        end_idx = min((page + 1) * current_limit, total_results)
        total_pages = (total_results + current_limit - 1) // current_limit if current_limit > 0 else 0

        st.markdown(f"##### Found {total_results} results. Showing {start_idx}-{end_idx}.")
        
        if total_pages > 1:
            nav_cols = st.columns([1,1,1,5,1,1])
            with nav_cols[0]:
                if page > 0:
                    if st.button("⏮️ First", use_container_width=True, key="pg_first"):
                        st.session_state['verification_page'] = 0
                        st.rerun()
            with nav_cols[1]:
                if page > 0:
                    if st.button("⬅️ Prev", use_container_width=True, key="pg_prev"):
                        st.session_state['verification_page'] = page - 1
                        st.rerun()
            with nav_cols[2]:
                 st.markdown(f"<div style='text-align: center; margin-top: 0.5em;'>Page {page + 1}/{total_pages}</div>", unsafe_allow_html=True)
            with nav_cols[4]:
                if page < total_pages - 1:
                    if st.button("Next ➡️", use_container_width=True, key="pg_next"):
                        st.session_state['verification_page'] = page + 1
                        st.rerun()
            with nav_cols[5]:
                if page < total_pages - 1:
                    if st.button("Last ⏭️", use_container_width=True, key="pg_last"):
                        st.session_state['verification_page'] = total_pages - 1
                        st.rerun()
        st.markdown("---")

        with st.expander("⚙️ Bulk Actions (for current page results)", expanded=False):
            col_expand, col_collapse = st.columns(2)
            with col_expand:
                if st.button("📖 Expand All on Page", use_container_width=True):
                    for idx, v_item in enumerate(verifications):
                        v_id = v_item.get('verificationId', f'no_id_{idx}')
                        st.session_state['verification_expanded_results'].add(f"{v_id}_{idx}")
                    st.rerun()
            with col_collapse:
                if st.button("📕 Collapse All on Page", use_container_width=True):
                    current_page_keys = {f"{v_item.get('verificationId', f'no_id_{idx}')}_{idx}" for idx, v_item in enumerate(verifications)}
                    st.session_state['verification_expanded_results'] = st.session_state['verification_expanded_results'] - current_page_keys
                    st.rerun()
        
        for idx, v in enumerate(verifications):
            verification_id = v.get('verificationId', f'no_id_{idx}')
            status = v.get('verificationStatus', 'UNKNOWN')
            result_key = f"{verification_id}_{idx}"
            is_expanded = result_key in st.session_state['verification_expanded_results']
            render_verification_card(v, idx, status, verification_id, result_key, is_expanded, api_client) # Pass api_client


def apply_custom_css():
    """Apply custom CSS. Simplified for dark theme consistency."""
    st.markdown("""
    <style>
    .main .block-container {
        padding-top: 1rem;
        max-width: 1200px;
    }

    .stExpander div[data-testid="stVerticalBlock"] h5 { 
        color: #FAFAFA; 
        margin-top: 0.5rem;
        margin-bottom: 0.8rem;
    }

    .verification-card {
        background-color: #262730; 
        border: 1px solid #444; 
        border-radius: 8px;
        padding: 1.5rem;
        margin-bottom: 1.5rem;
        transition: all 0.1s ease-in-out;
    }
    .verification-card:hover {
        border-color: #FF4B4B; 
        box-shadow: 0 0 10px rgba(255, 75, 75, 0.3); 
    }
    .card-expanded { 
        border-left: 3px solid #FF4B4B; 
    }
    
    .verification-card .stContainer,
    .verification-card .stColumn,
    .verification-card .stColumns,
    .verification-card .element-container {
        background-color: transparent !important;
    }
    
    .verification-id-display {
        font-size: 1.2em !important; 
        font-weight: 600 !important;
        color: #EAEAEA !important; 
        display: block; 
        margin-bottom: 0.25rem; 
    }

    .verification-card > div:nth-child(1) > div:nth-child(1) > div:nth-child(1) > div[data-testid="stButton"] button {
        font-size: 0.8em !important; 
        padding: 2px 6px !important; 
        line-height: 1.2 !important; 
        min-height: auto !important; 
        height: auto !important; 
        display: inline-flex; 
        align-items: center;
        justify-content: center;
    }


    .status-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        padding: 0.5rem 1rem; 
        border-radius: 16px; 
        font-weight: 600;
        font-size: 0.85rem;
        text-transform: uppercase;
        border: 1px solid transparent; 
    }
    .status-correct {
        background-color: #1C4A36; 
        color: #A6E6A6; 
        border-color: #2A6E49;
    }
    .status-incorrect {
        background-color: #5D1D23; 
        color: #F4AAAA; 
        border-color: #8B2C35;
    }
    .status-unknown {
        background-color: #664D03; 
        color: #FFDDAA; 
        border-color: #997404;
    }

    .metric-display { 
        background-color: #31333F; 
        border: 1px solid #4A4C5A;
        border-radius: 6px;
        padding: 1rem;
        text-align: center;
        margin-bottom: 0.5rem;
    }
    .metric-value {
        font-size: 1.8rem !important; 
        font-weight: 700 !important;
        color: #FAFAFA !important; 
        margin-bottom: 0.25rem !important;
    }
    .metric-label {
        font-size: 0.9rem !important;
        color: #A0A0A0 !important; 
        font-weight: 500 !important;
        text-transform: uppercase;
    }

    .details-section {
        margin-top: 0.5rem; /* Reduced margin */
        padding: 0.8rem; /* Reduced padding */
        background-color: #1E1E1E; 
        border-radius: 6px;
        border: 1px solid #333;
    }
    .details-section h6 { /* Section headers like "Basic Information" */
         color: #E0E0E0;
         font-weight: 600;
         margin-top: 0.2rem; /* Reduced top margin */
         margin-bottom: 0.4rem; /* Reduced bottom margin */
         font-size: 0.95em; /* Slightly smaller header */
    }
    .details-section p, .details-section div, .details-section span {
        font-size: 0.95rem !important; 
        color: #D0D0D0 !important;
    }
    .details-section code { 
        background-color: #333;
        color: #eee;
        padding: 0.2em 0.4em;
        border-radius: 3px;
    }
    
    .llm-image-container img {
        border: 1px solid #555;
        border-radius: 4px;
        max-height: 300px;
    }

    /* Enhanced image display styling */
    .stImage > div {
        border: 1px solid #444;
        border-radius: 8px;
        overflow: hidden;
        transition: all 0.1s ease-in-out;
    }
    .stImage > div:hover {
        border-color: #FF4B4B;
        box-shadow: 0 2px 8px rgba(255, 75, 75, 0.2);
    }

    /* Image loading spinner styling */
    .stSpinner > div {
        border-color: #FF4B4B !important;
    }

    /* Thinner HR */
    .details-section hr {
        border-top: 1px solid #383838 !important; /* Thinner and lighter color */
        margin-top: 0.5rem !important;
        margin-bottom: 0.5rem !important;
    }


    
    /* Styling for st.info box for LLM analysis to make it less prominent */
    .details-section .stAlert { /* Targets st.info, st.warning etc. within details */
        padding: 0.5rem 0.75rem !important; /* Reduced padding */
        font-size: 0.9em !important; /* Smaller font */
        border-radius: 4px !important;
    }
    .details-section .stAlert p { /* Target paragraph within the alert */
        font-size: 0.9em !important; /* Ensure paragraph text is also smaller */
        color: #D0D0D0 !important; /* Match other detail text color if needed */
    }

    /* Collapsible LLM Analysis section styling */
    .llm-analysis-container {
        background-color: #262730 !important;
        border: 1px solid #444 !important;
        border-radius: 8px !important;
        margin-bottom: 1rem !important;
        border-left: 3px solid #6366f1 !important;
        transition: all 0.1s ease-in-out !important;
    }

    .llm-analysis-container:hover {
        border-color: #8b5cf6 !important;
        box-shadow: 0 2px 8px rgba(139, 92, 246, 0.2) !important;
    }

    .llm-analysis-header {
        padding: 0.75rem 1rem !important;
        background-color: transparent !important;
    }

    .llm-analysis-content {
        padding: 1rem !important;
        border-top: 1px solid #444 !important;
        margin-top: 0.5rem !important;
        background-color: #1E1E1E !important;
        border-radius: 0 0 6px 6px !important;
    }

    /* LLM Analysis toggle button styling */
    .llm-analysis-container div[data-testid="stButton"] button {
        background-color: transparent !important;
        border: 1px solid #6366f1 !important;
        color: #6366f1 !important;
        font-size: 0.9em !important;
        padding: 4px 8px !important;
        border-radius: 4px !important;
        transition: all 0.1s ease !important;
    }

    .llm-analysis-container div[data-testid="stButton"] button:hover {
        background-color: #6366f1 !important;
        color: white !important;
        transform: scale(1.05) !important;
    }

    /* Status indicator styling */
    .llm-status-available {
        color: #A6E6A6 !important;
        font-weight: 600 !important;
    }

    .llm-status-processing {
        color: #FFDDAA !important;
        font-weight: 600 !important;
    }


    @media (max-width: 768px) {
        .verification-card { 
            padding: 1rem;
        }
        .verification-id-display {
            font-size: 1.1em !important; 
        }
        .details-section h6 {
            font-size: 0.9em;
        }
    }
    </style>
    """, unsafe_allow_html=True)


def render_verification_card(v, idx, status, verification_id, result_key, is_expanded, api_client):
    card_class = "verification-card card-expanded" if is_expanded else "verification-card"
    st.markdown(f'<div class="{card_class}">', unsafe_allow_html=True)

    col1, col2, col3 = st.columns([0.7, 3, 1.5])

    with col1:
        icon = "➖" if is_expanded else "➕"
        if st.button(icon, key=f"toggle_{result_key}", help="Expand/Collapse details", use_container_width=True):
            if is_expanded:
                st.session_state['verification_expanded_results'].discard(result_key)
            else:
                st.session_state['verification_expanded_results'].add(result_key)
            st.rerun()

    with col2:
        verification_date_str = v.get('verificationAt', 'N/A')
        verification_type = v.get("verificationType", "N/A")
        try:
            dt_obj = datetime.fromisoformat(verification_date_str.replace('Z', '+00:00'))
            formatted_date = dt_obj.strftime('%Y-%m-%d %H:%M')
        except ValueError:
            formatted_date = verification_date_str

        st.markdown(f"""
        <div style="background: transparent;">
            <strong style="font-size: 1.1em; color: #FAFAFA;">Verification #{idx + 1 + st.session_state.get('verification_page',0) * st.session_state.get('verification_last_limit',20)}</strong><br>
            <span class="verification-id-display">ID: {verification_id}</span>
            <small style="color: #A0A0A0; font-size: 0.9em;">Type: {verification_type}</small><br>
            <small style="color: #A0A0A0; font-size: 0.9em;">📅 {formatted_date}</small>
        </div>
        """, unsafe_allow_html=True)

    with col3:
        if status == "CORRECT":
            status_class = "status-correct"
            status_text = "✓ Correct"
        elif status == "INCORRECT":
            status_class = "status-incorrect"
            status_text = "✗ Incorrect"
        else:
            status_class = "status-unknown"
            status_text = "? Unknown"
        st.markdown(f'<div class="status-badge {status_class}">{status_text}</div>', unsafe_allow_html=True)
    
    if is_expanded:
        # Cache key for storing LLM analysis data
        llm_cache_key = f"llm_data_{verification_id}"

        # Fetch and cache LLM analysis data only once when verification card is first expanded
        if llm_cache_key not in st.session_state:
            with st.spinner("🔄 Loading verification details..."):
                logger.info(f"Fetching LLM analysis data for verification {verification_id} (first time)")

                # Fetch detailed verification data
                detailed_verification = get_detailed_verification_data(api_client, verification_id)

                # Fetch LLM analysis from conversation API
                llm_analysis_content, turn2_metadata = get_llm_analysis_from_conversation(api_client, verification_id)

                # Fallback to traditional analysis extraction if conversation API fails
                if not llm_analysis_content:
                    llm_analysis_content = extract_llm_analysis(v, detailed_verification)

                # Cache the data in session state
                st.session_state[llm_cache_key] = {
                    'detailed_verification': detailed_verification,
                    'llm_analysis_content': llm_analysis_content,
                    'turn2_metadata': turn2_metadata,
                    'cached_at': datetime.now().isoformat()
                }
                logger.info(f"Cached LLM analysis data for verification {verification_id}")
        else:
            logger.info(f"Using cached LLM analysis data for verification {verification_id}")

        # Retrieve cached data
        cached_data = st.session_state[llm_cache_key]
        detailed_verification = cached_data['detailed_verification']
        llm_analysis_content = cached_data['llm_analysis_content']
        turn2_metadata = cached_data['turn2_metadata']

        # The div with class "details-section" will have reduced padding via CSS
        st.markdown('<div class="details-section">', unsafe_allow_html=True)

        detail_cols = st.columns([2,2,1])
        with detail_cols[0]:
            # h6 CSS will make this smaller
            st.markdown("<h6>📋 Basic Information</h6>", unsafe_allow_html=True) 
            st.markdown(f'<p><strong>Machine ID:</strong> <code>{v.get("vendingMachineId", "N/A")}</code></p>', unsafe_allow_html=True)
            if verification_date_str != 'N/A':
                 st.markdown(f'<p><strong>Full Timestamp:</strong> {verification_date_str}</p>', unsafe_allow_html=True)

        with detail_cols[1]:
            st.markdown("<h6>📊 Performance</h6>", unsafe_allow_html=True)
            accuracy = v.get('overallAccuracy', None)
            correct_positions = v.get('correctPositions', None)
            discrepant_positions = v.get('discrepantPositions', None)

            if accuracy is not None:
                st.markdown(f"""
                <div class="metric-display">
                    <div class="metric-value">{accuracy}%</div>
                    <div class="metric-label">Accuracy</div>
                </div>
                """, unsafe_allow_html=True)
            
            metric_sub_cols = st.columns(2)
            with metric_sub_cols[0]:
                if correct_positions is not None:
                    st.markdown(f"""
                    <div class="metric-display">
                        <div class="metric-value">{correct_positions}</div>
                        <div class="metric-label">Correct Pos.</div>
                    </div>
                    """, unsafe_allow_html=True)
            with metric_sub_cols[1]:
                if discrepant_positions is not None:
                    st.markdown(f"""
                    <div class="metric-display">
                        <div class="metric-value">{discrepant_positions}</div>
                        <div class="metric-label">Discrepant Pos.</div>
                    </div>
                    """, unsafe_allow_html=True)
            if accuracy is None and correct_positions is None and discrepant_positions is None:
                 st.markdown("<p><small>No performance metrics available.</small></p>", unsafe_allow_html=True)

        with detail_cols[2]:
            st.markdown("<h6>⚙️ Actions</h6>", unsafe_allow_html=True)
            if st.button("📋 Copy ID", key=f"copy_{verification_id}_{idx}", use_container_width=True, help="Copy Verification ID to clipboard"):
                st.code(verification_id, language=None) 
                st.toast(f"ID {verification_id} copied to clipboard!", icon="✅")

        # HR CSS will make this thinner
        st.markdown("<hr>", unsafe_allow_html=True)

        # Verification Images section
        st.markdown("<h6>🖼️ Verification Images</h6>", unsafe_allow_html=True)

        # Pre-load image URLs for display
        image_urls_to_display = {}

        # Get reference image URL
        ref_s3_uri = v.get('referenceImageUrl')
        if ref_s3_uri:
            ref_bucket_name, ref_key = parse_s3_uri(ref_s3_uri)
            if ref_bucket_name and ref_key:
                ref_bucket_type = determine_bucket_type(ref_bucket_name)
                if ref_bucket_type and api_client:
                    try:
                        url_response = api_client.get_image_url(ref_key, ref_bucket_type)
                        ref_image_url = url_response.get('presignedUrl') or url_response.get('url')
                        if ref_image_url:
                            image_urls_to_display['reference'] = ref_image_url
                    except Exception as e_img:
                        logger.error(f"Error getting reference image URL for {ref_key}: {e_img}")

        # Get checking image URL
        chk_s3_uri = v.get('checkingImageUrl')
        if chk_s3_uri:
            chk_bucket_name, chk_key = parse_s3_uri(chk_s3_uri)
            if chk_bucket_name and chk_key:
                chk_bucket_type = determine_bucket_type(chk_bucket_name)
                if chk_bucket_type and api_client:
                    try:
                        url_response = api_client.get_image_url(chk_key, chk_bucket_type)
                        chk_image_url = url_response.get('presignedUrl') or url_response.get('url')
                        if chk_image_url:
                            image_urls_to_display['checking'] = chk_image_url
                    except Exception as e_img:
                        logger.error(f"Error getting checking image URL for {chk_key}: {e_img}")

        # Enhanced image display with better error handling
        llm_img_cols = st.columns(2)

        with llm_img_cols[0]:
            st.markdown("<p style='text-align:center; font-weight:bold; color: #E0E0E0;'>📷 Reference Image</p>", unsafe_allow_html=True)

            if 'reference' in image_urls_to_display:
                ref_s3_uri = v.get('referenceImageUrl')
                ref_bucket_name, ref_key = parse_s3_uri(ref_s3_uri) if ref_s3_uri else (None, None)

                st.image(
                    image_urls_to_display['reference'],
                    use_column_width=True,
                    caption=f"📁 {ref_key.split('/')[-1]}" if ref_key else "Reference Image"
                )
                # Show image details
                if ref_key:
                    st.markdown(f"<small style='color: #A0A0A0;'>🗂️ Path: {ref_key}</small>", unsafe_allow_html=True)
            else:
                st.info("ℹ️ No reference image available")

        with llm_img_cols[1]:
            st.markdown("<p style='text-align:center; font-weight:bold; color: #E0E0E0;'>🔍 Checking Image</p>", unsafe_allow_html=True)

            if 'checking' in image_urls_to_display:
                chk_s3_uri = v.get('checkingImageUrl')
                chk_bucket_name, chk_key = parse_s3_uri(chk_s3_uri) if chk_s3_uri else (None, None)

                st.image(
                    image_urls_to_display['checking'],
                    use_column_width=True,
                    caption=f"📁 {chk_key.split('/')[-1]}" if chk_key else "Checking Image"
                )
                # Show image details
                if chk_key:
                    st.markdown(f"<small style='color: #A0A0A0;'>🗂️ Path: {chk_key}</small>", unsafe_allow_html=True)
            else:
                st.info("ℹ️ No checking image available")

        # Enhanced verification details section
        st.markdown("---")
        st.markdown("<h6>🤖 Analysis & Additional Details</h6>", unsafe_allow_html=True)

        # Collapsible LLM Analysis section with progressive disclosure (using cached data)
        llm_analysis_key = f"llm_analysis_{verification_id}_{idx}"
        is_llm_expanded = st.session_state.get(llm_analysis_key, False)

        # Use cached LLM analysis data (already fetched when verification card was expanded)
        llm_cache_key = f"llm_data_{verification_id}"
        cached_llm_data = st.session_state.get(llm_cache_key, {})
        cached_llm_analysis_content = cached_llm_data.get('llm_analysis_content')
        cached_turn2_metadata = cached_llm_data.get('turn2_metadata')
        cached_detailed_verification = cached_llm_data.get('detailed_verification')

        # Create collapsible header with card-based design
        st.markdown("""
        <div style="background-color: #262730; border: 1px solid #444; border-radius: 8px; margin-bottom: 1rem; border-left: 3px solid #6366f1; transition: all 0.1s ease-in-out;">
        """, unsafe_allow_html=True)

        # Header with expand/collapse functionality
        header_col1, header_col2, header_col3 = st.columns([0.5, 4, 1])

        with header_col1:
            # Dropdown arrow indicator
            arrow_icon = "🔽" if is_llm_expanded else "▶️"
            if st.button(arrow_icon, key=f"toggle_llm_{llm_analysis_key}", help="Expand/Collapse LLM Analysis", use_container_width=True):
                st.session_state[llm_analysis_key] = not is_llm_expanded
                st.rerun()

        with header_col2:
            # Essential summary information (always visible)
            st.markdown("""
            <div style="padding: 0.5rem 0;">
                <strong style="color: #FAFAFA; font-size: 1.1em;">🧠 LLM Analysis</strong><br>
                <small style="color: #A0A0A0;">AI-powered verification analysis and insights</small>
            </div>
            """, unsafe_allow_html=True)

        with header_col3:
            # Status indicator using cached data
            if cached_llm_analysis_content:
                st.markdown('<div style="text-align: center; padding: 0.5rem 0;"><span style="color: #A6E6A6; font-size: 0.9em;">✅ Available</span></div>', unsafe_allow_html=True)
            else:
                st.markdown('<div style="text-align: center; padding: 0.5rem 0;"><span style="color: #FFDDAA; font-size: 0.9em;">⏳ Processing</span></div>', unsafe_allow_html=True)

        # Progressive disclosure - show full details when expanded (using cached data for instant display)
        if is_llm_expanded:
            st.markdown('<div style="padding: 1rem; border-top: 1px solid #444; margin-top: 0.5rem;">', unsafe_allow_html=True)

            if cached_llm_analysis_content:
                # Show metadata header if available
                if cached_turn2_metadata:
                    col1, col2 = st.columns([3, 1])
                    with col1:
                        st.markdown("**📋 Final Analysis Report**")
                    with col2:
                        st.markdown(f"<small style='color: #A0A0A0;'>Turn {cached_turn2_metadata.get('turn', 2)}</small>", unsafe_allow_html=True)

                # Display the analysis content with appropriate Streamlit components
                st.success(cached_llm_analysis_content)

                # Show additional context if available
                if cached_turn2_metadata:
                    with st.expander("📊 Analysis Metadata", expanded=False):
                        st.json(cached_turn2_metadata)
            else:
                st.info("Analysis not yet available or processing in progress.")
                st.markdown("""
                <div style="padding: 1rem; background-color: #1E1E1E; border-radius: 6px; border: 1px solid #333; margin-top: 0.5rem;">
                    <p style="color: #D0D0D0; margin: 0; font-size: 0.9em;">
                        <strong>ℹ️ Analysis Status:</strong> The AI analysis for this verification is still being processed.
                        Please check back in a few moments for detailed insights.
                    </p>
                </div>
                """, unsafe_allow_html=True)

            st.markdown('</div>', unsafe_allow_html=True)

        st.markdown('</div>', unsafe_allow_html=True)

        # Raw data section (collapsible) using cached detailed verification data
        raw_data_available = False
        raw_result_data = v.get('result') or (cached_detailed_verification.get('result') if cached_detailed_verification else None)
        raw_summary_data = v.get('verificationSummary') or (cached_detailed_verification.get('verificationSummary') if cached_detailed_verification else None)

        if raw_result_data or raw_summary_data:
            raw_data_available = True

        if raw_data_available:
            with st.expander("🔍 Show Raw Verification Data", expanded=False):
                if raw_result_data:
                    st.write("**Result Details:**")
                    st.json(raw_result_data)
                if raw_summary_data:
                    st.write("**Summary Details:**")
                    st.json(raw_summary_data)
        
        st.markdown('</div>', unsafe_allow_html=True) 

    st.markdown('</div>', unsafe_allow_html=True)
