package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &SecretResource{}
var _ resource.ResourceWithImportState = &SecretResource{}

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

type SecretResource struct {
	client *Client
}

type SecretResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Private types.Bool   `tfsdk:"private"`
	Values  types.Map    `tfsdk:"values"`
	UserID  types.String `tfsdk:"user_id"`
	Created types.String `tfsdk:"created"`
	Updated types.String `tfsdk:"updated"`
}

func (r *SecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a ZenML secret. Secret values are stored in Terraform state; use a secured remote backend.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Secret identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the secret",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"private": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether the secret is only accessible to the user that created it",
				Default:             booldefault.StaticBool(false),
			},
			// Sensitive=true hides the value from normal terraform output
			"values": schema.MapAttribute{
				Required:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
				MarkdownDescription: "Key-value pairs stored in the secret",
			},
			"user_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the user that owns the secret",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the secret was created",
			},
			"updated": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the secret was last updated",
			},
		},
	}
}

func (r *SecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func secretValuesFromModel(ctx context.Context, data *SecretResourceModel, diags *diag.Diagnostics) map[string]string {
	values := map[string]string{}
	diags.Append(data.Values.ElementsAs(ctx, &values, false)...)
	return values
}

func (r *SecretResource) populateSecretModel(
	ctx context.Context,
	secret *SecretResponse,
	data *SecretResourceModel,
	diags *diag.Diagnostics,
) {
	if secret.Body == nil {
		diags.AddError("Invalid API Response", "The ZenML server returned a secret without a response body.")
		return
	}

	values := make(map[string]string, len(secret.Body.Values))
	for key, value := range secret.Body.Values {
		if value == nil {
			diags.AddAttributeError(
				path.Root("values"),
				"Unable to Read Secret Values",
				"ZenML did not return all values for this secret. Ensure the account configured for the provider can read secret values.",
			)
			return
		}
		values[key] = *value
	}

	valueMap, valueDiags := types.MapValueFrom(ctx, types.StringType, values)
	diags.Append(valueDiags...)
	if diags.HasError() {
		return
	}

	data.ID = types.StringValue(secret.ID)
	data.Name = types.StringValue(secret.Name)
	data.Private = types.BoolValue(secret.Body.Private)
	data.Values = valueMap
	data.Created = types.StringValue(secret.Body.Created)
	data.Updated = types.StringValue(secret.Body.Updated)
	if secret.Body.UserID == nil {
		data.UserID = types.StringNull()
	} else {
		data.UserID = types.StringValue(*secret.Body.UserID)
	}
}

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := secretValuesFromModel(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "creating secret")
	secret, err := r.client.CreateSecret(ctx, SecretRequest{
		Name:    data.Name.ValueString(),
		Private: data.Private.ValueBool(),
		Values:  values,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create secret, got error: %s", err))
		return
	}

	r.populateSecretModel(ctx, secret, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GetSecret(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read secret, got error: %s", err))
		return
	}
	// if the secret is not found, the resource was likely deleted outside of Terraform
	// Terraform will propose to recreate it.
	if secret == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.populateSecretModel(ctx, secret, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	values := secretValuesFromModel(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updating secret")
	secret, err := r.client.UpdateSecret(ctx, data.ID.ValueString(), SecretUpdate{
		Name:    data.Name.ValueString(),
		Private: data.Private.ValueBool(),
		Values:  values,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update secret, got error: %s", err))
		return
	}

	r.populateSecretModel(ctx, secret, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "deleting secret")
	if err := r.client.DeleteSecret(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete secret, got error: %s", err))
	}
}

func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
