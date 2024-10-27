package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServiceConnector() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceServiceConnectorRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workspace": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}
