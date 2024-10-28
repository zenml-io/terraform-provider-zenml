// resource_stack.go
package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackCreate,
		Read:   resourceStackRead,
		Update: resourceStackUpdate,
		Delete: resourceStackDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"components": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Map of component types to component IDs",
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceStackCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	// Get the workspace from configuration or use a default
	workspace := "default" // You may want to make this configurable

	stack := StackRequest{
		Name: d.Get("name").(string),
	}

	// Handle components
	if v, ok := d.GetOk("components"); ok {
		components := make(map[string][]string)
		for k, v := range v.(map[string]interface{}) {
			// Convert single ID to array of IDs since API expects array
			components[k] = []string{v.(string)}
		}
		stack.Components = components
	}

	resp, err := client.CreateStack(workspace, stack)
	if err != nil {
		return err
	}

	d.SetId(resp.ID)
	return resourceStackRead(d, m)
}

func resourceStackRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	stack, err := client.GetStack(d.Id())
	if err != nil {
		// Handle 404 by removing from state
		d.SetId("")
		return nil
	}

	d.Set("name", stack.Name)

	// Handle components - flatten the array structure to single IDs
	if stack.Metadata != nil && stack.Metadata.Components != nil {
		components := make(map[string]string)
		for compType, compArray := range stack.Metadata.Components {
			if len(compArray) > 0 {
				// Take first component ID for each type
				components[compType] = compArray[0].ID
			}
		}
		d.Set("components", components)
	}

	// Handle labels if present
	if stack.Metadata != nil && stack.Metadata.Labels != nil {
		d.Set("labels", stack.Metadata.Labels)
	}

	return nil
}

func resourceStackUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	name := d.Get("name").(string)
	update := StackUpdate{
		Name: &name,
	}

	// Handle components
	if d.HasChange("components") {
		components := make(map[string][]string)
		for k, v := range d.Get("components").(map[string]interface{}) {
			// Convert single ID to array of IDs
			components[k] = []string{v.(string)}
		}
		update.Components = components
	}

	// Handle labels
	if d.HasChange("labels") {
		if v, ok := d.GetOk("labels"); ok {
			labels := make(map[string]string)
			for k, v := range v.(map[string]interface{}) {
				labels[k] = v.(string)
			}
			update.Labels = labels
		}
	}

	_, err := client.UpdateStack(d.Id(), update)
	if err != nil {
		return err
	}

	return resourceStackRead(d, m)
}

func resourceStackDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	err := client.DeleteStack(d.Id())
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
