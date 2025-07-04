package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWorkspace() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWorkspaceRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The ID of the workspace",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the workspace",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the workspace",
			},
			"tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Tags for the workspace",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
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

func dataSourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var workspace *WorkspaceResponse
	var err error

	if id, ok := d.GetOk("id"); ok {
		workspace, err = client.GetWorkspace(ctx, id.(string))
	} else if name, ok := d.GetOk("name"); ok {
		// Search for workspace by name
		params := &ListParams{
			Filter: map[string]string{
				"name": name.(string),
			},
		}
		
		workspaces, err := client.ListWorkspaces(ctx, params)
		if err != nil {
			return diag.FromErr(err)
		}
		
		if len(workspaces.Items) == 0 {
			return diag.Errorf("workspace with name '%s' not found", name.(string))
		}
		
		if len(workspaces.Items) > 1 {
			return diag.Errorf("multiple workspaces found with name '%s'", name.(string))
		}
		
		workspace = &workspaces.Items[0]
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if workspace == nil {
		return diag.Errorf("workspace not found")
	}

	d.SetId(workspace.ID)
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