// internal/provider/resource_stack_component_test.go
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccStackComponent_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackComponentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackComponentExists("zenml_stack_component.test"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "name", "test-store"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "type", "artifact_store"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "flavor", "gcp"),
				),
			},
			{
				ResourceName:      "zenml_stack_component.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStackComponent_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackComponentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackComponentExists("zenml_stack_component.test"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "name", "test-store"),
				),
			},
			{
				Config: testAccStackComponentConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackComponentExists("zenml_stack_component.test"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "name", "updated-store"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "labels.environment", "staging"),
				),
			},
		},
	})
}

func TestAccStackComponent_withConnector(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackComponentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackComponentConfig_withConnector(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackComponentExists("zenml_stack_component.test"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "name", "test-store"),
					resource.TestCheckResourceAttrPair(
						"zenml_stack_component.test", "connector",
						"zenml_service_connector.test", "id"),
				),
			},
		},
	})
}

func testAccCheckStackComponentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		client := testAccProvider.Meta().(*Client)
		_, err := client.GetComponent(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Stack Component not found: %v", err)
		}

		return nil
	}
}

func testAccCheckStackComponentDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zenml_stack_component" {
			continue
		}

		_, err := client.GetComponent(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Stack Component still exists")
		}
	}

	return nil
}

func testAccStackComponentConfig_basic() string {
	return `
resource "zenml_stack_component" "test" {
	name      = "test-store"
	type      = "artifact_store"
	flavor    = "gcp"
	user      = "user-uuid"
	workspace = "workspace-uuid"
	
	configuration = {
		path = "gs://test-bucket/artifacts"
	}
	
	labels = {
		environment = "test"
	}
}
`
}

func testAccStackComponentConfig_update() string {
	return `
resource "zenml_stack_component" "test" {
	name      = "updated-store"
	type      = "artifact_store"
	flavor    = "gcp"
	user      = "user-uuid"
	workspace = "workspace-uuid"
	
	configuration = {
		path = "gs://test-bucket/artifacts"
		location = "us-central1"
	}
	
	labels = {
		environment = "staging"
		team        = "ml"
	}
}
`
}

func testAccStackComponentConfig_withConnector() string {
	return `
resource "zenml_service_connector" "test" {
	name        = "test-connector"
	type        = "gcp"
	auth_method = "service-account"
	user        = "user-uuid"
	workspace   = "workspace-uuid"
	
	resource_types = ["artifact-store"]
	
	configuration = {
		project_id = "test-project"
	}
	
	secrets = {
		service_account_json = jsonencode({
			"type": "service_account",
			"project_id": "test-project"
		})
	}
}

resource "zenml_stack_component" "test" {
	name      = "test-store"
	type      = "artifact_store"
	flavor    = "gcp"
	user      = "user-uuid"
	workspace = "workspace-uuid"
	
	configuration = {
		path = "gs://test-bucket/artifacts"
	}
	
	connector = zenml_service_connector.test.id
	
	labels = {
		environment = "test"
	}
}
`
}