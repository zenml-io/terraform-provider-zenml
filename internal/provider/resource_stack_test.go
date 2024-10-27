package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccStack_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackExists("zenml_stack.test"),
					resource.TestCheckResourceAttr(
						"zenml_stack.test", "name", "test-stack"),
				),
			},
		},
	})
}

func testAccStackConfig_basic() string {
	return `
resource "zenml_stack" "test" {
  name = "test-stack"
  components = {
    "artifact_store" = "test-store-id"
  }
  labels = {
    "environment" = "test"
  }
}
`
}