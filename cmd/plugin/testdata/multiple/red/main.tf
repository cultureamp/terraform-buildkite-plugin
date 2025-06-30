terraform {
  required_version = ">= 1.0.0"

  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

provider "local" {}

resource "local_file" "example-01" {
  filename = "${path.module}/output/01.txt"
  content  = var.file_content
}

variable "file_content" {
  default = "Red is the best color!"
}
