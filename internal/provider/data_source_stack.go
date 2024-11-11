package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStack() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for ZenML stacks",
		ReadContext: dataSourceStackRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the stack",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"workspace": {
				Description: "Name of the workspace (defaults to 'default')",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
			},
			"name": {
				Description: "Name of the stack",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"components": {
				Description: "Components configured in the stack",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
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
					},
				},
			},
			"labels": {
				Description: "Labels associated with the stack",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Description: "Timestamp when the stack was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated": {
				Description: "Timestamp when the stack was last updated",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceStackRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	id := d.Get("id").(string)
	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	var stack *StackResponse
	var err error

	if id != "" {
		// Get stack by ID
		stack, err = c.GetStack(ctx, id)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error getting stack: %v", err))
		}
	} else if name != "" {
		// List stacks with filter to find by name
		params := &ListParams{
			Filter: map[string]string{
				"name":      name,
				"workspace": workspace,
			},
		}

		stacks, err := c.ListStacks(ctx, params)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing stacks: %v", err))
		}

		if len(stacks.Items) == 0 {
			return diag.FromErr(fmt.Errorf("no stack found with name %s in workspace %s", name, workspace))
		}

		stack = &stacks.Items[0]
	} else {
		return diag.FromErr(fmt.Errorf("either 'id' or 'name' must be set"))
	}

	if stack == nil {
		// Stack not found
		d.SetId("")
		return nil
	}

	d.SetId(stack.ID)

	if err := d.Set("name", stack.Name); err != nil {
		return diag.FromErr(err)
	}

	if stack.Metadata != nil {

		if err := d.Set("labels", stack.Metadata.Labels); err != nil {
			return diag.FromErr(err)
		}

		// Handle components
		var components []map[string]string

		// We need to keep a sorted list of components, otherwise the order
		// of components will change on each read

		// Extract keys
		keys := make([]string, 0, len(stack.Metadata.Components))
		for key := range stack.Metadata.Components {
			keys = append(keys, key)
		}

		// Sort keys
		sort.Strings(keys)

		// Iterate over sorted keys
		for _, key := range keys {
			componentList := stack.Metadata.Components[key]
			var componentData map[string]string
			for _, component := range componentList {
				// Only take the first component of each type
				componentData = map[string]string{
					"id":     component.ID,
					"name":   component.Name,
					"type":   component.Body.Type,
					"flavor": component.Body.Flavor,
				}
				break
			}
			components = append(components, componentData)
		}
		if err := d.Set("components", components); err != nil {
			return diag.FromErr(err)
		}
	}

	if stack.Body != nil {
		if err := d.Set("created", stack.Body.Created); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("updated", stack.Body.Updated); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
