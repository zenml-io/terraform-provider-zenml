package provider

import (
	"encoding/json"
	"fmt"
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"components": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceStackCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	stack := Stack{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Components:  make(map[string]Component),
	}

	// Convert components from the schema
	componentsRaw := d.Get("components").(map[string]interface{})
	for k, v := range componentsRaw {
		var component Component
		if err := json.Unmarshal([]byte(v.(string)), &component); err != nil {
			return fmt.Errorf("error parsing component %s: %v", k, err)
		}
		stack.Components[k] = component
	}

	stackID, err := client.CreateStack(stack)
	if err != nil {
		return fmt.Errorf("error creating stack: %v", err)
	}

	d.SetId(stackID)
	return resourceStackRead(d, m)
}