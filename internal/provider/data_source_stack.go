package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStack() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStackRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"components": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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

func dataSourceStackRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	// If ID is provided, do direct lookup
	if id, ok := d.GetOk("id"); ok {
		stack, err := client.GetStack(id.(string))
		if err != nil {
			return fmt.Errorf("error reading stack with id %s: %v", id, err)
		}
		return setStackData(d, stack)
	}

	// Check if name is provided
	name, ok := d.GetOk("name")
	if !ok {
		return fmt.Errorf("either 'id' or 'name' must be provided")
	}

	// Look up by name
	stacks, err := client.ListStacks(nil) // nil for default pagination
	if err != nil {
		return fmt.Errorf("error listing stacks: %v", err)
	}

	for _, stack := range stacks.Items {
		if stack.Name == name.(string) {
			return setStackData(d, &stack)
		}
	}

	return fmt.Errorf("no stack found with name: %s", name)
}

func setStackData(d *schema.ResourceData, stack *StackResponse) error {
	d.SetId(stack.ID)
	d.Set("name", stack.Name)

	components := make(map[string]string)
	for k, v := range stack.Components {
		components[k] = v.ID
	}
	d.Set("components", components)

	if stack.Labels != nil {
		d.Set("labels", stack.Labels)
	}

	return nil
}
