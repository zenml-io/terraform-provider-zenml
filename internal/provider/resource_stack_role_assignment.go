package provider

import (
	"context"
	"strings"

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
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the role to assign",
			},
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
			"assignment_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the role assignment",
			},
		},
	}
}

func resourceStackRoleAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	roleID := d.Get("role_id").(string)
	stackID := d.Get("stack_id").(string)
	
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
		RoleID:      roleID,
		UserID:      userID,
		TeamID:      teamID,
		// Note: Stack-level assignments might need special handling in the real API
	}

	_, err := client.CreateRoleAssignment(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create a composite ID: role_id:assignee_id:stack:stack_id
	var assigneeID string
	if userID != nil {
		assigneeID = *userID
	} else {
		assigneeID = *teamID
	}
	d.SetId(roleID + ":" + assigneeID + ":stack:" + stackID)

	return resourceStackRoleAssignmentRead(ctx, d, meta)
}

func resourceStackRoleAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	// Parse composite ID: role_id:assignee_id:stack:stack_id
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 4 || idParts[2] != "stack" {
		return diag.Errorf("invalid ID format, expected role_id:assignee_id:stack:stack_id")
	}
	
	roleID := idParts[0]
	assigneeID := idParts[1]
	stackID := idParts[3]

	// For stack-level assignments, we need to check via the RBAC API
	// This is a simplified implementation
	assignments, err := client.ListRoleAssignments(ctx, roleID, &ListParams{})
	if err != nil {
		return diag.FromErr(err)
	}

	var assignment *RoleAssignmentResponse
	for _, a := range assignments.Items {
		// For stack assignments, we'd need to check via a different mechanism
		// This is simplified for the example
		if (a.User != nil && a.User.ID == assigneeID) || (a.Team != nil && a.Team.ID == assigneeID) {
			assignment = &a
			break
		}
	}

	if assignment == nil {
		d.SetId("")
		return nil
	}

	d.Set("role_id", assignment.Role.ID)
	d.Set("stack_id", stackID)
	
	if assignment.User != nil {
		d.Set("user_id", assignment.User.ID)
	}
	if assignment.Team != nil {
		d.Set("team_id", assignment.Team.ID)
	}

	return nil
}

func resourceStackRoleAssignmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Role assignments in the real API are typically immutable
	// If any changes occur, we need to delete and recreate
	return diag.Errorf("role assignments cannot be updated - please delete and recreate")
}

func resourceStackRoleAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	// Parse composite ID: role_id:assignee_id:stack:stack_id
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 4 || idParts[2] != "stack" {
		return diag.Errorf("invalid ID format, expected role_id:assignee_id:stack:stack_id")
	}
	
	roleID := idParts[0]
	assigneeID := idParts[1]
	// stackID := idParts[3] // Not currently used but available if needed

	// Find the specific assignment to delete
	assignments, err := client.ListRoleAssignments(ctx, roleID, &ListParams{})
	if err != nil {
		return diag.FromErr(err)
	}

	for _, assignment := range assignments.Items {
		if (assignment.User != nil && assignment.User.ID == assigneeID) || (assignment.Team != nil && assignment.Team.ID == assigneeID) {
			// For the real API, we'd need the specific assignment ID to delete
			// This is a simplified implementation
			err := client.DeleteRoleAssignment(ctx, roleID, assigneeID)
			if err != nil {
				return diag.FromErr(err)
			}
			break
		}
	}

	return nil
}