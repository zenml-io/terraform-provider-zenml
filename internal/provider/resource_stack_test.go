package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStack_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack.test", "name", "test-stack"),
					resource.TestCheckResourceAttrSet(
						"zenml_stack.test", "id"),
					resource.TestCheckResourceAttr(
						"zenml_stack.test", "labels.environment", "test"),
				),
			},
			{
				ResourceName:      "zenml_stack.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStack_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack.test", "name", "test-stack"),
				),
			},
			{
				Config: testAccStackConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack.test", "name", "updated-stack"),
					resource.TestCheckResourceAttr(
						"zenml_stack.test", "labels.environment", "production"),
				),
			},
		},
	})
}

func testAccStackConfig_basic() string {
	return fmt.Sprintf(`
%s

resource "zenml_stack_component" "artifact_store" {
  name   = "test-store"
  type   = "artifact_store"
  flavor = "local"
  
  configuration = {
    path = "/tmp/artifacts"
  }
}

resource "zenml_stack_component" "orchestrator" {
  name   = "test-orchestrator"
  type   = "orchestrator"
  flavor = "local"
}

resource "zenml_stack" "test" {
  name = "test-stack"
  
  components = {
    "artifact_store" = zenml_stack_component.artifact_store.id
    "orchestrator"   = zenml_stack_component.orchestrator.id
  }
  
  labels = {
    environment = "test"
  }
}
`, testAccProviderConfig())
}

func testAccStackConfig_updated() string {
	return fmt.Sprintf(`
%s

resource "zenml_stack_component" "artifact_store" {
  name   = "test-store"
  type   = "artifact_store"
  flavor = "local"
  
  configuration = {
    path = "/tmp/artifacts"
  }
}

resource "zenml_stack_component" "orchestrator" {
  name   = "test-orchestrator"
  type   = "orchestrator"
  flavor = "local"
}

resource "zenml_stack" "test" {
  name = "updated-stack"
  
  components = {
    "artifact_store" = zenml_stack_component.artifact_store.id
    "orchestrator"   = zenml_stack_component.orchestrator.id
  }
  
  labels = {
    environment = "production"
    team        = "platform"
  }
}
`, testAccProviderConfig())
}
