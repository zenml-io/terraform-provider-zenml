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
		"zenml": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ZENML_SERVER_URL"); v == "" {
		t.Fatal("ZENML_SERVER_URL must be set for acceptance tests")
	}
	creds := []string{"ZENML_API_KEY", "ZENML_API_TOKEN"}
	v := ""
	for _, cred := range creds {
		v = os.Getenv(cred)
		if v != "" {
			break
		}
	}
	if v == "" {
		t.Fatal(
			"ZENML_API_KEY or ZENML_API_TOKEN must be set for acceptance tests",
		)
	}
}
