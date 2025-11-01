package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceStack_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceStackConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.zenml_stack.test", "name", "test-stack-data"),
					resource.TestCheckResourceAttrSet(
						"data.zenml_stack.test", "id"),
				),
			},
		},
	})
}

func testAccDataSourceStackConfig_basic() string {
	return fmt.Sprintf(`
%s

resource "zenml_stack_component" "artifact_store" {
  name   = "test-store-data"
  type   = "artifact_store"
  flavor = "local"
}

resource "zenml_stack" "test" {
  name = "test-stack-data"
  
  components = {
    "artifact_store" = zenml_stack_component.artifact_store.id
  }
}

data "zenml_stack" "test" {
  name = zenml_stack.test.name
}
`, testAccProviderConfig())
}
