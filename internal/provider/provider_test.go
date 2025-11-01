package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"zenml": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProvider(t *testing.T) {
	ctx := context.Background()
	p := New("test")()

	// Test that the provider schema can be retrieved
	schemaResp := &provider.SchemaResponse{}
	p.Schema(ctx, provider.SchemaRequest{}, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Provider schema has errors: %v", schemaResp.Diagnostics)
	}

	// Test that the provider metadata can be retrieved
	metadataResp := &provider.MetadataResponse{}
	p.Metadata(ctx, provider.MetadataRequest{}, metadataResp)

	if metadataResp.TypeName != "zenml" {
		t.Fatalf("Expected provider type name 'zenml', got: %s", metadataResp.TypeName)
	}
}

func testAccPreCheck(t *testing.T) {
	// Check for required environment variables
	if v := os.Getenv("ZENML_SERVER_URL"); v == "" {
		t.Fatal("ZENML_SERVER_URL must be set for acceptance tests")
	}

	// Check for authentication credentials
	creds := []string{"ZENML_API_KEY", "ZENML_API_TOKEN"}
	hasAuth := false
	for _, cred := range creds {
		if v := os.Getenv(cred); v != "" {
			hasAuth = true
			break
		}
	}
	if !hasAuth {
		t.Fatal("ZENML_API_KEY or ZENML_API_TOKEN must be set for acceptance tests")
	}
}

// Example acceptance test function
func TestAccProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(),
				Check:  resource.ComposeTestCheckFunc(
				// Add basic provider checks here
				),
			},
		},
	})
}

func testAccProviderConfig() string {
	return `
provider "zenml" {
  server_url = "` + os.Getenv("ZENML_SERVER_URL") + `"
  api_key    = "` + os.Getenv("ZENML_API_KEY") + `"
}
`
}
