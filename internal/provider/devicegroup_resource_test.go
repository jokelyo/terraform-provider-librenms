// Copyright (c) HashiCorp, Inc.

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
resource "librenms_devicegroup" "test1" {
  name = "test group"
  type = "dynamic"
  rules_json = "{\"condition\":\"AND\",\"rules\":[{\"id\":\"access_points.name\",\"field\":\"access_points.name\",\"type\":\"string\",\"input\":\"text\",\"operator\":\"equal\",\"value\":\"accesspoint1\"}],\"valid\":true}"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("librenms_devicegroup.test1", "name", "test group"),
					resource.TestCheckResourceAttr("librenms_devicegroup.test1", "type", "dynamic"),
					// Verify dynamic values have any value set in the state.
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
resource "librenms_devicegroup" "test1" {
  name = "test group"
  description = "This is a test group"
  type = "dynamic"
  rules_json = "{\"condition\":\"AND\",\"rules\":[{\"id\":\"access_points.name\",\"field\":\"access_points.name\",\"type\":\"string\",\"input\":\"text\",\"operator\":\"equal\",\"value\":\"accesspoint1\"}],\"valid\":true}"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify device group updated
					resource.TestCheckResourceAttr("librenms_devicegroup.test1", "description", "This is a test group"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
