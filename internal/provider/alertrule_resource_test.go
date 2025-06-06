package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const alertRuleSetupConfig = `
resource "librenms_device" "test_device1" {
  hostname  = "192.168.5.5"
  force_add = true

  snmp_v2c = {
    community = "public"
  }
}
resource "librenms_device" "test_device2" {
  hostname  = "192.168.5.6"
  force_add = true

  snmp_v2c = {
    community = "public"
  }
}

resource "librenms_devicegroup" "test1" {
  name = "test group static"
  type = "static"

  # Add out of order to verify it doesn't affect plan
  devices = [
    librenms_device.test_device2.id,
	librenms_device.test_device1.id,
  ]
}
resource "librenms_devicegroup" "test2" {
  name = "test group static 2"
  type = "static"

  # Add out of order to verify it doesn't affect plan
  devices = [
    librenms_device.test_device2.id,
	librenms_device.test_device1.id,
  ]
}

resource "librenms_location" "test_location" {
  name = "test location"

  fixed_coordinates = true
  latitude = -45.0862462
  longitude = 37.4220648
}
resource "librenms_location" "test_location2" {
  name = "test location 2"

  fixed_coordinates = true
  latitude = -35.0862462
  longitude = 47.4220648
}
`

func TestAccAlertRuleGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + alertRuleSetupConfig + `
resource "librenms_alertrule" "testrule" {
  name  = "Test Rule (ICMP)"
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
}

resource "librenms_alertrule" "testrule2" {
  name  = "Test Rule (ICMP) Device Set"
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

  devices = [
    librenms_device.test_device2.id,
    librenms_device.test_device1.id
  ]
}

resource "librenms_alertrule" "testrule3" {
  name  = "Test Rule (ICMP) Group/Location Set"
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

  groups = [
	librenms_devicegroup.test2.id,
	librenms_devicegroup.test1.id
  ]

  locations = [
    librenms_location.test_location.id,
	librenms_location.test_location2.id
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("librenms_alertrule.testrule", "name", "Test Rule (ICMP)"),
					resource.TestCheckResourceAttr("librenms_alertrule.testrule", "severity", "critical"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("librenms_alertrule.testrule", "id"),
					resource.TestCheckResourceAttrSet("librenms_alertrule.testrule", "query"),
					resource.TestCheckResourceAttrSet("librenms_alertrule.testrule2", "id"),
					resource.TestCheckResourceAttrSet("librenms_alertrule.testrule2", "query"),
					resource.TestCheckResourceAttrSet("librenms_alertrule.testrule3", "id"),
					resource.TestCheckResourceAttrSet("librenms_alertrule.testrule3", "query"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "librenms_alertrule.testrule",
				ImportState:       true,
				ImportStateVerify: true,
				// These values are stored in the `extra` field in LibreNMS API after the create operation.
				ImportStateVerifyIgnore: []string{"delay", "interval", "max_alerts", "mute"},
			},
			// Update and Read testing
			{
				Config: providerConfig + alertRuleSetupConfig + `
resource "librenms_alertrule" "testrule" {
  name  = "Test Rule (ICMP)"
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
  max_alerts = 3

  disabled = false
  severity = "critical"
}

resource "librenms_alertrule" "testrule2" {
  name  = "Test Rule (ICMP) Device Set"
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
}

resource "librenms_alertrule" "testrule3" {
  name  = "Test Rule (ICMP) Location Set"

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
  severity = "warning"

  groups = [
	librenms_devicegroup.test2.id,
	librenms_devicegroup.test1.id
  ]

  locations = [
    librenms_location.test_location.id,
	librenms_location.test_location2.id
  ]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify testrule updated
					resource.TestCheckResourceAttr("librenms_alertrule.testrule", "max_alerts", "3"),
					// Verify testrule2 updated
					resource.TestCheckResourceAttr("librenms_alertrule.testrule2", "devices.#", "0"),
					// Verify testrule3 updated
					resource.TestCheckResourceAttr("librenms_alertrule.testrule3", "severity", "warning"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
