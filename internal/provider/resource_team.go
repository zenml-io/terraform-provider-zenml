package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"control_plane_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the control plane",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the team",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the team",
			},
			"members": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Email addresses of team members",
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

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	controlPlaneID := d.Get("control_plane_id").(string)
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	members := []string{}
	if v, ok := d.GetOk("members"); ok {
		members = convertSetToStringSlice(v.(*schema.Set))
	}

	req := TeamRequest{
		ControlPlaneID: controlPlaneID,
		Name:           name,
		Description:    &description,
		Members:        members,
	}

	team, err := client.CreateTeam(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(team.ID)

	return resourceTeamRead(ctx, d, meta)
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	team, err := client.GetTeam(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if team == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", team.Name)

	if team.Body != nil {
		d.Set("description", team.Body.Description)
		d.Set("control_plane_id", team.Body.ControlPlaneID)
		d.Set("created", team.Body.Created)
		d.Set("updated", team.Body.Updated)
	}

	if team.Metadata != nil && team.Metadata.Members != nil {
		memberEmails := make([]string, len(team.Metadata.Members))
		for i, member := range team.Metadata.Members {
			memberEmails[i] = member.Email
		}
		d.Set("members", memberEmails)
	}

	return nil
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var req TeamUpdate

	if d.HasChange("name") {
		name := d.Get("name").(string)
		req.Name = &name
	}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		req.Description = &description
	}

	if d.HasChange("members") {
		members := []string{}
		if v, ok := d.GetOk("members"); ok {
			members = convertSetToStringSlice(v.(*schema.Set))
		}
		req.Members = members
	}

	_, err := client.UpdateTeam(ctx, d.Id(), req)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceTeamRead(ctx, d, meta)
}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	err := client.DeleteTeam(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}