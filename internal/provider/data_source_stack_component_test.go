package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceStackComponent_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceStackComponentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.zenml_stack_component.test", "name", "test-component-data"),
					resource.TestCheckResourceAttr(
						"data.zenml_stack_component.test", "type", "artifact_store"),
					resource.TestCheckResourceAttr(
						"data.zenml_stack_component.test", "flavor", "local"),
					resource.TestCheckResourceAttrSet(
						"data.zenml_stack_component.test", "id"),
				),
			},
		},
	})
}

func testAccDataSourceStackComponentConfig_basic() string {
	return fmt.Sprintf(`
%s

resource "zenml_stack_component" "test" {
  name   = "test-component-data"
  type   = "artifact_store"
  flavor = "local"
}

data "zenml_stack_component" "test" {
  name = zenml_stack_component.test.name
  type = zenml_stack_component.test.type
}
`, testAccProviderConfig())
}
