// internal/provider/resource_service_connector_test.go
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccServiceConnector_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConnectorConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceConnectorExists("zenml_service_connector.test"),
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "name", "test-connector"),
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "type", "gcp"),
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "auth_method", "service-account"),
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "configuration.project_id", "test-project"),
				),
			},
			{
				// Test importing the resource
				ResourceName:      "zenml_service_connector.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Don't verify sensitive fields
				ImportStateVerifyIgnore: []string{"secrets"},
			},
		},
	})
}

func TestAccServiceConnector_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceConnectorConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceConnectorExists("zenml_service_connector.test"),
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "name", "test-connector"),
				),
			},
			{
				Config: testAccServiceConnectorConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceConnectorExists("zenml_service_connector.test"),
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "name", "updated-connector"),
					resource.TestCheckResourceAttr(
						"zenml_service_connector.test", "labels.environment", "staging"),
				),
			},
		},
	})
}

func testAccCheckServiceConnectorExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		client := testAccProvider.Meta().(*Client)
		_, err := client.GetServiceConnector(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Service Connector not found: %v", err)
		}

		return nil
	}
}

func testAccCheckServiceConnectorDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zenml_service_connector" {
			continue
		}

		_, err := client.GetServiceConnector(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Service Connector still exists")
		}
	}

	return nil
}

func testAccServiceConnectorConfig_basic() string {
	return `
resource "zenml_service_connector" "test" {
	name        = "test-connector"
	type        = "gcp"
	auth_method = "service-account"
	user        = "user-uuid"
	workspace   = "workspace-uuid"
	
	resource_types = [
		"artifact-store",
		"container-registry"
	]
	
	configuration = {
		project_id = "test-project"
	}
	
	secrets = {
		service_account_json = jsonencode({
			"type": "service_account",
			"project_id": "test-project"
		})
	}
	
	labels = {
		environment = "test"
	}
}
`
}

func testAccServiceConnectorConfig_update() string {
	return `
resource "zenml_service_connector" "test" {
	name        = "updated-connector"
	type        = "gcp"
	auth_method = "service-account"
	user        = "user-uuid"
	workspace   = "workspace-uuid"
	
	resource_types = [
		"artifact-store",
		"container-registry",
		"orchestrator"
	]
	
	configuration = {
		project_id = "test-project"
		region     = "us-central1"
	}
	
	secrets = {
		service_account_json = jsonencode({
			"type": "service_account",
			"project_id": "test-project"
		})
	}
	
	labels = {
		environment = "staging"
		team        = "ml"
	}
}
`
}
