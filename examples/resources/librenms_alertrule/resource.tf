# Example: create an alert rule for down devices
#
# Due to the complicated structure of the rulesets,
# I would recommend initially configuring the builder
# ruleset in the LibreNMS UI and then exporting from
# the API to get the correct JSON format.
#
# Example to get formatted builder output from an existing rule:
#   curl -H "X-Auth-Token: token" \
#   http://localhost:8000/api/v0/rules/2 | \
#   jq '.rules.[0].builder | fromjson'
#
# You can also just use the serialized JSON string
# as it's natively represented in the LibreNMS API output,
# but the readability is worse.
#
# builder = "{\"condition\": \"AND\", ....}"
#
resource "librenms_alertrule" "rule1" {
  name  = "Device Down (ICMP)"
  notes = "Alert when a device is down and the reason is ICMP"

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

  delay      = "11m"
  interval   = "5m"
  max_alerts = 1

  disabled = false
  severity = "critical"

  # defaults to all devices if devices is not defined
  # devices = [1, 2]
}
