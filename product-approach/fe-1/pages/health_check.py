import streamlit as st
import logging
import json

logger = logging.getLogger(__name__)

def app(api_client):
    st.title("Health Check")
    if st.button("Check System Health"):
        try:
            health_status = api_client.health_check()
            st.write("System Health:")
            
            # Display overall status with color
            status = health_status.get('status', 'unknown')
            if status == 'healthy':
                st.success(f"Overall Status: {status}")
            elif status == 'degraded':
                st.warning(f"Overall Status: {status}")
            else:
                st.error(f"Overall Status: {status}")
            
            # Display service details in expandable sections
            services = health_status.get('services', {})
            if services:
                st.subheader("Services Status")
                for service_name, service_info in services.items():
                    service_status = service_info.get('status', 'unknown')
                    col1, col2, col3 = st.columns([2, 1, 1])
                    with col1:
                        st.write(f"**{service_name.capitalize()}**")
                    with col2:
                        if service_status == 'healthy':
                            st.success(service_status)
                        elif service_status == 'degraded':
                            st.warning(service_status)
                        else:
                            st.error(service_status)
                    with col3:
                        if 'latency' in service_info:
                            st.write(f"{service_info['latency']} ms")
            
            # Display version and timestamp
            st.text(f"Version: {health_status.get('version', 'unknown')}")
            st.text(f"Timestamp: {health_status.get('timestamp', 'unknown')}")
            
        except Exception as e:
            logger.error(f"Health check failed: {str(e)}")
            st.error(str(e))