terraform {
  required_providers {
    librenms = {
      source = "hashicorp.com/edu/librenms"
    }
  }
}

provider "librenms" {
  host = "http://localhost:8000/"
  token = "poop"
}

// data "librenms_device" "example" {}
