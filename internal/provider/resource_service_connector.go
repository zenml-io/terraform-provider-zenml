// resource_service_connector.go
package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"context"
	"fmt"
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
				ValidateFunc: validation.StringInSlice(validConnectorTypes, false),
			},
			"auth_method": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"iam-role", "aws-access-keys", "web-identity",
					"service-account", "oauth2", "workload-identity",
					"service-principal", "managed-identity",
					"kubeconfig", "service-account",
				}, false),
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
				Computed: true,
				ForceNew: true,
			},
			"workspace": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
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

		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
			return validateServiceConnector(d)
		},

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceServiceConnectorCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	// Get the current user
	user, err := client.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("error getting current user: %w", err)
	}

	// Get workspace name, defaulting to "default"
	workspaceName := d.Get("workspace").(string)
	
	// Get the workspace ID
	workspace, err := client.GetWorkspaceByName(workspaceName)
	if err != nil {
		return fmt.Errorf("error getting workspace: %w", err)
	}

	connector := ServiceConnectorRequest{
		User:          user.ID,
		Workspace:     workspace.ID,  // Use the workspace ID from the response
		Name:          d.Get("name").(string),
		ConnectorType: d.Get("type").(string),
		AuthMethod:    d.Get("auth_method").(string),
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
		secretsMap := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			secretsMap[k] = v.(string)
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

	resp, err := client.CreateServiceConnector(connector.Workspace, connector)
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
	d.Set("type", connector.Body.ConnectorType)
	d.Set("auth_method", connector.Body.AuthMethod)
	d.Set("resource_types", connector.Body.ResourceTypes)

	if connector.Body != nil {
		d.Set("user", connector.Body.User.Name)
		if connector.Metadata != nil {
			d.Set("workspace", connector.Metadata.Workspace.Name)
			d.Set("configuration", connector.Metadata.Configuration)
			d.Set("labels", connector.Metadata.Labels)
		}
		// Don't set secrets back as they are sensitive
	}

	return nil
}

func resourceServiceConnectorUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	update := ServiceConnectorUpdate{}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		update.Name = &name
	}

	if d.HasChange("configuration") {
		configMap := make(map[string]interface{})
		for k, v := range d.Get("configuration").(map[string]interface{}) {
			configMap[k] = v
		}
		update.Configuration = configMap
	}

	if d.HasChange("secrets") {
		secretsMap := make(map[string]string)
		for k, v := range d.Get("secrets").(map[string]interface{}) {
			secretsMap[k] = v.(string)
		}
		update.Secrets = secretsMap
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
