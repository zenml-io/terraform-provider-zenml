// resource_service_connector.go
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceServiceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceConnectorCreate,
		ReadContext:   resourceServiceConnectorRead,
		UpdateContext: resourceServiceConnectorUpdate,
		DeleteContext: resourceServiceConnectorDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(validConnectorTypes, false),
			},
			"auth_method": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Optional: true,
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
				Computed: true,
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
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func getConnectorRequest(ctx context.Context, d *schema.ResourceData, client *Client) (*ServiceConnectorRequest, error) {

	// Get the current user
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %w", err)
	}

	connector := ServiceConnectorRequest{
		User:          user.ID,
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

	// Handle resource type
	if v, ok := d.GetOk("resource_type"); ok {
		resourceType := v.(string)
		resourceTypes := []string{resourceType}
		connector.ResourceTypes = resourceTypes
	} else {
		connector.ResourceTypes = []string{}
	}

	// Handle resource ID
	if v, ok := d.GetOk("resource_id"); ok {
		resourceID := v.(string)
		connector.ResourceID = &resourceID
	}

	// Handle labels
	if v, ok := d.GetOk("labels"); ok {
		labelsMap := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			labelsMap[k] = v.(string)
		}
		connector.Labels = labelsMap
	}

	return &connector, nil
}

func resourceServiceConnectorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	connector, err := getConnectorRequest(ctx, d, client)

	if err != nil {
		return diag.FromErr(err)
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *retry.RetryError {
		verify, err := client.VerifyServiceConnector(ctx, *connector)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error verifying service connector: %s", err))
		}

		if verify.Error != nil {
			return retry.RetryableError(fmt.Errorf("error verifying service connector: %s", *verify.Error))
		}

		resp, err := client.CreateServiceConnector(ctx, *connector)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error creating service connector: %s", err))
		}

		d.SetId(resp.ID)

		return nil

	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServiceConnectorRead(ctx, d, m)
}

func resourceServiceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	connector, err := client.GetServiceConnector(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting service connector: %s", err))
	}

	if connector == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", connector.Name)

	if connector.Body != nil {

		if connector.Body.ResourceID != nil {
			d.Set("resource_id", connector.Body.ResourceID)
		}

		connector_type := ""

		// Unmarshal the connector type, which can be either a string or a struct
		// Try to unmarshal as string
		err = json.Unmarshal(connector.Body.ConnectorType, &connector_type)
		if err != nil {
			var type_struct ServiceConnectorType
			// Try to unmarshal as struct
			if err = json.Unmarshal(connector.Body.ConnectorType, &type_struct); err == nil {
				connector_type = type_struct.ConnectorType
			} else {
				return diag.FromErr(fmt.Errorf("error unmarshalling connector type: %s", err))
			}

		}
		d.Set("type", connector_type)

		d.Set("auth_method", connector.Body.AuthMethod)

		// If there are multiple resource types, leave the resource_type field empty
		if len(connector.Body.ResourceTypes) == 1 {
			d.Set("resource_type", connector.Body.ResourceTypes[0])
		} else {
			d.Set("resource_type", "")
		}
	}

	if connector.Metadata != nil {
		d.Set("configuration", connector.Metadata.Configuration)
		d.Set("labels", connector.Metadata.Labels)
	}

	return nil
}

func resourceServiceConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	connector, err := getConnectorRequest(ctx, d, client)

	if err != nil {
		return diag.FromErr(err)
	}

	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *retry.RetryError {
		resources, err := client.VerifyServiceConnector(ctx, *connector)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error verifying service connector: %s", err))
		}

		if resources.Error != nil {
			return retry.RetryableError(fmt.Errorf("error verifying service connector: %s", *resources.Error))
		}

		update := ServiceConnectorUpdate{}

		if d.HasChange("name") {
			name := d.Get("name").(string)
			update.Name = &name
		}

		// The `configuration` field represents a full valid configuration update,
		// not just a partial update. If it is set (i.e. not None) in the update,
		// the value will replace the existing configuration value. For this
		// reason, we always include the configuration in the update request.

		// Handle configuration
		configMap := make(map[string]interface{})
		if v, ok := d.GetOk("configuration"); ok {
			for k, v := range v.(map[string]interface{}) {
				configMap[k] = v
			}
		}
		update.Configuration = configMap

		// The `labels` field is also a full labels update: if set (i.e. not
		// `None`), all existing labels are removed and replaced by the new labels
		// in the update.

		labelsMap := make(map[string]string)
		if d.HasChange("labels") {
			for k, v := range d.Get("labels").(map[string]interface{}) {
				labelsMap[k] = v.(string)
			}
		}
		update.Labels = labelsMap

		// The `resource_id` field value is also a full replacement value: if not
		// set in the request, the resource ID is removed from the service
		// connector.
		if v, ok := d.GetOk("resource_id"); ok {
			resourceID := v.(string)
			update.ResourceID = &resourceID
		} else {
			update.ResourceID = nil
		}

		// Handle resource type
		if v, ok := d.GetOk("resource_type"); ok {
			resourceType := v.(string)
			resourceTypes := []string{resourceType}
			update.ResourceTypes = resourceTypes
		} else {
			update.ResourceTypes = []string{}
		}

		_, err = client.UpdateServiceConnector(ctx, d.Id(), update)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Error updating service connector: %s", err))
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServiceConnectorRead(ctx, d, m)
}

func resourceServiceConnectorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	err := client.DeleteServiceConnector(ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting service connector: %s", err))
	}

	d.SetId("")
	return nil
}
