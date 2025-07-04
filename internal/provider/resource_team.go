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
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the organization",
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
				Description: "Email addresses or user IDs of team members",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	organizationID := d.Get("organization_id").(string)
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	members := []string{}
	if v, ok := d.GetOk("members"); ok {
		members = convertSetToStringSlice(v.(*schema.Set))
	}

	req := TeamRequest{
		OrganizationID: organizationID,
		Name:           name,
		Description:    &description,
	}

	team, err := client.CreateTeam(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(team.ID)

	// Add members to the team (separate API calls in the real API)
	for _, memberID := range members {
		err := client.AddTeamMember(ctx, team.ID, memberID)
		if err != nil {
			return diag.Errorf("failed to add member %s to team: %v", memberID, err)
		}
	}

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
	d.Set("description", team.Description)
	d.Set("member_count", team.MemberCount)
	d.Set("created", team.Created)
	d.Set("updated", team.Updated)

	// Get team members (separate API call in the real API)
	members, err := client.ListTeamMembers(ctx, team.ID)
	if err != nil {
		// Don't fail if we can't get members, just log and continue
		// In a real implementation, you might want to handle this differently
	} else {
		memberIDs := make([]string, len(members))
		for i, member := range members {
			if member.User != nil {
				memberIDs[i] = member.User.ID
			}
		}
		d.Set("members", memberIDs)
	}

	return nil
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	// Update team basic info
	if d.HasChange("name") || d.HasChange("description") {
		var req TeamUpdate

		if d.HasChange("name") {
			name := d.Get("name").(string)
			req.Name = &name
		}

		if d.HasChange("description") {
			description := d.Get("description").(string)
			req.Description = &description
		}

		_, err := client.UpdateTeam(ctx, d.Id(), req)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle member changes (separate API calls in the real API)
	if d.HasChange("members") {
		old, new := d.GetChange("members")
		oldMembers := convertSetToStringSlice(old.(*schema.Set))
		newMembers := convertSetToStringSlice(new.(*schema.Set))

		// Find members to remove
		for _, oldMember := range oldMembers {
			found := false
			for _, newMember := range newMembers {
				if oldMember == newMember {
					found = true
					break
				}
			}
			if !found {
				err := client.RemoveTeamMember(ctx, d.Id(), oldMember)
				if err != nil {
					return diag.Errorf("failed to remove member %s from team: %v", oldMember, err)
				}
			}
		}

		// Find members to add
		for _, newMember := range newMembers {
			found := false
			for _, oldMember := range oldMembers {
				if newMember == oldMember {
					found = true
					break
				}
			}
			if !found {
				err := client.AddTeamMember(ctx, d.Id(), newMember)
				if err != nil {
					return diag.Errorf("failed to add member %s to team: %v", newMember, err)
				}
			}
		}
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