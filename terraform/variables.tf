variable "region" {
  description = "The AWS region to deploy to"
}

variable "project_name" {
  description = "The name of the project"
  type        = string
  default     = "coastguard-demo"
}

variable "repo" {
  description = "The name of the repo"
  type        = string
  default     = "github.com/mackee/coastguard"
}
