package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProjectRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The ID of the project",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the project",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the workspace",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the project",
			},
			"tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Tags for the project",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
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

func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	workspaceID := d.Get("workspace_id").(string)
	var project *ProjectResponse
	var err error

	if id, ok := d.GetOk("id"); ok {
		project, err = client.GetProject(ctx, workspaceID, id.(string))
	} else if name, ok := d.GetOk("name"); ok {
		// Search for project by name
		params := &ListParams{
			Filter: map[string]string{
				"name": name.(string),
			},
		}
		
		projects, err := client.ListProjects(ctx, workspaceID, params)
		if err != nil {
			return diag.FromErr(err)
		}
		
		if len(projects.Items) == 0 {
			return diag.Errorf("project with name '%s' not found in workspace '%s'", name.(string), workspaceID)
		}
		
		if len(projects.Items) > 1 {
			return diag.Errorf("multiple projects found with name '%s' in workspace '%s'", name.(string), workspaceID)
		}
		
		project = &projects.Items[0]
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if project == nil {
		return diag.Errorf("project not found")
	}

	d.SetId(project.ID)
	d.Set("name", project.Name)
	d.Set("workspace_id", workspaceID)

	if project.Body != nil {
		d.Set("description", project.Body.Description)
		d.Set("created", project.Body.Created)
		d.Set("updated", project.Body.Updated)
	}

	if project.Metadata != nil {
		d.Set("tags", project.Metadata.Tags)
		d.Set("metadata", project.Metadata.Metadata)
	}

	return nil
}