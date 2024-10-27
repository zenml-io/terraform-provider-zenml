package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStackComponentRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	// Try to find by ID first
	if id, ok := d.GetOk("id"); ok {
		component, err := client.GetComponent(id.(string))
		if err != nil {
			return fmt.Errorf("error reading stack component: %v", err)
		}
		d.SetId(component.ID)
		return setStackComponentFields(d, component)
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

	component, err := client.GetComponentByName(name.(string), workspaceStr)
	if err != nil {
		return fmt.Errorf("error reading stack component: %v", err)
	}

	d.SetId(component.ID)
	return setStackComponentFields(d, component)
}

func setStackComponentFields(d *schema.ResourceData, component *ComponentResponse) error {
	d.Set("name", component.Name)
	d.Set("type", component.Type)
	d.Set("flavor", component.Flavor)

	if component.Body != nil {
		d.Set("configuration", component.Body.Configuration)
		if component.Body.Workspace != "" {
			d.Set("workspace", component.Body.Workspace)
		}
	}

	return nil
}

func dataSourceStackComponent() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStackComponentRead,

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
				Optional: true,
			},
		},
	}
}
