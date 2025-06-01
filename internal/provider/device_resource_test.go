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
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				// ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "librenms_device" "test" {
  hostname = "1.1.1.1"
  port     = 163
  icmp_only = {}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify device updated
					resource.TestCheckResourceAttr("librenms_device.test", "port", "163"),
					//resource.TestCheckResourceAttr("librenms_device.test", "snmp_v2c.community", "test2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
