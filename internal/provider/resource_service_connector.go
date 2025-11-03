// resource_service_connector.go
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

var _ resource.Resource = &ServiceConnectorResource{}
var _ resource.ResourceWithImportState = &ServiceConnectorResource{}
var _ resource.ResourceWithConfigValidators = &ServiceConnectorResource{}

func NewServiceConnectorResource() resource.Resource {
	return &ServiceConnectorResource{}
}

type ServiceConnectorResource struct {
	client *Client
}

type ServiceConnectorResourceModel struct {
	ID            types.String   `tfsdk:"id"`
	Name          types.String   `tfsdk:"name"`
	Type          types.String   `tfsdk:"type"`
	AuthMethod    types.String   `tfsdk:"auth_method"`
	ResourceType  types.String   `tfsdk:"resource_type"`
	ResourceID    types.String   `tfsdk:"resource_id"`
	Configuration types.Map      `tfsdk:"configuration"`
	Labels        types.Map      `tfsdk:"labels"`
	ExpiresAt     types.String   `tfsdk:"expires_at"`
	User          types.String   `tfsdk:"user"`
	Created       types.String   `tfsdk:"created"`
	Updated       types.String   `tfsdk:"updated"`
	Verify        types.Bool     `tfsdk:"verify"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

func (r *ServiceConnectorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_connector"
}

func (r *ServiceConnectorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Service connector resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service connector identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the service connector",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the service connector",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(validConnectorTypes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auth_method": schema.StringAttribute{
				MarkdownDescription: "Authentication method for the service connector",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type": schema.StringAttribute{
				MarkdownDescription: "Resource type for the service connector",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_id": schema.StringAttribute{
				MarkdownDescription: "Resource ID for the service connector",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"configuration": schema.MapAttribute{
				MarkdownDescription: "Configuration for the service connector",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Labels for the service connector",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "Expiration time for the service connector (RFC3339 format)",
				Computed:            true,
			},
			"verify": schema.BoolAttribute{
				MarkdownDescription: "Whether to verify the service connector configuration before creating or updating it",
				Optional:            true,
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "The owner of the service connector",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the service connector was created",
				Computed:            true,
			},
			"updated": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the service connector was last updated",
				Computed:            true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *ServiceConnectorResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&serviceConnectorConfigValidator{},
	}
}

type serviceConnectorConfigValidator struct{}

func (v serviceConnectorConfigValidator) Description(ctx context.Context) string {
	return "Validates service connector configuration"
}

func (v serviceConnectorConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates service connector configuration"
}

func (v serviceConnectorConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ServiceConnectorResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	connectorType := data.Type.ValueString()
	authMethod := data.AuthMethod.ValueString()

	if !data.Type.IsUnknown() && connectorType != "" {
		validType := false
		for _, t := range validConnectorTypes {
			if t == connectorType {
				validType = true
				break
			}
		}
		if !validType {
			resp.Diagnostics.AddAttributeError(
				path.Root("type"),
				"Invalid connector type",
				fmt.Sprintf("Invalid connector type %q. Valid types are: %s",
					connectorType, strings.Join(validConnectorTypes, ", ")),
			)
			return
		}
	}

	if !data.AuthMethod.IsUnknown() && authMethod != "" {
		if methods, ok := validAuthMethods[connectorType]; ok {
			validMethod := false
			for _, m := range methods {
				if m == authMethod {
					validMethod = true
					break
				}
			}
			if !validMethod {
				resp.Diagnostics.AddAttributeError(
					path.Root("auth_method"),
					"Invalid auth method",
					fmt.Sprintf("Invalid auth_method %q for connector type %q. Valid methods are: %s",
						authMethod, connectorType,
						strings.Join(validAuthMethods[connectorType], ", ")),
				)
			}
		}
	}

	if !data.ResourceType.IsNull() && data.ResourceType.ValueString() != "" {
		validTypes := validResourceTypes[connectorType]
		resourceType := data.ResourceType.ValueString()
		valid := false
		for _, t := range validTypes {
			if t == resourceType {
				valid = true
				break
			}
		}
		if !valid {
			resp.Diagnostics.AddAttributeError(
				path.Root("resource_type"),
				"Invalid resource type",
				fmt.Sprintf("Invalid resource type %q for connector type %q. Valid types are: %s",
					resourceType, connectorType, strings.Join(validTypes, ", ")),
			)
		}
	}

	// NOTE: we intentionally omit validating the configuration here
	// for two reasons:
	// 1. The configuration can be derived from resources and data
	//    sources that are not available during plan time.
	// 2. The configuration are validated by the ZenML server
	//    when the connector is validated / created and we don't want to
	//    duplicate that logic here.
}

func (r *ServiceConnectorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ServiceConnectorResource) buildServiceConnectorRequest(
	ctx context.Context,
	data *ServiceConnectorResourceModel,
	diags *diag.Diagnostics,
) *ServiceConnectorRequest {
	user, err := r.client.GetCurrentUser(ctx)
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get current user, got error: %s", err),
		)
		return nil
	}

	configuration := make(map[string]interface{})
	if !data.Configuration.IsNull() {
		configElements := make(map[string]types.String, len(data.Configuration.Elements()))
		diags.Append(data.Configuration.ElementsAs(ctx, &configElements, false)...)
		if diags.HasError() {
			return nil
		}

		for k, v := range configElements {
			configuration[k] = v.ValueString()
		}
	}

	labels := make(map[string]string)
	if !data.Labels.IsNull() {
		labelElements := make(map[string]types.String, len(data.Labels.Elements()))
		diags.Append(data.Labels.ElementsAs(ctx, &labelElements, false)...)
		if diags.HasError() {
			return nil
		}

		for k, v := range labelElements {
			labels[k] = v.ValueString()
		}
	}

	resourceTypes := []string{}
	if !data.ResourceType.IsNull() && data.ResourceType.ValueString() != "" {
		resourceTypes = []string{data.ResourceType.ValueString()}
	}

	connectorReq := &ServiceConnectorRequest{
		User:          user.ID,
		Name:          data.Name.ValueString(),
		ConnectorType: data.Type.ValueString(),
		AuthMethod:    data.AuthMethod.ValueString(),
		ResourceTypes: resourceTypes,
		Configuration: configuration,
		Labels:        labels,
	}

	if !data.ResourceID.IsNull() && data.ResourceID.ValueString() != "" {
		resourceID := data.ResourceID.ValueString()
		connectorReq.ResourceID = &resourceID
	}

	return connectorReq
}

func (r *ServiceConnectorResource) populateServiceConnectorModel(
	ctx context.Context,
	connector *ServiceConnectorResponse,
	data *ServiceConnectorResourceModel,
	diags *diag.Diagnostics,
	updateConfiguration bool,
) {
	data.ID = types.StringValue(connector.ID)
	data.Name = types.StringValue(connector.Name)

	var oldUpdated string
	if connector.Body != nil {
		var connectorType string
		if err := json.Unmarshal(connector.Body.ConnectorType, &connectorType); err != nil {
			var connectorTypeObj struct {
				ConnectorType string `json:"connector_type"`
			}
			if err := json.Unmarshal(connector.Body.ConnectorType, &connectorTypeObj); err == nil {
				connectorType = connectorTypeObj.ConnectorType
			}
		}
		if connectorType != "" {
			data.Type = types.StringValue(connectorType)
		}

		data.AuthMethod = types.StringValue(connector.Body.AuthMethod)

		if len(connector.Body.ResourceTypes) == 1 {
			data.ResourceType = types.StringValue(connector.Body.ResourceTypes[0])
		}

		if connector.Body.ResourceID != nil {
			data.ResourceID = types.StringValue(*connector.Body.ResourceID)
		}

		if connector.Body.ExpiresAt != nil {
			data.ExpiresAt = types.StringValue(*connector.Body.ExpiresAt)
		} else {
			data.ExpiresAt = types.StringNull()
		}

		if connector.Body.User != nil {
			data.User = types.StringValue(connector.Body.User.ID)
		} else {
			data.User = types.StringNull()
		}

		data.Created = types.StringValue(connector.Body.Created)

		if !data.Updated.IsNull() {
			oldUpdated = data.Updated.ValueString()
		}
		data.Updated = types.StringValue(connector.Body.Updated)
	}

	if connector.Metadata != nil {
		if connector.Metadata.Configuration != nil {
			timestampUnchanged := oldUpdated != "" &&
				oldUpdated == connector.Body.Updated

			if updateConfiguration && !timestampUnchanged {
				cfg, changed := MergeOrCompareConfiguration(
					ctx,
					data.Configuration,
					connector.Metadata.Configuration,
					diags,
				)
				if !diags.HasError() && changed {
					data.Configuration = cfg
				}
			}
		}

		if connector.Metadata.Labels != nil {
			labelMap := make(map[string]attr.Value)
			for k, v := range connector.Metadata.Labels {
				labelMap[k] = types.StringValue(v)
			}
			labelValue, labelDiags := types.MapValue(types.StringType, labelMap)
			diags.Append(labelDiags...)
			if !diags.HasError() {
				data.Labels = labelValue
			}
		}
	}
}

func (r *ServiceConnectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceConnectorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve timeout for verification.
	createTimeout, tDiags := data.Timeouts.Create(ctx, 5*time.Minute)
	resp.Diagnostics.Append(tDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectorReq := r.buildServiceConnectorRequest(ctx, &data, &resp.Diagnostics)
	if connectorReq == nil {
		return
	}

	verify := true
	if !data.Verify.IsNull() {
		verify = data.Verify.ValueBool()
	}

	if verify {
		retryErr := retry.RetryContext(ctx, createTimeout, func() *retry.RetryError {
			validation, err := r.client.VerifyServiceConnector(ctx, *connectorReq)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("unable to verify service connector configuration, got error: %s", err))
			}
			if validation.Error != nil {
				return retry.RetryableError(fmt.Errorf("Error verifying service connector configuration: %s", *validation.Error))
			}
			return nil
		})
		if retryErr != nil {
			resp.Diagnostics.AddError("Verification Error", retryErr.Error())
			return
		}
	}

	tflog.Trace(ctx, "creating service connector")

	connector, err := r.client.CreateServiceConnector(ctx, *connectorReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create service connector, got error: %s", err))
		return
	}

	r.populateServiceConnectorModel(ctx, connector, &data, &resp.Diagnostics, false)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a service connector")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceConnectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceConnectorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	connector, err := r.client.GetServiceConnector(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read service connector, got error: %s", err))
		return
	}

	if connector == nil {
		// Connector was deleted outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	r.populateServiceConnectorModel(ctx, connector, &data, &resp.Diagnostics, true)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceConnectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServiceConnectorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve timeout for verification.
	updateTimeout, tDiags := data.Timeouts.Update(ctx, 5*time.Minute)
	resp.Diagnostics.Append(tDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	connectorReq := r.buildServiceConnectorRequest(ctx, &data, &resp.Diagnostics)
	if connectorReq == nil {
		return
	}

	verify := true
	if !data.Verify.IsNull() {
		verify = data.Verify.ValueBool()
	}

	if verify {
		retryErr := retry.RetryContext(ctx, updateTimeout, func() *retry.RetryError {
			validation, err := r.client.VerifyServiceConnector(ctx, *connectorReq)
			if err != nil {
				return retry.NonRetryableError(fmt.Errorf("unable to verify service connector configuration, got error: %s", err))
			}
			if validation.Error != nil {
				return retry.RetryableError(fmt.Errorf("Error verifying service connector configuration: %s", *validation.Error))
			}
			return nil
		})
		if retryErr != nil {
			resp.Diagnostics.AddError("Verification Error", retryErr.Error())
			return
		}
	}

	updateReq := ServiceConnectorUpdate{
		Configuration: connectorReq.Configuration,
		Labels:        connectorReq.Labels,
		ResourceTypes: connectorReq.ResourceTypes,
	}

	if !data.Name.IsNull() {
		name := data.Name.ValueString()
		updateReq.Name = &name
	}

	if connectorReq.ResourceID != nil {
		updateReq.ResourceID = connectorReq.ResourceID
	}

	tflog.Trace(ctx, "updating service connector")

	connector, err := r.client.UpdateServiceConnector(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update service connector, got error: %s", err))
		return
	}

	r.populateServiceConnectorModel(ctx, connector, &data, &resp.Diagnostics, true)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceConnectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceConnectorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "deleting service connector")

	err := r.client.DeleteServiceConnector(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service connector, got error: %s", err))
		return
	}
}

func (r *ServiceConnectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
