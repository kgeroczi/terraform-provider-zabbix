package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"zabbix": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	t.Helper()

	if !testAccAnyEnvSet("ZABBIX_URL", "ZABBIX_SERVER_URL") {
		t.Fatalf("one of environment variables must be set: ZABBIX_URL, ZABBIX_SERVER_URL")
	}

	hasTokenAuth := testAccAnyEnvSet("ZABBIX_TOKEN", "ZABBIX_API_TOKEN")
	hasUserPassAuth := testAccAnyEnvSet("ZABBIX_USER", "ZABBIX_USERNAME") && testAccAnyEnvSet("ZABBIX_PASS", "ZABBIX_PASSWORD")

	if !hasTokenAuth && !hasUserPassAuth {
		t.Fatalf("set token auth (ZABBIX_TOKEN or ZABBIX_API_TOKEN) or user/pass auth (ZABBIX_USER or ZABBIX_USERNAME, and ZABBIX_PASS or ZABBIX_PASSWORD)")
	}
}

func testAccAnyEnvSet(envNames ...string) bool {
	for _, envName := range envNames {
		if os.Getenv(envName) != "" {
			return true
		}
	}

	return false
}

func testAccRequireEnv(t *testing.T, envNames ...string) {
	t.Helper()

	for _, envName := range envNames {
		if os.Getenv(envName) == "" {
			t.Skipf("skipping acceptance test; missing environment variable %s", envName)
		}
	}
}
