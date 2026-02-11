package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*hintResource)(nil)
	_ resource.ResourceWithConfigure   = (*hintResource)(nil)
	_ resource.ResourceWithImportState = (*hintResource)(nil)
)

func NewHintResource() resource.Resource {
	return &hintResource{}
}

type hintResource struct {
	fm *Framework
}

type hintResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	ChallengeID  types.String   `tfsdk:"challenge_id"`
	Title        types.String   `tfsdk:"title"`
	Content      types.String   `tfsdk:"content"`
	Cost         types.Int64    `tfsdk:"cost"`
	Requirements []types.String `tfsdk:"requirements"`
}

func (r *hintResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hint"
}

func (r *hintResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A hint for a challenge to help players solve it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the hint, used internally to handle the CTFd corresponding object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"challenge_id": schema.StringAttribute{
				MarkdownDescription: "Challenge of the hint.",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the hint, displayed to end users before unlocking.",
				Optional:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "Content of the hint as displayed to the end-user.",
				Required:            true,
			},
			"cost": schema.Int64Attribute{
				MarkdownDescription: "Cost of the hint, and if any specified, the end-user will consume its own (or team) points to get it.",
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(0),
			},
			"requirements": schema.SetAttribute{
				MarkdownDescription: "List of the other hints it depends on.",
				ElementType:         types.StringType,
				Computed:            true,
				Optional:            true,
				Default:             setdefault.StaticValue(basetypes.NewSetValueMust(types.StringType, []attr.Value{})),
			},
		},
	}
}

func (r *hintResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *hintResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data hintResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create hint
	reqs := make([]int, 0, len(data.Requirements))
	for _, preq := range data.Requirements {
		id, _ := strconv.Atoi(preq.ValueString())
		reqs = append(reqs, id)
	}
	res, err := r.fm.Client.PostHints(ctx, &api.PostHintsParams{
		ChallengeID: utils.Atoi(data.ChallengeID.ValueString()),
		Title:       data.Title.ValueStringPointer(),
		Content:     data.Content.ValueString(),
		Cost:        int(data.Cost.ValueInt64()),
		Requirements: api.Requirements{
			Prerequisites: reqs,
		},
	}, WithTracerProvider(r.fm.Tp))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create hint, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a hint")

	// Save computed attributes in state
	data.ID = types.StringValue(strconv.Itoa(res.ID))

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hintResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data hintResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve hint
	h, err := r.fm.Client.GetHint(ctx, data.ID.ValueString(), &api.GetHintParams{
		Preview: utils.Ptr(true), // mimic a preview to get the hint even if not unlocked by the admin
	}, WithTracerProvider(r.fm.Tp))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get hint %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	// XXX cannot get hint by ID, so we need to query them all
	hints, err := r.fm.Client.GetChallengeHints(ctx, strconv.Itoa(h.ChallengeID), WithTracerProvider(r.fm.Tp))
	hint := (*api.Hint)(nil)
	for _, h := range hints {
		if h.ID == utils.Atoi(data.ID.ValueString()) {
			hint = h
			break
		}
	}
	if hint == nil {
		resp.Diagnostics.AddError(
			"CTFd Error",
			fmt.Sprintf("Unable to get hint %s of challenge %s, got error: %s", data.ID.ValueString(), data.ChallengeID.ValueString(), err),
		)
		return
	}

	// Upsert values
	data.ChallengeID = types.StringValue(strconv.Itoa(h.ChallengeID))
	data.Title = types.StringPointerValue(hint.Title)
	data.Content = types.StringValue(*hint.Content)
	data.Cost = types.Int64Value(int64(hint.Cost))
	reqs := make([]basetypes.StringValue, 0, len(hint.Requirements.Prerequisites))
	for _, preq := range hint.Requirements.Prerequisites {
		reqs = append(reqs, types.StringValue(strconv.Itoa(preq)))
	}
	data.Requirements = reqs

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hintResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data hintResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update hint
	preqs := make([]int, 0, len(data.Requirements))
	for _, preq := range data.Requirements {
		id, _ := strconv.Atoi(preq.ValueString())
		preqs = append(preqs, id)
	}
	if _, err := r.fm.Client.PatchHint(ctx, data.ID.ValueString(), &api.PatchHintsParams{
		ChallengeID: utils.Atoi(data.ChallengeID.ValueString()),
		Title:       data.Title.ValueStringPointer(),
		Content:     data.Content.ValueString(),
		Cost:        int(data.Cost.ValueInt64()),
		Requirements: api.Requirements{
			Prerequisites: preqs,
		},
	}, WithTracerProvider(r.fm.Tp)); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update hint %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *hintResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data hintResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.fm.Client.DeleteHint(ctx, data.ID.ValueString(), WithTracerProvider(r.fm.Tp)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete hint %s, got error: %s", data.ID.ValueString(), err))
		return
	}
}

func (r *hintResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}
