package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceCreate,
		ReadContext:   resourceWorkspaceRead,
		UpdateContext: resourceWorkspaceUpdate,
		DeleteContext: resourceWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the workspace",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the workspace",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Tags for the workspace",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Metadata for the workspace",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the workspace",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the workspace",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Update timestamp",
			},
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	tags := []string{}
	if v, ok := d.GetOk("tags"); ok {
		tags = convertSetToStringSlice(v.(*schema.Set))
	}
	metadata := map[string]string{}
	if v, ok := d.GetOk("metadata"); ok {
		metadata = convertMapToStringMap(v.(map[string]interface{}))
	}

	req := WorkspaceRequest{
		Name:        name,
		Description: &description,
		Tags:        tags,
		Metadata:    metadata,
	}

	workspace, err := client.CreateWorkspace(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(workspace.ID)

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	workspace, err := client.GetWorkspace(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if workspace == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", workspace.Name)
	d.Set("url", workspace.URL)
	d.Set("status", workspace.Status)

	if workspace.Body != nil {
		d.Set("description", workspace.Body.Description)
		d.Set("created", workspace.Body.Created)
		d.Set("updated", workspace.Body.Updated)
	}

	if workspace.Metadata != nil {
		d.Set("tags", workspace.Metadata.Tags)
		d.Set("metadata", workspace.Metadata.Metadata)
	}

	return nil
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var req WorkspaceUpdate

	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
	}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		req.Description = &description
	}

	if d.HasChange("tags") {
		tags := []string{}
		if v, ok := d.GetOk("tags"); ok {
			tags = convertSetToStringSlice(v.(*schema.Set))
		}
		req.Tags = tags
	}

	if d.HasChange("metadata") {
		metadata := map[string]string{}
		if v, ok := d.GetOk("metadata"); ok {
			metadata = convertMapToStringMap(v.(map[string]interface{}))
		}
		req.Metadata = metadata
	}

	_, err := client.UpdateWorkspace(ctx, d.Id(), req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceWorkspaceRead(ctx, d, meta)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	err := client.DeleteWorkspace(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for workspace to be fully deleted
	time.Sleep(5 * time.Second)

	return nil
}

func convertSetToStringSlice(set *schema.Set) []string {
	result := make([]string, set.Len())
	for i, v := range set.List() {
		result[i] = v.(string)
	}
	return result
}

func convertMapToStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = v.(string)
	}
	return result
}