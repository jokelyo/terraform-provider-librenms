terraform {
  required_providers {
    librenms = {
      source = "registry.terraform.io/jokelyo/librenms"
    }
  }
}

provider "librenms" {
  host  = "http://localhost:8000/"
  token = "poop"
}

// data "librenms_device" "example" {}
