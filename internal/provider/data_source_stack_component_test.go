package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestDataSourceStackComponent_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprint(`
					data "zenml_stack_component" "orchestrator" {
						name        = "default"
						type        = "orchestrator"
					}
				`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.zenml_stack_component.orchestrator", "name", "default"),
					resource.TestCheckResourceAttr(
						"data.zenml_stack_component.orchestrator", "type", "orchestrator"),
				),
			},
		},
	})
}
