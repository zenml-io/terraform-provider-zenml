package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceStack() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for ZenML stacks",
		ReadContext: dataSourceStackRead,
		Schema: map[string]*schema.Schema{
			"workspace": {
				Description: "Name of the workspace",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name of the stack",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description of the stack",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"components": {
				Description: "Components configured in the stack",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
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
							"configuration": {
								Type:     schema.TypeMap,
								Computed: true,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
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
			"user": {
				Description: "User who created the stack",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceStackRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)

	// List stacks with filter to find by name
	params := &ListParams{
		Filter: map[string]string{
			"name":      name,
			"workspace": workspace,
		},
	}

	stacks, err := c.ListStacks(params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing stacks: %v", err))
	}

	if len(stacks.Items) == 0 {
		return diag.FromErr(fmt.Errorf("no stack found with name %s in workspace %s", name, workspace))
	}

	stack := stacks.Items[0]
	d.SetId(stack.ID)

	if err := d.Set("name", stack.Name); err != nil {
		return diag.FromErr(err)
	}

	if stack.Metadata != nil {
		if err := d.Set("description", stack.Metadata.Description); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("labels", stack.Metadata.Labels); err != nil {
			return diag.FromErr(err)
		}

		// Handle components
		components := make(map[string][]interface{})
		for componentType, componentList := range stack.Metadata.Components {
			componentData := make([]interface{}, len(componentList))
			for i, component := range componentList {
				componentData[i] = map[string]interface{}{
					"id":            component.ID,
					"name":          component.Name,
					"type":          component.Body.Type,
					"flavor":        component.Body.Flavor,
					"configuration": component.Metadata.Configuration,
				}
			}
			components[componentType] = componentData
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

		if stack.Body.User != nil {
			userData := map[string]interface{}{
				"id":       stack.Body.User.ID,
				"name":     stack.Body.User.Name,
				"active":   stack.Body.User.Body.Active,
				"is_admin": stack.Body.User.Body.IsAdmin,
			}
			if err := d.Set("user", userData); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}
