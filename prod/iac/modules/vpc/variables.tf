variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones to use for the subnets"
  type        = list(string)
}

variable "name_prefix" {
  description = "Prefix to use for resource names"
  type        = string
}

variable "common_tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "create_nat_gateway" {
  description = "Whether to create a NAT Gateway for private subnets"
  type        = bool
  default     = true
}

variable "container_port" {
  description = "Port on which the container is listening"
  type        = number
  default     = 3000  # Default React port
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = ""
}

variable "name_suffix" {
  description = "Suffix to append to resource names"
  type        = string
  default     = ""
}
