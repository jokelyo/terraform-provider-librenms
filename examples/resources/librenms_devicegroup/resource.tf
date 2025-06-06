# create a dynamic device group with devices
# that have a sysDescr containing "cloud"
resource "librenms_devicegroup" "my_dynamic_group" {
  name        = "my_dynamic_group"
  description = "my cloud devices"
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
    "valid" : true,
  })
}

# create a static device group with a single device
resource "librenms_devicegroup" "my_group" {
  name = "my_group"
  type = "static"

  devices = [
    librenms_device.device.id,
  ]
}
