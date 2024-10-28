package provider

import (
	"testing"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
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
	return fmt.Sprintf(`
resource "zenml_stack_component" "artifact_store" {
    name      = "test-store"
    type      = "artifact_store"
    flavor    = "local"
    workspace = "%s"
    user      = "%s"
    
    configuration = {
        path = "/tmp/artifacts"
    }
}

resource "zenml_stack" "test" {
    name      = "test-stack"
    workspace = "%s"
    
    components = {
        "artifact_store" = zenml_stack_component.artifact_store.id
    }
    
    labels = {
        environment = "test"
    }
}
`, os.Getenv("ZENML_WORKSPACE"), os.Getenv("ZENML_USER_ID"), os.Getenv("ZENML_WORKSPACE"))
}

func testAccCheckStackDestroy(s *terraform.State) error {
	// Implementation needed
	return nil
}

// Add this new function
func testAccCheckStackExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Stack ID is set")
		}

		// Add actual backend verification
		client := testAccProvider.Meta().(*Client)
		_, err := client.GetStack(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error fetching stack with ID %s: %s", rs.Primary.ID, err)
		}

		return nil
	}
}
