# create a static device group with a single device using the computed resource id
resource "librenms_devicegroup" "my_group" {
  name = "my_group"
  type = "static"

  devices = [
    librenms_device.device.id,
  ]
}

# create a dynamic device group with devices that have a sysDescr containing "cloud"
resource "librenms_devicegroup" "my_dynamic_group" {
  name        = "my_dynamic_group"
  description = "includes devices with sysDescr containing 'cloud'"
  type        = "dynamic"

  rules = {
    "condition" : "AND",
    "rules" : [
      {
        "id" : "devices.sysDescr",
        "field" : "devices.sysDescr",
        "operator" : "contains",
        "value" : "cloud"
      }
    ]
  }

  # complex, nested rulesets can be configured using rules_json.
  # rules_json = jsonencode({....})
}
