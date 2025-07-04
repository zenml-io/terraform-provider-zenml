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
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the team",
			},
			"member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of team members",
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
	d.Set("description", team.Description)
	d.Set("member_count", team.MemberCount)
	d.Set("created", team.Created)
	d.Set("updated", team.Updated)

	return nil
}