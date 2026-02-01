package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/utils"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*flagResource)(nil)
	_ resource.ResourceWithConfigure   = (*flagResource)(nil)
	_ resource.ResourceWithImportState = (*flagResource)(nil)
)

func NewFlagResource() resource.Resource {
	return &flagResource{}
}

type flagResource struct {
	client *Client
}

type flagResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ChallengeID types.String `tfsdk:"challenge_id"`
	Content     types.String `tfsdk:"content"`
	Data        types.String `tfsdk:"data"`
	Type        types.String `tfsdk:"type"`
}

func (r *flagResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flag"
}

func (r *flagResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A flag to solve the challenge.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the flag, used internally to handle the CTFd corresponding object.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"challenge_id": schema.StringAttribute{
				MarkdownDescription: "Challenge of the flag.",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The actual flag to match. Consider using the convention `MYCTF{value}` with `MYCTF` being the shortcode of your event's name and `value` depending on each challenge.",
				Required:            true,
				Sensitive:           true,
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "The flag sensitivity information, either case_sensitive or case_insensitive",
				Optional:            true,
				Computed:            true,
				// default value is "" (empty string) according to Web UI
				Default: stringdefault.StaticString("case_sensitive"),
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						types.StringValue("case_sensitive"),
						types.StringValue("case_insensitive"),
					}),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the flag, could be either static or regex",
				Optional:            true,
				Computed:            true,
				// default value is "static" according to ctfcli
				Default: stringdefault.StaticString("static"),
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						types.StringValue("static"),
						types.StringValue("regex"),
					}),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *flagResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *github.com/ctfer-io/go-ctfd/api.Client, got: %T. Please open an issue at https://github.com/ctfer-io/terraform-provider-ctfd", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *flagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data flagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create flag
	res, err := r.client.PostFlags(ctx, &api.PostFlagsParams{
		Challenge: utils.Atoi(data.ChallengeID.ValueString()),
		Content:   data.Content.ValueString(),
		Data:      data.Data.ValueString(),
		Type:      data.Type.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create flag, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a flag")

	// Save computed attributes in state
	data.ID = types.StringValue(strconv.Itoa(res.ID))

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *flagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data flagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve flag
	res, err := r.client.GetFlag(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read flag %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Upsert values
	data.ChallengeID = types.StringValue(strconv.Itoa(res.ChallengeID))
	data.Content = types.StringValue(res.Content)
	data.Data = types.StringValue(res.Data)
	data.Type = types.StringValue(res.Type)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *flagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data flagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update flag
	if _, err := r.client.PatchFlag(ctx, data.ID.ValueString(), &api.PatchFlagParams{
		ID:      data.ID.ValueString(),
		Content: data.Content.ValueString(),
		Data:    data.Data.ValueString(),
		Type:    data.Type.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update flag %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *flagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data flagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFlag(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete flag %s, got error: %s", data.ID.ValueString(), err))
		return
	}
}

func (r *flagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}
