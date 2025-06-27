# modules/api_gateway/locals.tf

locals {
  # CORS configuration - Use wildcard by default to avoid dependency cycle
  cors_origin = var.cors_enabled ? "*" : ""
}
