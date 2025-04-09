terraform {
  required_version = "~> 1.11.3"

  required_providers {
    aws = {
      version = "~> 5.94.1"
    }
  }

  backend "s3" {}
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      ManagedBy = "Terraform"
      Service   = var.project_name
      Repo      = var.repo
    }
  }
}
