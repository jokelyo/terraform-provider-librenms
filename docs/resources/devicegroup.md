# librenms_devicegroup (Resource)

## Example Usage

```terraform
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
```

## LibreNMS Rule Definitions

Due to the complicated structure of the rulesets,
I would recommend initially configuring the dynamic
rules in the LibreNMS UI and then exporting from
the API to get the correct JSON format.

Example to get formatted rule output from an existing rule.
Update `0` in the jq command to the relevant device group ID:
```shell
   curl -H "X-Auth-Token: token" \
     https://librenms.mydomain.com/api/v0/devicegroups | \
     jq '.groups[0].rules'
```

You can also just use a serialized JSON string but the readability is worse:

```terraform
  rules = "{\"condition\": \"AND\", ....}"
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The device group name.
- `type` (String) The device group type, [`dynamic`, `static`].

### Optional

- `description` (String) The device group description.
- `devices` (Set of Number) The set of device IDs in the group. This is only applicable for static device groups.
- `rules` (String) The rules for dynamic device groups, in serialized JSON format. This is only applicable for dynamic device groups. Using an encoded string supports the arbitrarily-deep nested structure of the LibreNMS rulesets.

### Read-Only

- `id` (Number) The unique numeric identifier of the LibreNMS device group.



## Import

Import is supported using the following syntax:

```shell
# Device Groups can be imported by specifying their numeric identifier.
terraform import librenms_devicegroup.example 123
```
