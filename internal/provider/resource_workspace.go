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
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Display name of the workspace",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the workspace",
			},
			"logo_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Logo URL of the workspace",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID of the organization (optional)",
			},
			"is_managed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the workspace is managed by ZenML Pro",
			},
			"server_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server URL of the workspace",
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
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	logoURL := d.Get("logo_url").(string)
	organizationID := d.Get("organization_id").(string)
	isManaged := d.Get("is_managed").(bool)

	req := WorkspaceRequest{
		Name:        &name,
		DisplayName: &displayName,
		IsManaged:   isManaged,
	}

	if description != "" {
		req.Description = &description
	}
	if logoURL != "" {
		req.LogoURL = &logoURL
	}
	if organizationID != "" {
		req.OrganizationID = &organizationID
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
	d.Set("display_name", workspace.DisplayName)
	d.Set("description", workspace.Description)
	d.Set("logo_url", workspace.LogoURL)
	d.Set("is_managed", workspace.IsManaged)
	d.Set("status", workspace.Status)
	d.Set("created", workspace.Created)
	d.Set("updated", workspace.Updated)

	// Set server URL from ZenML service
	if workspace.ZenMLService.Status != nil && workspace.ZenMLService.Status.ServerURL != nil {
		d.Set("server_url", *workspace.ZenMLService.Status.ServerURL)
	}

	return nil
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var req WorkspaceUpdate

	if d.HasChange("display_name") {
		displayName := d.Get("display_name").(string)
		req.DisplayName = &displayName
	}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		req.Description = &description
	}

	if d.HasChange("logo_url") {
		logoURL := d.Get("logo_url").(string)
		req.LogoURL = &logoURL
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