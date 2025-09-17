package provider

import (
	"context"
	"encoding/json"
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
			"id": {
				Description: "ID of the stack component",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "Name of the stack component",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"type": {
				Description: "Type of the stack component",
				Type:        schema.TypeString,
				Optional:    true,
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
					"deployer",
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
			"connector_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
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
		},
	}
}

func dataSourceStackComponentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	id := d.Get("id").(string)
	name := d.Get("name").(string)
	componentType := d.Get("type").(string)

	var component *ComponentResponse = nil
	var err error = nil

	if id != "" {
		component, err = c.GetComponent(ctx, id)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error getting stack component: %v", err))
		}
	} else if name != "" && componentType != "" {
		// List components with filters
		params := &ListParams{
			Filter: map[string]string{
				"name": name,
				"type": componentType,
			},
		}

		components, err := c.ListStackComponents(ctx, params)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing stack components: %v", err))
		}

		if len(components.Items) == 0 {
			return diag.FromErr(fmt.Errorf("no component found with name %s and type %s",
				name, componentType))
		}

		component = &components.Items[0]

	} else {
		return diag.FromErr(fmt.Errorf("either 'id' or 'name' and 'type' must be set"))
	}

	if component == nil {
		// Component not found
		d.SetId("")
		return nil
	}

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
	}

	if component.Metadata != nil {
		if err := d.Set("configuration", component.Metadata.Configuration); err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("labels", component.Metadata.Labels); err != nil {
			return diag.FromErr(err)
		}

		if component.Metadata.Connector != nil {

			connectorType := ""
			resourceId := ""
			resourceTypes := []string{}

			if component.Metadata.Connector.Body != nil {

				// Unmarshal the connector type, which can be either a string or a struct
				// Try to unmarshal as string
				err := json.Unmarshal(component.Metadata.Connector.Body.ConnectorType, &connectorType)
				if err != nil {
					var typeStruct ServiceConnectorType
					// Try to unmarshal as struct
					if err := json.Unmarshal(component.Metadata.Connector.Body.ConnectorType, &typeStruct); err == nil {
						connectorType = typeStruct.ConnectorType
					}
				}

				if component.Metadata.Connector.Body.ResourceID != nil {
					resourceId = *component.Metadata.Connector.Body.ResourceID
				}

				resourceTypes = component.Metadata.Connector.Body.ResourceTypes
			}

			connector := []interface{}{
				map[string]interface{}{
					"id":             component.Metadata.Connector.ID,
					"name":           component.Metadata.Connector.Name,
					"connector_type": connectorType,
					"resource_id":    resourceId,
					"resource_types": resourceTypes,
				},
			}
			if err := d.Set("connector", connector); err != nil {
				return diag.FromErr(err)
			}
		}

		if component.Metadata.ConnectorResourceID != nil {
			if err := d.Set("connector_resource_id", *component.Metadata.ConnectorResourceID); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := d.Set("connector_resource_id", ""); err != nil {
				return diag.FromErr(err)
			}
		}

	}

	return nil
}
