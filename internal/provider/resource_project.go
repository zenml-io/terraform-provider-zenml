package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the workspace",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the project",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the project",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Tags for the project",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Metadata for the project",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	workspaceID := d.Get("workspace_id").(string)
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

	req := ProjectRequest{
		WorkspaceID: workspaceID,
		Name:        name,
		Description: &description,
		Tags:        tags,
		Metadata:    metadata,
	}

	project, err := client.CreateProject(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.ID)

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	workspaceID := d.Get("workspace_id").(string)
	project, err := client.GetProject(ctx, workspaceID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if project == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", project.Name)

	if project.Body != nil {
		d.Set("description", project.Body.Description)
		d.Set("workspace_id", project.Body.WorkspaceID)
		d.Set("created", project.Body.Created)
		d.Set("updated", project.Body.Updated)
	}

	if project.Metadata != nil {
		d.Set("tags", project.Metadata.Tags)
		d.Set("metadata", project.Metadata.Metadata)
	}

	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var req ProjectUpdate
	workspaceID := d.Get("workspace_id").(string)

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

	_, err := client.UpdateProject(ctx, workspaceID, d.Id(), req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	workspaceID := d.Get("workspace_id").(string)
	err := client.DeleteProject(ctx, workspaceID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}