import streamlit as st
import logging
from datetime import datetime, timedelta

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Verifications")
    
    # Filter controls
    with st.expander("Filters", expanded=True):
        col1, col2 = st.columns(2)
        
        with col1:
            status_filter = st.selectbox(
                "Status", 
                ["all", "CORRECT", "INCORRECT"],
                format_func=lambda x: "All" if x == "all" else x
            )
            
            vending_machine_id = st.text_input("Vending Machine ID")
        
        with col2:
            date_option = st.selectbox(
                "Date Range",
                ["All time", "Last 24 hours", "Last 7 days", "Last 30 days", "Custom"],
            )
            
            if date_option == "Custom":
                from_date = st.date_input("From Date", datetime.now() - timedelta(days=7))
                to_date = st.date_input("To Date", datetime.now())
    
    # Build query parameters
    params = {}
    
    if status_filter != "all":
        params["verificationStatus"] = status_filter
        
    if vending_machine_id:
        params["vendingMachineId"] = vending_machine_id
        
    if date_option != "All time":
        to_date_str = datetime.now().isoformat()
        
        if date_option == "Last 24 hours":
            from_date_str = (datetime.now() - timedelta(days=1)).isoformat()
        elif date_option == "Last 7 days":
            from_date_str = (datetime.now() - timedelta(days=7)).isoformat()
        elif date_option == "Last 30 days":
            from_date_str = (datetime.now() - timedelta(days=30)).isoformat()
        elif date_option == "Custom":
            from_date_str = datetime.combine(from_date, datetime.min.time()).isoformat()
            to_date_str = datetime.combine(to_date, datetime.max.time()).isoformat()
            
        params["fromDate"] = from_date_str
        params["toDate"] = to_date_str
    
    # Pagination controls
    col1, col2 = st.columns(2)
    with col1:
        limit = st.number_input("Results per page", min_value=5, max_value=100, value=20)
        params["limit"] = limit
    
    with col2:
        page = st.session_state.get('verification_page', 0)
        offset = page * limit
        params["offset"] = offset
    
    # Sorting
    sort_options = {
        "newest": "verificationAt:desc",
        "oldest": "verificationAt:asc",
        "accuracy_high": "overallAccuracy:desc",
        "accuracy_low": "overallAccuracy:asc"
    }
    
    sort_by = st.selectbox(
        "Sort by",
        list(sort_options.keys()),
        format_func=lambda x: {
            "newest": "Newest first",
            "oldest": "Oldest first", 
            "accuracy_high": "Highest accuracy first",
            "accuracy_low": "Lowest accuracy first"
        }[x]
    )
    
    params["sortBy"] = sort_options[sort_by]
    
    # Fetch verifications
    try:
        response = api_client.list_verifications(params)
        verifications = response.get("results", [])
        pagination = response.get("pagination", {})
        
        # Display pagination info
        total = pagination.get("total", 0)
        st.write(f"Showing {offset+1}-{min(offset+limit, total)} of {total} verifications")
        
        # Pagination controls
        col1, col2, col3 = st.columns([1, 3, 1])
        with col1:
            if page > 0:
                if st.button("Previous"):
                    st.session_state['verification_page'] = page - 1
                    st.rerun()
        
        with col3:
            if pagination.get("nextOffset") is not None:
                if st.button("Next"):
                    st.session_state['verification_page'] = page + 1
                    st.rerun()
        
        # Display verifications
        if not verifications:
            st.info("No verifications found matching the criteria.")
        else:
            for v in verifications:
                with st.container():
                    col1, col2, col3 = st.columns([3, 2, 1])
                    
                    with col1:
                        verification_id = v.get('verificationId', 'N/A')
                        st.subheader(f"ID: {verification_id}")
                        st.write(f"Machine: {v.get('vendingMachineId', 'N/A')}")
                        st.write(f"Date: {v.get('verificationAt', 'N/A')}")
                    
                    with col2:
                        status = v.get('verificationStatus', 'UNKNOWN')
                        if status == 'CORRECT':
                            st.success(f"Status: {status}")
                        else:
                            st.error(f"Status: {status}")
                        
                        accuracy = v.get('overallAccuracy', 0)
                        st.metric("Accuracy", f"{accuracy}%")
                        st.write(f"Correct: {v.get('correctPositions', 0)}")
                        st.write(f"Discrepant: {v.get('discrepantPositions', 0)}")
                    
                    with col3:
                        # Details functionality removed - verification details page no longer available
                        st.write("Details view removed")
                
                st.divider()
                
    except Exception as e:
        logger.error(f"List verifications failed: {str(e)}")
        st.error(str(e))