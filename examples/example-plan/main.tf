terraform {
  required_providers {
    librenms = {
      source = "registry.terraform.io/jokelyo/librenms"
    }
  }
  required_version = ">= 0.1.0"
}

# See https://github.com/jokelyo/librenms-dev which:
# - Provides a local LibreNMS instance in Docker for testing purposes
# - Includes a terraform plan for creating a test GCP environment
#
provider "librenms" {
  host = "http://localhost:8000"
  # providing token using LIBRENMS_TOKEN environment variable
  # token = "token"
}

resource "librenms_device" "compute_vm_2" {
  hostname = "34.123.68.95"

  snmp_v2c = {
    community = "5581eb63764a093c"
  }
}

# create a static device group with a single device using the computed resource id
resource "librenms_devicegroup" "farts" {
  name        = "farts"
  description = "fartastic"
  type        = "static"
  devices = [
    librenms_device.compute_vm_2.id,
  ]
}

# create a dynamic device group with devices that have a sysDescr containing "cloud"
resource "librenms_devicegroup" "dynamic_farts" {
  name        = "dynamic_farts"
  description = "fartastic"
  type        = "dynamic"

  rules = jsonencode({
    "condition" : "AND",
    "rules" : [
      {
        "id" : "devices.sysDescr",
        "field" : "devices.sysDescr",
        "operator" : "contains",
        "value" : "cloud"
      }
    ],
    "joins" : [],
    "valid" : true
  })
}

resource "librenms_alertrule" "cloud_device_down_icmp" {
  name  = "Cloud Devices Down (ICMP)"
  notes = "Alert when a device is down and the reason is ICMP"

  # Due to their complicated structure, I would recommend initially configuring the rule in the LibreNMS UI
  # and then exporting from the API to get the correct JSON format.
  builder = jsonencode({
    "condition" : "AND",
    "rules" : [
      {
        "id" : "macros.device_down",
        "field" : "macros.device_down",
        "type" : "integer",
        "input" : "radio",
        "operator" : "equal",
        "value" : "1"
      },
      {
        "id" : "devices.status_reason",
        "field" : "devices.status_reason",
        "type" : "string",
        "input" : "text",
        "operator" : "equal",
        "value" : "icmp"
      }
    ],
    "valid" : true
  })

  # you can also just use a serialized JSON string as it's represented in the LibreNMS API output
  # builder = "{\"condition\": \"AND\",....}"

  delay      = "11m"
  interval   = "5m"
  max_alerts = 1

  disabled = false
  severity = "critical"

  # defaults to all devices if devices is not defined
  # devices = [1, 2]

  # can also specify group IDs to limit the alert rule to specific device groups
  groups = [
    librenms_devicegroup.dynamic_farts.id,
  ]
}

output "librenms_device_compute_vm_2" {
  value = librenms_device.compute_vm_2.hostname
}

output "librenms_devicegroup_farts" {
  value = librenms_devicegroup.farts
}

output "librenms_devicegroup_dynamic_farts" {
  value = librenms_devicegroup.dynamic_farts
}
