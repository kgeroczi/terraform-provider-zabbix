package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceService(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServiceBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_service.test", "algorithm", "1"),
					resource.TestCheckResourceAttr("zabbix_service.test", "weight", "10"),
				),
			},
			{
				Config: testAccResourceServiceUpdated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_service.test", "algorithm", "2"),
					resource.TestCheckResourceAttr("zabbix_service.test", "description", "updated by acceptance test"),
				),
			},
		},
	})
}

func testAccResourceServiceBasic() string {
	return `
resource "zabbix_service" "test" {
  name        = "tf-acc-service"
  algorithm   = 1
  weight      = 10
  description = "created by acceptance test"
}
`
}

func testAccResourceServiceUpdated() string {
	return `
resource "zabbix_service" "test" {
  name        = "tf-acc-service"
  algorithm   = 2
  weight      = 20
  description = "updated by acceptance test"
}
`
}
