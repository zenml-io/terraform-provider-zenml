package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServiceConnector() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for ZenML service connectors",
		ReadContext: dataSourceServiceConnectorRead,
		Schema: map[string]*schema.Schema{
			"workspace": {
				Description: "Name of the workspace (defaults to 'default')",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
			},
			"name": {
				Description: "Name of the service connector",
				Type:        schema.TypeString,
				Required:    true,
			},
			"connector_type": {
				Description: "Type of the service connector",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"configuration": {
				Description: "Configuration of the service connector",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"labels": {
				Description: "Labels associated with the service connector",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"resource_id": {
				Description: "Resource ID associated with the service connector",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "Expiration timestamp of the service connector",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceServiceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	connector, err := c.GetServiceConnectorByName(workspace, name)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting service connector: %v", err))
	}

	d.SetId(connector.ID)
	
	if err := d.Set("connector_type", connector.Body.ConnectorType); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("configuration", connector.Metadata.Configuration); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("labels", connector.Metadata.Labels); err != nil {
		return diag.FromErr(err)
	}

	if connector.Body.ResourceID != nil {
		if err := d.Set("resource_id", *connector.Body.ResourceID); err != nil {
			return diag.FromErr(err)
		}
	}

	if connector.Body.ExpiresAt != nil {
		if err := d.Set("expires_at", *connector.Body.ExpiresAt); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
