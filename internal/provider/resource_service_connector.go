// resource_service_connector.go
package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceServiceConnector() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceConnectorCreate,
		Read:   resourceServiceConnectorRead,
		Update: resourceServiceConnectorUpdate,
		Delete: resourceServiceConnectorDelete,

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
					"aws", "gcp", "azure", "kubernetes",
					"github", "gitlab", "bitbucket", "docker",
					"mysql", "postgres", "snowflake", "databricks",
				}, false),
			},
			"auth_method": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"configuration": {
				Type:      schema.TypeMap,
				Required:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"secrets": {
				Type:      schema.TypeMap,
				Optional:  true,
				Sensitive: true,
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

func resourceServiceConnectorCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	connector := ServiceConnectorBody{
		User:      d.Get("user").(string),
		Workspace: d.Get("workspace").(string),
	}

	// Handle configuration
	if v, ok := d.GetOk("configuration"); ok {
		configMap := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			configMap[k] = v
		}
		connector.Configuration = configMap
	}

	// Handle secrets
	if v, ok := d.GetOk("secrets"); ok {
		secretsMap := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			secretsMap[k] = v
		}
		connector.Secrets = secretsMap
	}

	// Handle resource types
	if v, ok := d.GetOk("resource_types"); ok {
		resourceTypesSet := v.(*schema.Set)
		resourceTypes := make([]string, resourceTypesSet.Len())
		for i, rt := range resourceTypesSet.List() {
			resourceTypes[i] = rt.(string)
		}
		connector.ResourceTypes = resourceTypes
	}

	// Handle labels
	if v, ok := d.GetOk("labels"); ok {
		labelsMap := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			labelsMap[k] = v.(string)
		}
		connector.Labels = labelsMap
	}

	resp, err := client.CreateServiceConnector(connector)
	if err != nil {
		return err
	}

	d.SetId(resp.ID)
	return resourceServiceConnectorRead(d, m)
}

func resourceServiceConnectorRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	connector, err := client.GetServiceConnector(d.Id())
	if err != nil {
		d.SetId("")
		return nil
	}

	d.Set("name", connector.Name)
	d.Set("type", connector.Type)
	d.Set("auth_method", connector.AuthMethod)

	if connector.Body != nil {
		d.Set("user", connector.Body.User)
		d.Set("workspace", connector.Body.Workspace)
		d.Set("configuration", connector.Body.Configuration)
		d.Set("resource_types", connector.Body.ResourceTypes)

		if connector.Body.Labels != nil {
			d.Set("labels", connector.Body.Labels)
		}

		// Don't set secrets back as they are sensitive
	}

	return nil
}

func resourceServiceConnectorUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	update := ServiceConnectorUpdate{
		Name: d.Get("name").(string),
	}

	if d.HasChange("configuration") {
		configMap := make(map[string]interface{})
		for k, v := range d.Get("configuration").(map[string]interface{}) {
			configMap[k] = v
		}
		update.Configuration = configMap
	}

	if d.HasChange("secrets") {
		secretsMap := make(map[string]interface{})
		for k, v := range d.Get("secrets").(map[string]interface{}) {
			secretsMap[k] = v
		}
		update.Secrets = secretsMap
	}

	if d.HasChange("resource_types") {
		resourceTypesSet := d.Get("resource_types").(*schema.Set)
		resourceTypes := make([]string, resourceTypesSet.Len())
		for i, rt := range resourceTypesSet.List() {
			resourceTypes[i] = rt.(string)
		}
		update.ResourceTypes = resourceTypes
	}

	if d.HasChange("labels") {
		labelsMap := make(map[string]string)
		for k, v := range d.Get("labels").(map[string]interface{}) {
			labelsMap[k] = v.(string)
		}
		update.Labels = labelsMap
	}

	_, err := client.UpdateServiceConnector(d.Id(), update)
	if err != nil {
		return err
	}

	return resourceServiceConnectorRead(d, m)
}

func resourceServiceConnectorDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	err := client.DeleteServiceConnector(d.Id())
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
