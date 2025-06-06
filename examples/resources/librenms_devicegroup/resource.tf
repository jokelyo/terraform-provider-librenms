# create a dynamic device group with devices
# that have a sysDescr containing "cloud"
#
# Due to the complicated structure of the rulesets,
# I would recommend initially configuring the dynamic
# rules in the LibreNMS UI and then exporting from
# the API to get the correct JSON format.
#
# Example to get formatted rules from an existing group.
# Change '0' in the jq command to the array index of the
# appropriate group:
#
#   curl -H "X-Auth-Token: token" \
#     http://localhost:8000/api/v0/devicegroups | \
#     jq '.groups.[0].rules | fromjson'
#
# You can also just use the serialized JSON string
# as it's natively represented in the LibreNMS API output,
# but the readability is worse.
#
# rules = "{\"condition\": \"AND\", ....}"
#
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
