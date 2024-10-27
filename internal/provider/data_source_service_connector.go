package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServiceConnectorRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	// Try to find by ID first
	if id, ok := d.GetOk("id"); ok {
		connector, err := client.GetServiceConnector(id.(string))
		if err != nil {
			return fmt.Errorf("error reading service connector: %v", err)
		}
		if connector == nil {
			return fmt.Errorf("service connector with id %s not found", id)
		}
		d.SetId(connector.ID)
		return setServiceConnectorFields(d, connector)
	}

	// Try to find by name and workspace
	name, hasName := d.GetOk("name")
	workspace, hasWorkspace := d.GetOk("workspace")

	if !hasName {
		return fmt.Errorf("either id or name must be specified")
	}

	var workspaceStr string
	if hasWorkspace {
		workspaceStr = workspace.(string)
	}

	connector, err = client.GetServiceConnectorByName(name.(string), workspaceStr)
	if err != nil {
		return fmt.Errorf("error reading service connector: %v", err)
	}
	if connector == nil {
		return fmt.Errorf("service connector with name %s not found", name)
	}

	d.SetId(connector.ID)
	return setServiceConnectorFields(d, connector)
}

func setServiceConnectorFields(d *schema.ResourceData, connector *ServiceConnectorResponse) error {
	if connector == nil {
		return fmt.Errorf("cannot set fields from nil connector")
	}
	
	if err := d.Set("name", connector.Name); err != nil {
		return fmt.Errorf("error setting name: %v", err)
	}
	if err := d.Set("type", connector.Type); err != nil {
		return fmt.Errorf("error setting type: %v", err)
	}
	if err := d.Set("auth_method", connector.AuthMethod); err != nil {
		return fmt.Errorf("error setting auth_method: %v", err)
	}

	if connector.Body != nil {
		if connector.Body.ResourceTypes != nil {
			if err := d.Set("resource_types", connector.Body.ResourceTypes); err != nil {
				return fmt.Errorf("error setting resource_types: %v", err)
			}
		}
		if connector.Body.Workspace != "" {
			if err := d.Set("workspace", connector.Body.Workspace); err != nil {
				return fmt.Errorf("error setting workspace: %v", err)
			}
		}
	}

	return nil
}

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
