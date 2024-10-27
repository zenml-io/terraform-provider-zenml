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

		ImportState: schema.ImportStatePassthrough,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceStackCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	stack := StackUpdate{
		Name: d.Get("name").(string),
	}

	// Handle components
	if v, ok := d.GetOk("components"); ok {
		components := make(map[string]Component)
		for k, v := range v.(map[string]interface{}) {
			components[k] = Component{
				ID: v.(string),
			}
		}
		stack.Components = components
	}

	// Handle labels
	if v, ok := d.GetOk("labels"); ok {
		labels := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			labels[k] = v.(string)
		}
		stack.Labels = labels
	}

	resp, err := client.CreateStack(stack)
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

	// Handle components
	components := make(map[string]string)
	for k, v := range stack.Components {
		components[k] = v.ID
	}
	d.Set("components", components)

	// Handle labels
	if stack.Labels != nil {
		d.Set("labels", stack.Labels)
	}

	return nil
}

func resourceStackUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	update := StackUpdate{
		Name: d.Get("name").(string),
	}

	if d.HasChange("components") {
		components := make(map[string]Component)
		for k, v := range d.Get("components").(map[string]interface{}) {
			components[k] = Component{
				ID: v.(string),
			}
		}
		update.Components = components
	}

	if d.HasChange("labels") {
		labels := make(map[string]string)
		for k, v := range d.Get("labels").(map[string]interface{}) {
			labels[k] = v.(string)
		}
		update.Labels = labels
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
