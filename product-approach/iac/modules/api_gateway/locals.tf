# modules/api_gateway/locals.tf

locals {
  # CORS configuration
  cors_origin = var.cors_enabled ? (var.streamlit_service_url != "" ? var.streamlit_service_url : "*") : ""
}