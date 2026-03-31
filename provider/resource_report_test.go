package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceReport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccRequireEnv(t, "ZABBIX_TEST_REPORT_USERID", "ZABBIX_TEST_REPORT_DASHBOARDID")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceReportBasic(
					os.Getenv("ZABBIX_TEST_REPORT_USERID"),
					os.Getenv("ZABBIX_TEST_REPORT_DASHBOARDID"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_report.test", "period", "1"),
					resource.TestCheckResourceAttr("zabbix_report.test", "cycle", "1"),
				),
			},
			{
				Config: testAccResourceReportUpdated(
					os.Getenv("ZABBIX_TEST_REPORT_USERID"),
					os.Getenv("ZABBIX_TEST_REPORT_DASHBOARDID"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zabbix_report.test", "period", "2"),
					resource.TestCheckResourceAttr("zabbix_report.test", "description", "updated by acceptance test"),
				),
			},
		},
	})
}

func testAccResourceReportBasic(userID, dashboardID string) string {
	return fmt.Sprintf(`
resource "zabbix_report" "test" {
  userid      = %q
  dashboardid = %q
  name        = "tf-acc-report"

  period      = 1
  cycle       = 1
  status      = 1
  description = "created by acceptance test"

  user {
    userid = %q
  }
}
`, userID, dashboardID, userID)
}

func testAccResourceReportUpdated(userID, dashboardID string) string {
	return fmt.Sprintf(`
resource "zabbix_report" "test" {
  userid      = %q
  dashboardid = %q
  name        = "tf-acc-report"

  period      = 2
  cycle       = 2
  status      = 1
  description = "updated by acceptance test"

  user {
    userid = %q
  }
}
`, userID, dashboardID, userID)
}
