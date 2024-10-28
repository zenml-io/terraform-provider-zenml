package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func dataSourceStackComponentRead(d *schema.ResourceData, m interface{}) error {
	client, ok := m.(*Client)
	if !ok {
		return fmt.Errorf("invalid client type: expected *Client")
	}

	// Get the component by ID
	component, err := client.GetComponent(d.Get("id").(string))
	if err != nil {
		return fmt.Errorf("error getting component: %w", err)
	}

	if component.Body == nil {
		return fmt.Errorf("received empty response body")
	}

	// Always access fields through component.Body
	d.SetId(component.ID)
	d.Set("name", component.Body.Name)
	d.Set("type", component.Body.Type)
	d.Set("flavor", component.Body.Flavor)
	d.Set("configuration", component.Body.Configuration)
	d.Set("workspace", component.Body.Workspace)
	d.Set("user", component.Body.User)

	if component.Body.ConnectorResourceID != "" {
		d.Set("connector_resource_id", component.Body.ConnectorResourceID)
	}
	
	if component.Body.Labels != nil {
		d.Set("labels", component.Body.Labels)
	}

	return nil
}

func setStackComponentFields(d *schema.ResourceData, component *ComponentResponse) error {
	if err := d.Set("name", component.Name); err != nil {
		return fmt.Errorf("error setting name: %v", err)
	}

	if component.Metadata != nil {
		if err := d.Set("configuration", component.Metadata.Configuration); err != nil {
			return fmt.Errorf("error setting configuration: %v", err)
		}

		if component.Metadata.Workspace != nil {
			if err := d.Set("workspace", component.Metadata.Workspace.Name); err != nil {
				return fmt.Errorf("error setting workspace: %v", err)
			}
		}

		if component.Metadata.ConnectorResourceID != nil {
			if err := d.Set("connector_resource_id", component.Metadata.ConnectorResourceID); err != nil {
				return fmt.Errorf("error setting connector_resource_id: %v", err)
			}
		}

		if err := d.Set("labels", component.Metadata.Labels); err != nil {
			return fmt.Errorf("error setting labels: %v", err)
		}
	}

	return nil
}

func dataSourceStackComponent() *schema.Resource {
	return &schema.Resource{
		ReadContext: schema.ReadContextFunc(func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
			if err := dataSourceStackComponentRead(d, m); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}),

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration": {
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"workspace": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connector_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
