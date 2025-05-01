variable "reference_bucket_name" {
  description = "Name of the S3 bucket for reference layout images"
  type        = string
}

variable "checking_bucket_name" {
  description = "Name of the S3 bucket for checking images"
  type        = string
}

variable "results_bucket_name" {
  description = "Name of the S3 bucket for verification results"
  type        = string
}

variable "force_destroy" {
  description = "Boolean that indicates all objects should be deleted from the bucket when the bucket is destroyed"
  type        = bool
  default     = false
}

variable "reference_lifecycle_rules" {
  description = "Lifecycle rules for the reference bucket"
  type = list(object({
    id                                     = string
    enabled                                = bool
    prefix                                 = optional(string)
    expiration_days                        = optional(number)
    noncurrent_version_expiration_days     = optional(number)
    abort_incomplete_multipart_upload_days = optional(number)
  }))
  default = []
}

variable "checking_lifecycle_rules" {
  description = "Lifecycle rules for the checking bucket"
  type = list(object({
    id                                     = string
    enabled                                = bool
    prefix                                 = optional(string)
    expiration_days                        = optional(number)
    noncurrent_version_expiration_days     = optional(number)
    abort_incomplete_multipart_upload_days = optional(number)
  }))
  default = []
}

variable "results_lifecycle_rules" {
  description = "Lifecycle rules for the results bucket"
  type = list(object({
    id                                     = string
    enabled                                = bool
    prefix                                 = optional(string)
    expiration_days                        = optional(number)
    noncurrent_version_expiration_days     = optional(number)
    abort_incomplete_multipart_upload_days = optional(number)
  }))
  default = []
}

variable "common_tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}