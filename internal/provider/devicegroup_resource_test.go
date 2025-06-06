package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
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

# Create a device group with static rules
resource "librenms_devicegroup" "test0" {
  name = "test group static"
  type = "static"

  # Add out of order to verify it doesn't affect plan
  devices = [
    librenms_device.test_device2.id,
    librenms_device.test_device1.id,
  ]
}

# Create a device group with dynamic rules
resource "librenms_devicegroup" "test1" {
  name = "test group dynamic"
  type = "dynamic"
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
    "joins": [],
    "valid" : true
  })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// check dynamic group attributes
					resource.TestCheckResourceAttr("librenms_devicegroup.test0", "name", "test group static"),
					resource.TestCheckResourceAttr("librenms_devicegroup.test0", "type", "static"),
					resource.TestCheckResourceAttr("librenms_devicegroup.test0", "devices.#", "2"),
					// check static group attributes
					resource.TestCheckResourceAttr("librenms_devicegroup.test1", "name", "test group dynamic"),
					resource.TestCheckResourceAttr("librenms_devicegroup.test1", "type", "dynamic"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("librenms_devicegroup.test0", "id"),
					resource.TestCheckResourceAttrSet("librenms_devicegroup.test1", "id"),
				),
			},
			// ImportState testing
			//{
			//	ResourceName:      "librenms_devicegroup.test1",
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//	// The last_updated attribute does not exist in the LibreNMS
			//	// API, therefore there is no value for it during import.
			//	// ImportStateVerifyIgnore: []string{"last_updated"},
			//},
			// Update and Read testing
			{
				Config: providerConfig + `
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

# Modify member list
resource "librenms_devicegroup" "test0" {
  name = "test group"
  type = "static"
  devices = [
	librenms_device.test_device1.id,
  ]
}

resource "librenms_devicegroup" "test1" {
  name = "test group dynamic"
  description = "This is a test group"
  type = "dynamic"
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
    "joins": [],
    "valid" : true
  })
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify device groups updated
					resource.TestCheckResourceAttr("librenms_devicegroup.test0", "devices.#", "1"),
					resource.TestCheckResourceAttr("librenms_devicegroup.test1", "description", "This is a test group"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
