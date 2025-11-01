package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStackComponent_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackComponentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "name", "test-store"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "type", "artifact_store"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "flavor", "local"),
					resource.TestCheckResourceAttrSet(
						"zenml_stack_component.test", "id"),
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackComponentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "name", "test-store"),
				),
			},
			{
				Config: testAccStackComponentConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "name", "updated-store"),
				),
			},
		},
	})
}

func TestAccStackComponent_withLabels(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackComponentConfig_withLabels(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "name", "test-store-labeled"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "labels.environment", "test"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "labels.team", "platform"),
				),
			},
		},
	})
}

func TestAccStackComponent_withConfiguration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackComponentConfig_withConfiguration(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "name", "test-store-config"),
					resource.TestCheckResourceAttr(
						"zenml_stack_component.test", "configuration.path", "/tmp/zenml"),
				),
			},
		},
	})
}

func testAccStackComponentConfig_basic() string {
	return fmt.Sprintf(`
%s

resource "zenml_stack_component" "test" {
  name   = "test-store"
  type   = "artifact_store"
  flavor = "local"
}
`, testAccProviderConfig())
}

func testAccStackComponentConfig_updated() string {
	return fmt.Sprintf(`
%s

resource "zenml_stack_component" "test" {
  name   = "updated-store"
  type   = "artifact_store"
  flavor = "local"
}
`, testAccProviderConfig())
}

func testAccStackComponentConfig_withLabels() string {
	return fmt.Sprintf(`
%s

resource "zenml_stack_component" "test" {
  name   = "test-store-labeled"
  type   = "artifact_store"
  flavor = "local"
  
  labels = {
    environment = "test"
    team        = "platform"
  }
}
`, testAccProviderConfig())
}

func testAccStackComponentConfig_withConfiguration() string {
	return fmt.Sprintf(`
%s

resource "zenml_stack_component" "test" {
  name   = "test-store-config"
  type   = "artifact_store"
  flavor = "local"
  
  configuration = {
    path = "/tmp/zenml"
  }
}
`, testAccProviderConfig())
}
