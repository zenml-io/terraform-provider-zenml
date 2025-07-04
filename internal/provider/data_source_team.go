package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTeam() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTeamRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The ID of the team",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the team",
			},
			"control_plane_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the control plane",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the team",
			},
			"members": {
				Type:        schema.TypeSet,
				Computed:    true,
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

func dataSourceTeamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	var team *TeamResponse
	var err error

	if id, ok := d.GetOk("id"); ok {
		team, err = client.GetTeam(ctx, id.(string))
	} else if name, ok := d.GetOk("name"); ok {
		// Search for team by name
		params := &ListParams{
			Filter: map[string]string{
				"name": name.(string),
			},
		}
		
		teams, err := client.ListTeams(ctx, params)
		if err != nil {
			return diag.FromErr(err)
		}
		
		if len(teams.Items) == 0 {
			return diag.Errorf("team with name '%s' not found", name.(string))
		}
		
		if len(teams.Items) > 1 {
			return diag.Errorf("multiple teams found with name '%s'", name.(string))
		}
		
		team = &teams.Items[0]
	} else {
		return diag.Errorf("either 'id' or 'name' must be specified")
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if team == nil {
		return diag.Errorf("team not found")
	}

	d.SetId(team.ID)
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