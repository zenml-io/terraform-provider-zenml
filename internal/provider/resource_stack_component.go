// resource_stack_component.go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceStackComponent() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStackComponentCreate,
		ReadContext:   resourceStackComponentRead,
		UpdateContext: resourceStackComponentUpdate,
		DeleteContext: resourceStackComponentDelete,

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
			// Validate that if connector_resource_id is set, connector should also be set
			connector, hasConnector := d.GetOk("connector")
			connectorResourceID, hasConnectorResourceID := d.GetOk("connector_resource_id")

			if hasConnectorResourceID && connectorResourceID.(string) != "" && (!hasConnector || connector.(string) == "") {
				return fmt.Errorf("connector must be set when connector_resource_id is specified")
			}

			return nil
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
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
				ForceNew:     true,
			},
			"flavor": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				// We cannot delete service connectors while they are still in
				// use by a component, so we need to force new components when
				// the connector is changed.
				ForceNew: true,
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

func resourceStackComponentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client type: expected *Client"))
	}
	if client == nil {
		return diag.FromErr(fmt.Errorf("client is nil"))
	}

	// Get the current user
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting current user: %w", err))
	}

	workspaceName := d.Get("workspace").(string)

	// Get the workspace ID
	workspace, err := client.GetWorkspaceByName(ctx, workspaceName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting workspace: %w", err))
	}
	if workspace == nil {
		return diag.FromErr(fmt.Errorf("workspace not found: %s", workspaceName))
	}

	// Create the component request
	component := ComponentRequest{
		User:          user.ID, // Add the user ID
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
	resp, err := client.CreateComponent(ctx, workspace.ID, component)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok {
			return diag.FromErr(fmt.Errorf("API error: %s", apiErr.Error()))
		}
		return diag.FromErr(fmt.Errorf("failed to create component: %w", err))
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

func resourceStackComponentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, ok := m.(*Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client type: expected *Client"))
	}

	component, err := client.GetComponent(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting component: %w", err))
	}

	if component == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", component.Name)

	if component.Body != nil {
		d.Set("type", component.Body.Type)
		d.Set("flavor", component.Body.Flavor)
	}

	if component.Metadata != nil {
		d.Set("configuration", component.Metadata.Configuration)

		if component.Metadata.Workspace.Name != "default" {
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

func resourceStackComponentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	// Create update with proper string pointer for name
	name := d.Get("name").(string)
	update := ComponentUpdate{
		Name: &name,
	}

	// type and flavor are immutable, so we don't need to check for changes

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

	// The connector ID and connector resource ID fields are special: they
	// must always be set in the update request, even if they are not being
	// changed, because a missing or null value is used to clear the field.

	if v, ok := d.GetOk("connector_id"); ok {
		str := v.(string)
		update.ConnectorID = &str

		if v, ok := d.GetOk("connector_resource_id"); ok {
			str := v.(string)
			update.ConnectorResourceID = &str
		} else {
			update.ConnectorResourceID = nil
		}
	} else {
		update.ConnectorID = nil
	}

	_, err := client.UpdateComponent(ctx, d.Id(), update)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating component: %w", err))
	}

	return resourceStackComponentRead(ctx, d, m)
}

func resourceStackComponentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	err := client.DeleteComponent(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting component: %w", err))
	}

	d.SetId("")
	return nil
}

// resource_stack_component.go
