package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLocationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "librenms_device" "test_device" {
  hostname  = "192.168.5.5"
  force_add = true

  location = librenms_location.test_location.name

  snmp_v2c = {
    community = "public"
  }
}

resource "librenms_location" "test_location" {
  name = "test location"

  fixed_coordinates = true
  latitude = -45.0862462
  longitude = 37.4220648
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("librenms_location.test_location", "latitude", "-45.0862462"),
					resource.TestCheckResourceAttr("librenms_location.test_location", "name", "test location"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("librenms_location.test_location", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "librenms_location.test_location",
				ImportState:       true,
				ImportStateVerify: true,
				// The timestamp attribute is a computed value.
				ImportStateVerifyIgnore: []string{"timestamp"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "librenms_device" "test_device" {
  hostname  = "192.168.5.5"
  force_add = true

  location = librenms_location.test_location.name

  snmp_v2c = {
    community = "public"
  }
}

resource "librenms_location" "test_location" {
  name = "test location"

  fixed_coordinates = true
  latitude = -35.0862462
  longitude = 37.4220648
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify location updated
					resource.TestCheckResourceAttr("librenms_location.test_location", "latitude", "-35.0862462"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
