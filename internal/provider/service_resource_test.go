package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceResource(t *testing.T) {
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

resource "librenms_service" "test" {
  device_id = librenms_device.test.id
  name = "service test"

  ignore = false
  parameters = "-t 10 -c 5"
  target = "1.1.1.1"
  type = "ping"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("librenms_service.test", "target", "1.1.1.1"),
					resource.TestCheckResourceAttr("librenms_service.test", "name", "service test"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("librenms_service.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "librenms_service.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the LibreNMS
				// API, therefore there is no value for it during import.
				// ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "librenms_device" "test" {
  hostname = "1.1.1.1"
  port     = 161
  icmp_only = {}
  force_add = true
}

resource "librenms_service" "test" {
  device_id = librenms_device.test.id
  name = "service test"

  ignore = false
  parameters = "-t 10 -c 5"
  target = "1.1.1.2"
  type = "ping"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify service updated
					resource.TestCheckResourceAttr("librenms_service.test", "target", "1.1.1.2"),
					//resource.TestCheckResourceAttr("librenms_service.test", "snmp_v2c.community", "test2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
