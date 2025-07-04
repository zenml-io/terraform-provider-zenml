package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStackRoleAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStackRoleAssignmentCreate,
		ReadContext:   resourceStackRoleAssignmentRead,
		UpdateContext: resourceStackRoleAssignmentUpdate,
		DeleteContext: resourceStackRoleAssignmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"stack_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the stack",
			},
			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID of the user (mutually exclusive with team_id)",
			},
			"team_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID of the team (mutually exclusive with user_id)",
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role to assign (e.g., 'admin', 'write', 'read')",
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

func resourceStackRoleAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	stackID := d.Get("stack_id").(string)
	role := d.Get("role").(string)
	
	var userID, teamID *string
	if v, ok := d.GetOk("user_id"); ok {
		s := v.(string)
		userID = &s
	}
	if v, ok := d.GetOk("team_id"); ok {
		s := v.(string)
		teamID = &s
	}

	// Validate that exactly one of user_id or team_id is provided
	if (userID == nil && teamID == nil) || (userID != nil && teamID != nil) {
		return diag.Errorf("exactly one of user_id or team_id must be provided")
	}

	req := RoleAssignmentRequest{
		ResourceID:   stackID,
		ResourceType: "stack",
		UserID:       userID,
		TeamID:       teamID,
		Role:         role,
	}

	assignment, err := client.CreateRoleAssignment(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(assignment.ID)

	return resourceStackRoleAssignmentRead(ctx, d, meta)
}

func resourceStackRoleAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	assignment, err := client.GetRoleAssignment(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if assignment == nil {
		d.SetId("")
		return nil
	}

	d.Set("stack_id", assignment.ResourceID)

	if assignment.Body != nil {
		d.Set("role", assignment.Body.Role)
		d.Set("created", assignment.Body.Created)
		d.Set("updated", assignment.Body.Updated)
		
		if assignment.Body.UserID != nil {
			d.Set("user_id", *assignment.Body.UserID)
		}
		if assignment.Body.TeamID != nil {
			d.Set("team_id", *assignment.Body.TeamID)
		}
	}

	return nil
}

func resourceStackRoleAssignmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	if d.HasChange("role") {
		role := d.Get("role").(string)
		req := RoleAssignmentUpdate{
			Role: role,
		}

		_, err := client.UpdateRoleAssignment(ctx, d.Id(), req)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceStackRoleAssignmentRead(ctx, d, meta)
}

func resourceStackRoleAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	err := client.DeleteRoleAssignment(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}