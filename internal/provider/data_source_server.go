package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServer() *schema.Resource {
	return &schema.Resource{
		Description: "Data source for global ZenML server information",
		ReadContext: dataSourceServerRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Server name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"version": {
				Description: "Server version",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"deployment_type": {
				Description: "Server deployment type",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"auth_scheme": {
				Description: "Server authentication scheme",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"server_url": {
				Description: "Server API URL",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"dashboard_url": {
				Description: "Server dashboard URL",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"pro_dashboard_url": {
				Description: "ZenML Pro dashboard URL",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"pro_api_url": {
				Description: "ZenML Pro API URL",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"pro_organization_id": {
				Description: "ZenML Pro organization ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"pro_organization_name": {
				Description: "ZenML Pro organization name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"pro_workspace_id": {
				Description: "ZenML Pro workspace ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"pro_workspace_name": {
				Description: "ZenML Pro workspace name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"metadata": {
				Description: "Server metadata",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	server, err := c.GetServerInfo(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error fetching server info: %v", err))
	}

	d.SetId(server.ID)

	if err := d.Set("name", server.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("version", server.Version); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("deployment_type", server.DeploymentType); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("auth_scheme", server.AuthScheme); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("server_url", server.ServerURL); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("dashboard_url", server.DashboardURL); err != nil {
		return diag.FromErr(err)
	}

	if server.ProDashboardURL != nil {
		if err := d.Set("pro_dashboard_url", *server.ProDashboardURL); err != nil {
			return diag.FromErr(err)
		}
	}

	if server.ProAPIURL != nil {
		if err := d.Set("pro_api_url", *server.ProAPIURL); err != nil {
			return diag.FromErr(err)
		}
	}

	if server.ProOrganizationID != nil {
		if err := d.Set("pro_organization_id", *server.ProOrganizationID); err != nil {
			return diag.FromErr(err)
		}
	}

	if server.ProOrganizationName != nil {
		if err := d.Set("pro_organization_name", *server.ProOrganizationName); err != nil {
			return diag.FromErr(err)
		}
	}

	if server.ProWorkspaceID != nil {
		if err := d.Set("pro_workspace_id", *server.ProWorkspaceID); err != nil {
			return diag.FromErr(err)
		}
	}

	if server.ProWorkspaceName != nil {
		if err := d.Set("pro_workspace_name", *server.ProWorkspaceName); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("metadata", server.Metadata); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
