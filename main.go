package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"server_url": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   false,
				Description: "ZenML server URL",
			},
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "ZenML API key",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"zenml_stack":            resourceStack(),
			"zenml_stack_component": resourceStackComponent(),
			"zenml_service_connector": resourceServiceConnector(),
		},
	}
}