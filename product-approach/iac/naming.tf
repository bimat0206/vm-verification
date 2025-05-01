locals {
  # Generate resource name based on standard pattern
  generate_name = function(resource_type, resource_name) {
    lower(join("-", compact([local.name_prefix, resource_type, resource_name, local.name_suffix])))
  }
}
