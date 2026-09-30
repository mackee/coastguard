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

variable "allowed_domains" {
  description = "Google Workspace domains (hd claim) allowed to access. Users matching either allowed_domains or allowed_emails are allowed. If both are empty, all authenticated users are allowed."
  type        = list(string)
  default     = []
}

variable "allowed_emails" {
  description = "Email addresses allowed to access. Users matching either allowed_domains or allowed_emails are allowed. If both are empty, all authenticated users are allowed."
  type        = list(string)
  default     = []
}
