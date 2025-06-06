package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeviceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "librenms_device" "test" {
  hostname = "1.1.1.1"
  port     = 161
  icmp_only = {}
  force_add = true
}

resource "librenms_device" "test2" {
  hostname = "1.1.1.2"
  display = "Test Device 2"

  snmp_v1 = {
    community = "test"
  }
  force_add = true
}

resource "librenms_device" "test3" {
  hostname = "1.1.1.3"

  snmp_v2c = {
    community = "test"
  }
  force_add = true
}

resource "librenms_device" "test4" {
  hostname = "1.1.1.4"

  snmp_v3 = {
    auth_algorithm = "SHA"
    auth_level = "authPriv"
    auth_name = "user"
    auth_pass = "test"
    crypto_algorithm = "AES"
    crypto_pass = "test"
  }
  force_add = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("librenms_device.test", "hostname", "1.1.1.1"),
					resource.TestCheckResourceAttr("librenms_device.test", "port", "161"),
					//resource.TestCheckResourceAttr("librenms_device.test", "snmp_v2c.community", "test"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("librenms_device.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "librenms_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The force_add attribute does not exist in the LibreNMS
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"force_add"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "librenms_device" "test" {
  hostname = "1.1.1.1"
  display = "Test Device"

  port     = 163
  icmp_only = {}
  force_add = true
}

resource "librenms_device" "test2" {
  hostname = "1.1.1.2"
  port     = 161
  snmp_v1 = {
    community = "test"
  }
  force_add = true
}

resource "librenms_device" "test3" {
  hostname = "1.1.1.3"
  port     = 161
  snmp_v1 = {
    community = "test"
  }
  force_add = true
}

resource "librenms_device" "test4" {
  hostname = "1.1.1.4"

  snmp_v3 = {
    auth_algorithm = "SHA"
    auth_level = "authPriv"
    auth_name = "user"
    auth_pass = "test"
    crypto_algorithm = "AES"
    crypto_pass = "test2"
  }
  force_add = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify test device updated
					resource.TestCheckResourceAttr("librenms_device.test", "display", "Test Device"),
					resource.TestCheckResourceAttr("librenms_device.test", "port", "163"),
					// Verify test3 updated
					resource.TestCheckResourceAttr("librenms_device.test3", "snmp_v1.community", "test"),
					// Verify test4 updated
					resource.TestCheckResourceAttr("librenms_device.test4", "snmp_v3.crypto_algorithm", "AES"),
					//resource.TestCheckResourceAttr("librenms_device.test", "snmp_v2c.community", "test2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
