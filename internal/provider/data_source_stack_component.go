package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceStackComponent() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for ZenML stack components",
		ReadContext: dataSourceStackComponentRead,
		Schema: map[string]*schema.Schema{
			"workspace": {
				Description: "Name of the workspace (defaults to 'default')",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
			},
			"name": {
				Description: "Name of the stack component",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "Type of the stack component",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"alerter",
					"annotator",
					"artifact_store",
					"container_registry",
					"data_validator",
					"experiment_tracker",
					"feature_store",
					"image_builder",
					"model_deployer",
					"orchestrator",
					"step_operator",
					"model_registry",
				}, false),
			},
			"flavor": {
				Description: "Flavor of the stack component",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"configuration": {
				Description: "Configuration of the stack component",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Sensitive: true,
			},
			"labels": {
				Description: "Labels associated with the stack component",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"component_spec_path": {
				Description: "Path to the component specification file",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"connector": {
				Description: "Service connector configuration",
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
						"connector_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_types": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"created": {
				Description: "Timestamp when the component was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"updated": {
				Description: "Timestamp when the component was last updated",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"user": {
				Description: "User who created the component",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceStackComponentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	workspace := d.Get("workspace").(string)
	name := d.Get("name").(string)
	componentType := d.Get("type").(string)

	// List components with filters
	params := &ListParams{
		Filter: map[string]string{
			"name":      name,
			"workspace": workspace,
			"type":      componentType,
		},
	}

	components, err := c.ListStackComponents(workspace, params)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing stack components: %v", err))
	}

	if len(components.Items) == 0 {
		return diag.FromErr(fmt.Errorf("no component found with name %s and type %s in workspace %s", 
			name, componentType, workspace))
	}

	component := components.Items[0]
	d.SetId(component.ID)

	if err := d.Set("name", component.Name); err != nil {
		return diag.FromErr(err)
	}

	if component.Body != nil {
		if err := d.Set("flavor", component.Body.Flavor); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("created", component.Body.Created); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("updated", component.Body.Updated); err != nil {
			return diag.FromErr(err)
		}

		if component.Body.User != nil {
			userData := map[string]interface{}{
				"id":       component.Body.User.ID,
				"name":     component.Body.User.Name,
				"active":   component.Body.User.Body.Active,
				"is_admin": component.Body.User.Body.IsAdmin,
			}
			if err := d.Set("user", userData); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if component.Metadata != nil {
		if err := d.Set("configuration", component.Metadata.Configuration); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("labels", component.Metadata.Labels); err != nil {
			return diag.FromErr(err)
		}

		if component.Metadata.ComponentSpecPath != nil {
			if err := d.Set("component_spec_path", *component.Metadata.ComponentSpecPath); err != nil {
				return diag.FromErr(err)
			}
		}

		if component.Metadata.Connector != nil {
			connector := []interface{}{
				map[string]interface{}{
					"id":             component.Metadata.Connector.ID,
					"name":           component.Metadata.Connector.Name,
					"connector_type": component.Metadata.Connector.Body.ConnectorType,
					"resource_id":    component.Metadata.Connector.Body.ResourceID,
					"resource_types": component.Metadata.Connector.Body.ResourceTypes,
				},
			}
			if err := d.Set("connector", connector); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}
