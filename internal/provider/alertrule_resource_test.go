package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAlertRuleGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "librenms_alertrule" "testrule" {
  name  = "Cloud Devices Down (ICMP)"
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
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("librenms_alertrule.testrule", "name", "Cloud Devices Down (ICMP)"),
					resource.TestCheckResourceAttr("librenms_alertrule.testrule", "severity", "critical"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("librenms_alertrule.testrule", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "librenms_alertrule.testrule",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the LibreNMS
				// API, therefore there is no value for it during import.
				// ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "librenms_alertrule" "testrule" {
  name  = "Cloud Devices Down (ICMP)"
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
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify device group updated
					resource.TestCheckResourceAttr("librenms_alertrule.testrule", "max_alerts", "3"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
