package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceWorkspaceRoleAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkspaceRoleAssignmentCreate,
		ReadContext:   resourceWorkspaceRoleAssignmentRead,
		UpdateContext: resourceWorkspaceRoleAssignmentUpdate,
		DeleteContext: resourceWorkspaceRoleAssignmentDelete,
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
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the workspace",
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

func resourceWorkspaceRoleAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	roleID := d.Get("role_id").(string)
	workspaceID := d.Get("workspace_id").(string)
	
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
		WorkspaceID: &workspaceID,
		UserID:      userID,
		TeamID:      teamID,
	}

	_, err := client.CreateRoleAssignment(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create a composite ID: role_id:assignee_id:workspace_id
	var assigneeID string
	if userID != nil {
		assigneeID = *userID
	} else {
		assigneeID = *teamID
	}
	d.SetId(roleID + ":" + assigneeID + ":" + workspaceID)

	return resourceWorkspaceRoleAssignmentRead(ctx, d, meta)
}

func resourceWorkspaceRoleAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	// Parse composite ID: role_id:assignee_id:workspace_id
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 3 {
		return diag.Errorf("invalid ID format, expected role_id:assignee_id:workspace_id")
	}
	
	roleID := idParts[0]
	assigneeID := idParts[1]
	workspaceID := idParts[2]

	// For the real API, we would need to list role assignments and find the matching one
	// This is a simplified implementation
	assignments, err := client.ListRoleAssignments(ctx, roleID, &ListParams{})
	if err != nil {
		return diag.FromErr(err)
	}

	var assignment *RoleAssignmentResponse
	for _, a := range assignments.Items {
		if a.WorkspaceID != nil && *a.WorkspaceID == workspaceID {
			if (a.User != nil && a.User.ID == assigneeID) || (a.Team != nil && a.Team.ID == assigneeID) {
				assignment = &a
				break
			}
		}
	}

	if assignment == nil {
		d.SetId("")
		return nil
	}

	d.Set("role_id", assignment.Role.ID)
	d.Set("workspace_id", assignment.WorkspaceID)
	
	if assignment.User != nil {
		d.Set("user_id", assignment.User.ID)
	}
	if assignment.Team != nil {
		d.Set("team_id", assignment.Team.ID)
	}

	return nil
}

func resourceWorkspaceRoleAssignmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Role assignments in the real API are typically immutable
	// If any changes occur, we need to delete and recreate
	return diag.Errorf("role assignments cannot be updated - please delete and recreate")
}

func resourceWorkspaceRoleAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client)

	// Parse composite ID: role_id:assignee_id:workspace_id
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 3 {
		return diag.Errorf("invalid ID format, expected role_id:assignee_id:workspace_id")
	}
	
	roleID := idParts[0]
	assigneeID := idParts[1]
	workspaceID := idParts[2]

	// Find the specific assignment to delete
	assignments, err := client.ListRoleAssignments(ctx, roleID, &ListParams{})
	if err != nil {
		return diag.FromErr(err)
	}

	for _, assignment := range assignments.Items {
		if assignment.WorkspaceID != nil && *assignment.WorkspaceID == workspaceID {
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
	}

	return nil
}