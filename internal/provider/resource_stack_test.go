package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	var stackID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack.test", "name", "test-stack"),
					testAccCaptureResourceAttr("zenml_stack.test", "id", &stackID),
				),
			},
			{
				Config: testAccStackConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"zenml_stack.test", "name", "updated-stack"),
					resource.TestCheckResourceAttr(
						"zenml_stack.test", "labels.environment", "production"),
					testAccCheckResourceAttrEquals("zenml_stack.test", "id", &stackID),
				),
			},
		},
	})
}

func TestAccStack_noopReapplyDoesNotReplace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(),
			},
			{
				Config:             testAccStackConfig_basic(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccStack_updateRequiredComponentReplacesStack(t *testing.T) {
	var stackID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_withTwoOrchestrators(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zenml_stack.test", "name", "test-stack"),
					testAccCaptureResourceAttr("zenml_stack.test", "id", &stackID),
					resource.TestCheckResourceAttrPair(
						"zenml_stack.test", "components.orchestrator",
						"zenml_stack_component.orchestrator_a", "id",
					),
				),
			},
			{
				Config: testAccStackConfig_withTwoOrchestrators(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrChanged("zenml_stack.test", "id", &stackID),
					resource.TestCheckResourceAttrPair(
						"zenml_stack.test", "components.orchestrator",
						"zenml_stack_component.orchestrator_b", "id",
					),
				),
			},
		},
	})
}

func TestAccStack_updateOptionalComponentInPlace(t *testing.T) {
	var stackID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_withOptionalComponent(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zenml_stack.test", "name", "test-stack"),
					testAccCaptureResourceAttr("zenml_stack.test", "id", &stackID),
				),
			},
			{
				Config: testAccStackConfig_withOptionalComponent(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrEquals("zenml_stack.test", "id", &stackID),
				),
			},
		},
	})
}

func testAccCaptureResourceAttr(resourceName, attr string, dest *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		value, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("attribute %s missing on resource %s", attr, resourceName)
		}

		*dest = value
		return nil
	}
}

func testAccCheckResourceAttrEquals(resourceName, attr string, expected *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		value, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("attribute %s missing on resource %s", attr, resourceName)
		}

		if value != *expected {
			return fmt.Errorf("attribute %s on %s mismatch: got %q want %q", attr, resourceName, value, *expected)
		}

		return nil
	}
}

func testAccCheckResourceAttrChanged(resourceName, attr string, previous *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}

		value, ok := rs.Primary.Attributes[attr]
		if !ok {
			return fmt.Errorf("attribute %s missing on resource %s", attr, resourceName)
		}

		if value == *previous {
			return fmt.Errorf(
				"attribute %s on %s should have changed but still equals %q",
				attr, resourceName, *previous,
			)
		}

		*previous = value
		return nil
	}
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

func testAccStackConfig_withTwoOrchestrators(useSecond bool) string {
	orchestratorRef := "zenml_stack_component.orchestrator_a.id"
	if useSecond {
		orchestratorRef = "zenml_stack_component.orchestrator_b.id"
	}

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

resource "zenml_stack_component" "orchestrator_a" {
  name   = "test-orchestrator-a"
  type   = "orchestrator"
  flavor = "local"
}

resource "zenml_stack_component" "orchestrator_b" {
  name   = "test-orchestrator-b"
  type   = "orchestrator"
  flavor = "local"
}

resource "zenml_stack" "test" {
  name = "test-stack"

  components = {
    artifact_store = zenml_stack_component.artifact_store.id
    orchestrator   = %s
  }

  labels = {
    environment = "test"
  }
}
`, testAccProviderConfig(), orchestratorRef)
}

func testAccStackConfig_withOptionalComponent(useSecond bool) string {
	imageBuilderRef := "zenml_stack_component.image_builder_a.id"
	if useSecond {
		imageBuilderRef = "zenml_stack_component.image_builder_b.id"
	}

	return fmt.Sprintf(`
%s

resource "zenml_stack_component" "artifact_store" {
  name   = "test-store-opt"
  type   = "artifact_store"
  flavor = "local"

  configuration = {
    path = "/tmp/artifacts"
  }
}

resource "zenml_stack_component" "orchestrator" {
  name   = "test-orchestrator-opt"
  type   = "orchestrator"
  flavor = "local"
}

resource "zenml_stack_component" "image_builder_a" {
  name   = "test-image-builder-a"
  type   = "image_builder"
  flavor = "local"
}

resource "zenml_stack_component" "image_builder_b" {
  name   = "test-image-builder-b"
  type   = "image_builder"
  flavor = "local"
}

resource "zenml_stack" "test" {
  name = "test-stack"

  components = {
    artifact_store = zenml_stack_component.artifact_store.id
    orchestrator   = zenml_stack_component.orchestrator.id
    image_builder  = %s
  }

  labels = {
    environment = "test"
  }
}
`, testAccProviderConfig(), imageBuilderRef)
}
