// resource_stack_component.go
package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceStackComponent() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackComponentCreate,
		Read:   resourceStackComponentRead,
		Update: resourceStackComponentUpdate,
		Delete: resourceStackComponentDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"configuration": {
				Type:     schema.TypeMap,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"connector_resource_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"component_spec_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"connector": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		ImportState: schema.ImportStatePassthrough,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceStackComponentCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	component := ComponentBody{
		User:      d.Get("user").(string),
		Workspace: d.Get("workspace").(string),
	}

	// Handle configuration
	if v, ok := d.GetOk("configuration"); ok {
		configMap := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			configMap[k] = v
		}
		component.Configuration = configMap
	}

	// Handle optional fields
	if v, ok := d.GetOk("connector_resource_id"); ok {
		str := v.(string)
		component.ConnectorResourceID = &str
	}

	if v, ok := d.GetOk("labels"); ok {
		labelsMap := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			labelsMap[k] = v.(string)
		}
		component.Labels = labelsMap
	}

	if v, ok := d.GetOk("component_spec_path"); ok {
		str := v.(string)
		component.ComponentSpecPath = &str
	}

	if v, ok := d.GetOk("connector"); ok {
		str := v.(string)
		component.Connector = &str
	}

	resp, err := client.CreateComponent(component)
	if err != nil {
		return err
	}

	d.SetId(resp.ID)
	return resourceStackComponentRead(d, m)
}

// resource_stack_component.go (continued)

func resourceStackComponentRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	component, err := client.GetComponent(d.Id())
	if err != nil {
		// Handle 404 by removing from state
		d.SetId("")
		return nil
	}

	d.Set("name", component.Name)
	d.Set("type", component.Type)
	d.Set("flavor", component.Flavor)

	if component.Body != nil {
		d.Set("user", component.Body.User)
		d.Set("workspace", component.Body.Workspace)
		d.Set("configuration", component.Body.Configuration)
		
		if component.Body.ConnectorResourceID != nil {
			d.Set("connector_resource_id", *component.Body.ConnectorResourceID)
		}
		
		if component.Body.Labels != nil {
			d.Set("labels", component.Body.Labels)
		}
		
		if component.Body.ComponentSpecPath != nil {
			d.Set("component_spec_path", *component.Body.ComponentSpecPath)
		}
		
		if component.Body.Connector != nil {
			d.Set("connector", *component.Body.Connector)
		}
	}

	return nil
}

func resourceStackComponentUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	update := ComponentUpdate{
		Name: d.Get("name").(string),
	}

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

	if d.HasChange("component_spec_path") {
		if v, ok := d.GetOk("component_spec_path"); ok {
			str := v.(string)
			update.ComponentSpecPath = &str
		}
	}

	if d.HasChange("connector") {
		if v, ok := d.GetOk("connector"); ok {
			str := v.(string)
			update.Connector = &str
		}
	}

	_, err := client.UpdateComponent(d.Id(), update)
	if err != nil {
		return err
	}

	return resourceStackComponentRead(d, m)
}

func resourceStackComponentDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	err := client.DeleteComponent(d.Id())
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

// resource_stack_component.go
func resourceStackComponent() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackComponentCreate,
		Read:   resourceStackComponentRead,
		Update: resourceStackComponentUpdate,
		Delete: resourceStackComponentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			// ... existing fields ...
			"connector_resource_id": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "ID of a specific resource instance to gain access to through the connector",
			},
		},

		CustomizeDiff: validateStackComponent,
	}
}

func validateStackComponent(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	componentType := d.Get("type").(string)
	flavor := d.Get("flavor").(string)
	configuration := d.Get("configuration").(map[string]interface{})

	// Validate component type and flavor combination
	if err := validateComponentTypeAndFlavor(componentType, flavor); err != nil {
		return err
	}

	// Validate configuration based on component type and flavor
	if err := validateComponentConfiguration(componentType, flavor, configuration); err != nil {
		return err
	}

	return nil
}