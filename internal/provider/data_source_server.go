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

	server, err := c.GetServerInfo()
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

	if err := d.Set("metadata", server.Metadata); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
