package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStack() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStackRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
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
			return err
		}
		return setStackData(d, stack)
	}

	// Otherwise, look up by name
	name := d.Get("name").(string)
	stacks, err := client.ListStacks()
	if err != nil {
		return err
	}

	for _, stack := range stacks.Items {
		if stack.Name == name {
			return setStackData(d, &stack)
		}
	}

	return fmt.Errorf("no stack found with name: %s", name)
}