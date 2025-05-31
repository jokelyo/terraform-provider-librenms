terraform {
  required_providers {
    librenms = {
      source = "hashicorp.com/edu/librenms"
    }
  }
}

provider "librenms" {}

data "librenms_device" "example" {}
