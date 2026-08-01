package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

/*
TestAccSecret_basic verifies the complete Terraform resource lifecycle:
1. Create a public secret with two values and a server-generated ID.
2. Rename the secret and change it from public to private.
3. Update one value and remove another through full-replacement semantics.
4. Import the secret by UUID and verify that the imported state matches.
5. Delete the secret during Terraform's automatic test cleanup.

It requires ZENML_SERVER_URL and either ZENML_API_KEY or ZENML_API_TOKEN for
an active ZenML server account that can manage secrets.

Run with:

export ZENML_SERVER_URL=""
export ZENML_API_KEY=""
TF_ACC=1 go test ./internal/provider -run '^TestAccSecret_basic$' -v -count=1 -timeout 20m
*/

func TestAccSecret_basic(t *testing.T) {
	name := "terraform-test-" + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccSecretPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretConfig(name, false, map[string]string{
					"client_id":     "test-client",
					"client_secret": "test-secret",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zenml_secret.test", "name", name),
					resource.TestCheckResourceAttr("zenml_secret.test", "private", "false"),
					resource.TestCheckResourceAttr("zenml_secret.test", "values.%", "2"),
					resource.TestCheckResourceAttr("zenml_secret.test", "values.client_secret", "test-secret"),
					resource.TestCheckResourceAttrSet("zenml_secret.test", "id"),
				),
			},
			{
				Config: testAccSecretConfig(name+"-updated", true, map[string]string{
					"client_id": "updated-client",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("zenml_secret.test", "name", name+"-updated"),
					resource.TestCheckResourceAttr("zenml_secret.test", "private", "true"),
					resource.TestCheckResourceAttr("zenml_secret.test", "values.%", "1"),
					resource.TestCheckResourceAttr("zenml_secret.test", "values.client_id", "updated-client"),
					resource.TestCheckNoResourceAttr("zenml_secret.test", "values.client_secret"),
				),
			},
			{
				ResourceName:      "zenml_secret.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestSecretResourceRejectsRedactedValues(t *testing.T) {
	var diagnostics diag.Diagnostics
	resource := SecretResource{}
	resource.populateSecretModel(
		context.Background(),
		&SecretResponse{
			ID:   "secret-id",
			Name: "secret-name",
			Body: &SecretResponseBody{
				Values: map[string]*string{"password": nil},
			},
		},
		&SecretResourceModel{},
		&diagnostics,
	)

	if !diagnostics.HasError() {
		t.Fatal("expected redacted secret values to produce an error")
	}
}

func testAccSecretPreCheck(t *testing.T) {
	testAccPreCheck(t)

	ctx := context.Background()
	client := NewClient(
		os.Getenv("ZENML_SERVER_URL"),
		os.Getenv("ZENML_API_KEY"),
		os.Getenv("ZENML_API_TOKEN"),
	)
	value := "permission-check"
	secret, err := client.CreateSecret(ctx, SecretRequest{
		Name:   "terraform-permission-check-" + acctest.RandString(8),
		Values: map[string]string{"value": value},
	})
	if err != nil {
		t.Fatalf("acceptance tests require permission to create secrets: %s", err)
	}
	defer func() {
		if err := client.DeleteSecret(ctx, secret.ID); err != nil {
			t.Errorf("unable to delete secret permission check: %s", err)
		}
	}()

	secret, err = client.GetSecret(ctx, secret.ID)
	if err != nil {
		t.Fatalf("acceptance tests require permission to read secrets: %s", err)
	}
	if secret == nil || secret.Body == nil || secret.Body.Values["value"] == nil {
		t.Fatal("acceptance tests require an account that can read secret values")
	}
}

func testAccSecretConfig(name string, private bool, values map[string]string) string {
	return fmt.Sprintf(`
%s

resource "zenml_secret" "test" {
  name    = %q
  private = %t
  values  = %s
}
`, testAccProviderConfig(), name, private, terraformStringMap(values))
}

func terraformStringMap(values map[string]string) string {
	result := "{\n"
	for key, value := range values {
		result += fmt.Sprintf("    %s = %q\n", key, value)
	}
	return result + "  }"
}
