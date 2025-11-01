package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceConnector_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConnectorConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "name", "test-connector"),
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "type", "aws"),
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "auth_method", "secret-key"),
					resource.TestCheckResourceAttrSet(
						"zenml_service_connector.test", "id"),
				),
			},
			{
				ResourceName:            "zenml_service_connector.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secrets"},
			},
		},
	})
}

func testAccServiceConnectorConfig_basic() string {
	return fmt.Sprintf(`
%s

resource "zenml_service_connector" "test" {
  name        = "test-connector"
  type        = "aws"
  auth_method = "secret-key"
  
  configuration = {
    region = "us-east-1"
  }
  
  secrets = {
    aws_access_key_id     = "test-key"
    aws_secret_access_key = "test-secret"
  }
  
  labels = {
    environment = "test"
  }
}
`, testAccProviderConfig())
}
