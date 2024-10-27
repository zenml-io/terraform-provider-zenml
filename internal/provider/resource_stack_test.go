package provider

import (
	"testing"
	"fmt"
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

func testAccCheckStackDestroy(s *terraform.State) error {
	// Implementation needed
	return nil
}

// Add this new function
func testAccCheckStackExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Retrieve the resource by name from state
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Stack ID is set")
		}

		// Add code here to check if the stack exists in your system
		// This typically involves making an API call to your backend

		// For example:
		// client := testAccProvider.Meta().(*YourClientType)
		// _, err := client.GetStack(rs.Primary.ID)
		// if err != nil {
		// 	return fmt.Errorf("Error retrieving stack: %s", err)
		// }

		return nil
	}
}
