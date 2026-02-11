package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
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
	_ resource.Resource                = (*bracketResource)(nil)
	_ resource.ResourceWithConfigure   = (*bracketResource)(nil)
	_ resource.ResourceWithImportState = (*bracketResource)(nil)
)

func NewBracketResource() resource.Resource {
	return &bracketResource{}
}

type bracketResource struct {
	fm *Framework
}

type bracketResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
}

func (r *bracketResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bracket"
}

func (r *bracketResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A bracket for users or teams to compete in parallel.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the bracket, used internally to handle the CTFd corresponding object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name displayed to end-users (e.g. \"Students\", \"Interns\", \"Engineers\").",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description that explains the goal of this bracket.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the bracket, either \"users\" or \"teams\".",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("users"),
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						types.StringValue("users"),
						types.StringValue("teams"),
					}),
				},
			},
		},
	}
}

func (r *bracketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	fm, ok := req.ProviderData.(*Framework)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected %T, got: %T. Please open an issue at https://github.com/ctfer-io/terraform-provider-ctfd", (*Framework)(nil), req.ProviderData),
		)
		return
	}

	r.fm = fm
}

func (r *bracketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data bracketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create bracket
	res, err := r.fm.Client.PostBrackets(ctx, &api.PostBracketsParams{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Type:        data.Type.ValueString(),
	}, WithTracerProvider(r.fm.Tp))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create bracket, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a bracket")

	// Save computed attributes in state
	data.ID = types.StringValue(strconv.Itoa(res.ID))

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bracketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data bracketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// XXX cannot get bracket by ID, so we need to query them all
	brackets, err := r.fm.Client.GetBrackets(ctx, &api.GetBracketsParams{}, WithTracerProvider(r.fm.Tp))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get bracket %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	var bkt *api.Bracket
	for _, bracket := range brackets {
		if data.ID.ValueString() == strconv.Itoa(bracket.ID) {
			bkt = bracket
			break
		}
	}
	if bkt == nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get bracket %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Upsert values
	data.Name = types.StringValue(bkt.Name)
	data.Description = types.StringValue(bkt.Description)
	data.Type = types.StringValue(bkt.Type)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bracketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data bracketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update bracket
	if _, err := r.fm.Client.PatchBrackets(ctx, data.ID.ValueString(), &api.PatchBracketsParams{
		Name:        data.Name.ValueStringPointer(),
		Description: data.Description.ValueStringPointer(),
		Type:        data.Type.ValueStringPointer(),
	}, WithTracerProvider(r.fm.Tp)); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update bracket %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *bracketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data bracketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.fm.Client.DeleteBrackets(ctx, data.ID.ValueString(), WithTracerProvider(r.fm.Tp)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete bracket %s, got error: %s", data.ID.ValueString(), err))
		return
	}
}

func (r *bracketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}
