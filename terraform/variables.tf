variable "aws_region" {
  description = "AWS region for all resources."
  type        = string
  default     = "ap-southeast-2"
}

variable "stage_name" {
  description = "Deployment stage name."
  type        = string
  default     = "prod"
}

variable "canvas_base_url" {
  description = "Canvas API base url."
  type        = string
}

variable "canvas_page_size" {
  description = "Page size while fetching from Canvas API."
  type        = string
}

variable "canvas_access_token" {
  description = "Canvas admin account's access token."
  type        = string
}