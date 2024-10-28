// resource_stack_component.go
package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceStackComponent() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackComponentCreate,
		Read:   resourceStackComponentRead,
		Update: resourceStackComponentUpdate,
		Delete: resourceStackComponentDelete,

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
			// Validate that if connector is set, connector_resource_id should also be set
			connector, hasConnector := d.GetOk("connector")
			connectorResourceID, hasConnectorResourceID := d.GetOk("connector_resource_id")

			if hasConnector && connector.(string) != "" && (!hasConnectorResourceID || connectorResourceID.(string) == "") {
				return fmt.Errorf("connector_resource_id must be set when connector is specified")
			}

			if hasConnectorResourceID && connectorResourceID.(string) != "" && (!hasConnector || connector.(string) == "") {
				return fmt.Errorf("connector must be set when connector_resource_id is specified")
			}

			return nil
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(validComponentTypes, false),
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
			},
			"configuration": {
				Type:      schema.TypeMap,
				Optional:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"connector_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"connector_resource_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceStackComponentCreate(d *schema.ResourceData, m interface{}) error {
	client, ok := m.(*Client)
	if !ok {
		return fmt.Errorf("invalid client type: expected *Client")
	}
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	// Get the current user
	user, err := client.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("error getting current user: %w", err)
	}

	workspaceName := d.Get("workspace").(string)
	
	// Get the workspace ID
	workspace, err := client.GetWorkspaceByName(workspaceName)
	if err != nil {
		return fmt.Errorf("error getting workspace: %w", err)
	}

	// Create the component request
	component := ComponentRequest{
		User:          user.ID,           // Add the user ID
		Name:          d.Get("name").(string),
		Type:          d.Get("type").(string),
		Flavor:        d.Get("flavor").(string),
		Configuration: d.Get("configuration").(map[string]interface{}),
		Workspace:     workspace.ID,
	}

	// Handle optional fields
	if v, ok := d.GetOk("connector_id"); ok {
		connectorID := v.(string)
		component.ConnectorID = &connectorID
	}

	if v, ok := d.GetOk("connector_resource_id"); ok {
		resourceID := v.(string)
		component.ConnectorResourceID = &resourceID
	}

	if v, ok := d.GetOk("labels"); ok {
		labels := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			labels[k] = v.(string)
		}
		component.Labels = labels
	}

	// Make the API call
	resp, err := client.CreateComponent(workspace.ID, component)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			return fmt.Errorf("API error: %s", apiErr.Error())
		}
		return fmt.Errorf("failed to create component: %w", err)
	}

	// Set the ID from the response
	d.SetId(resp.ID)

	// Set other attributes from the response
	d.Set("name", resp.Name)
	if resp.Body != nil {
		d.Set("type", resp.Body.Type)
		d.Set("flavor", resp.Body.Flavor)
	}
	if resp.Metadata != nil {
		d.Set("configuration", resp.Metadata.Configuration)
		if resp.Metadata.ConnectorResourceID != nil {
			d.Set("connector_resource_id", *resp.Metadata.ConnectorResourceID)
		}
		if resp.Metadata.Labels != nil {
			d.Set("labels", resp.Metadata.Labels)
		}
	}

	return nil
}

func resourceStackComponentRead(d *schema.ResourceData, m interface{}) error {
	client, ok := m.(*Client)
	if !ok {
		return fmt.Errorf("invalid client type: expected *Client")
	}

	component, err := client.GetComponent(d.Id())
	if err != nil {
		return fmt.Errorf("error getting component: %w", err)
	}

	d.Set("name", component.Name)
	
	if component.Body != nil {
		d.Set("type", component.Body.Type)
		d.Set("flavor", component.Body.Flavor)
	}
	
	if component.Metadata != nil {
		d.Set("configuration", component.Metadata.Configuration)
		if component.Metadata.Workspace != nil {
			d.Set("workspace", component.Metadata.Workspace.Name)
		}
		if component.Metadata.ConnectorResourceID != nil {
			d.Set("connector_resource_id", *component.Metadata.ConnectorResourceID)
		}
		if component.Metadata.Labels != nil {
			d.Set("labels", component.Metadata.Labels)
		}
	}

	return nil
}

func resourceStackComponentUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	// Create update with proper string pointer for name
	name := d.Get("name").(string)
	update := ComponentUpdate{
		Name: &name,
	}

	if d.HasChange("configuration") {
		configMap := make(map[string]interface{})
		for k, v := range d.Get("configuration").(map[string]interface{}) {
			configMap[k] = v
		}
		update.Configuration = configMap
	}

	if d.HasChange("labels") {
		labelsMap := make(map[string]string)
		for k, v := range d.Get("labels").(map[string]interface{}) {
			labelsMap[k] = v.(string)
		}
		update.Labels = labelsMap
	}

	if d.HasChange("component_spec_path") {
		if v, ok := d.GetOk("component_spec_path"); ok {
			str := v.(string)
			update.ComponentSpecPath = &str
		}
	}

	if d.HasChange("connector_id") {
		if v, ok := d.GetOk("connector_id"); ok {
			str := v.(string)
			update.ConnectorID = &str
		}
	}

	_, err := client.UpdateComponent(d.Id(), update)
	if err != nil {
		return err
	}

	return resourceStackComponentRead(d, m)
}

func resourceStackComponentDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	err := client.DeleteComponent(d.Id())
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// resource_stack_component.go
