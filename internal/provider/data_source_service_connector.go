package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServiceConnector() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for ZenML service connectors",
		ReadContext: dataSourceServiceConnectorRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the service connector",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "Name of the service connector",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description: "Type of the service connector",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"auth_method": {
				Description: "Authentication method of the service connector",
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
				Sensitive: true,
			},
			"labels": {
				Description: "Labels associated with the service connector",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"resource_type": {
				Description: "Resource type associated with the service connector",
				Type:        schema.TypeString,
				Computed:    true,
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
			"created": {
				Description: "Timestamp when the service connector was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated": {
				Description: "Timestamp when the service connector was last updated",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceServiceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	name := d.Get("name").(string)
	id := d.Get("id").(string)

	var err error = nil
	var connector *ServiceConnectorResponse = nil

	if id != "" {
		connector, err = c.GetServiceConnector(ctx, id)
	} else if name != "" {
		connector, err = c.GetServiceConnectorByName(ctx, name)
	} else {
		return diag.FromErr(fmt.Errorf("either 'id' or 'name' must be set"))
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting service connector: %v", err))
	}
	if connector == nil {
		// Connector not found
		d.SetId("")
		return nil
	}

	d.SetId(connector.ID)

	if err := d.Set("name", connector.Name); err != nil {
		return diag.FromErr(err)
	}

	if connector.Body != nil {

		connector_type := ""

		// Unmarshal the connector type, which can be either a string or a struct
		// Try to unmarshal as string
		err = json.Unmarshal(connector.Body.ConnectorType, &connector_type)
		if err != nil {
			var type_struct ServiceConnectorType
			// Try to unmarshal as struct
			if err = json.Unmarshal(connector.Body.ConnectorType, &type_struct); err == nil {
				connector_type = type_struct.ConnectorType
			}
		}

		if err := d.Set("type", connector_type); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("auth_method", connector.Body.AuthMethod); err != nil {
			return diag.FromErr(err)
		}

		// If there are multiple resource types, leave the resource_type field empty
		if len(connector.Body.ResourceTypes) == 1 {
			d.Set("resource_type", connector.Body.ResourceTypes[0])
		} else {
			d.Set("resource_type", "")
		}

		if connector.Body.ResourceID != nil {
			if err := d.Set("resource_id", *connector.Body.ResourceID); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := d.Set("resource_id", ""); err != nil {
				return diag.FromErr(err)
			}
		}

		if connector.Body.ExpiresAt != nil {
			if err := d.Set("expires_at", *connector.Body.ExpiresAt); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := d.Set("expires_at", ""); err != nil {
				return diag.FromErr(err)
			}
		}

		if err := d.Set("created", connector.Body.Created); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("updated", connector.Body.Updated); err != nil {
			return diag.FromErr(err)
		}
	}

	if connector.Metadata != nil {
		if err := d.Set("configuration", connector.Metadata.Configuration); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("labels", connector.Metadata.Labels); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
