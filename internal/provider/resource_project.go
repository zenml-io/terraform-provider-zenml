// resource_project.go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique name of the project",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The display name of the project",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the project",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp when the project was created",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp when the project was last updated",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	projectRequest := ProjectRequest{
		Name:        d.Get("name").(string),
		DisplayName: d.Get("display_name").(string),
		Description: d.Get("description").(string),
	}

	project, err := client.CreateProject(ctx, projectRequest)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project.ID)

	return resourceProjectRead(ctx, d, m)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	project, err := client.GetProject(ctx, d.Id())
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("name", project.Name); err != nil {
		return diag.FromErr(err)
	}

	if project.Body != nil {
		if err := d.Set("display_name", project.Body.DisplayName); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("created", project.Body.Created); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("updated", project.Body.Updated); err != nil {
			return diag.FromErr(err)
		}
	}

	if project.Metadata != nil {
		if err := d.Set("description", project.Metadata.Description); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	projectUpdate := ProjectUpdate{}

	if d.HasChange("name") {
		name := d.Get("name").(string)
		projectUpdate.Name = &name
	}

	if d.HasChange("display_name") {
		displayName := d.Get("display_name").(string)
		projectUpdate.DisplayName = &displayName
	}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		projectUpdate.Description = &description
	}

	_, err := client.UpdateProject(ctx, d.Id(), projectUpdate)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceProjectRead(ctx, d, m)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	err := client.DeleteProject(ctx, d.Id())
	if err != nil {
		if isNotFoundError(err) {
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func isNotFoundError(err error) bool {
	return err != nil && (err.Error() == "404" || fmt.Sprintf("%v", err) == "404")
}