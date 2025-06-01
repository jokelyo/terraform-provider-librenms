terraform {
  required_providers {
    librenms = {
      source  = "hashicorp.com/edu/librenms"
    }
  }
  required_version = ">= 0.1.0"
}

provider "librenms" {
  host     = "http://localhost:8000"
  token    = "ac6d3047169b90688ad25e85f269ee08"
}

resource "librenms_device" "compute_vm_2" {
  hostname = "34.123.68.95"

  snmp_v2c = {
    community = "5581eb63764a093c"
  }
}

output "librenms_device_compute_vm_2" {
  value = librenms_device.compute_vm_2
}
