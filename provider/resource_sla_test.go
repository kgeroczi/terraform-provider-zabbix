package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSLA(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSLABasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_sla.test", "period", "1"),
					resource.TestCheckResourceAttr("zabbix_sla.test", "slo", "99.9"),
				),
			},
			{
				Config: testAccResourceSLAUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_sla.test", "period", "2"),
					resource.TestCheckResourceAttr("zabbix_sla.test", "description", "updated by acceptance test"),
				),
			},
		},
	})
}

func testAccResourceSLABasic() string {
	return `
resource "zabbix_sla" "test" {
  name        = "tf-acc-sla"
  period      = 1
  slo         = 99.9
  timezone    = "UTC"
  description = "created by acceptance test"
}
`
}

func testAccResourceSLAUpdated() string {
	return `
resource "zabbix_sla" "test" {
  name        = "tf-acc-sla"
  period      = 2
  slo         = 99.5
  timezone    = "UTC"
  description = "updated by acceptance test"
}
`
}
