package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStackComponent() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackComponentCreate,
		Read:   resourceStackComponentRead,
		Update: resourceStackComponentUpdate,
		Delete: resourceStackComponentDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
			},
			"configuration": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"service_connector_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}