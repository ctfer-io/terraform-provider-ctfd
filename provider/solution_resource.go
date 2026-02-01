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
	_ resource.Resource                = (*solutionResource)(nil)
	_ resource.ResourceWithConfigure   = (*solutionResource)(nil)
	_ resource.ResourceWithImportState = (*solutionResource)(nil)
)

func NewSolutionResource() resource.Resource {
	return &solutionResource{}
}

type solutionResource struct {
	client *Client
}

type solutionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ChallengeID types.String `tfsdk:"challenge_id"`
	Content     types.String `tfsdk:"content"`
	State       types.String `tfsdk:"state"`
}

func (r *solutionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_solution"
}

func (r *solutionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The solution to a challenge.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the solution, used internally to handle the CTFd corresponding object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"challenge_id": schema.StringAttribute{
				MarkdownDescription: "Challenge of the solution.",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The solution to the challenge, in markdown.",
				Optional:            true,
				Sensitive:           true, // if leaked, is close to leaking the flag directly
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "State of the solution, either hidden or visible.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("hidden"),
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						types.StringValue("hidden"),
						types.StringValue("visible"),
					}),
				},
			},
		},
	}
}

func (r *solutionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *solutionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data solutionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create solution
	res, err := r.client.PostSolutions(ctx, &api.PostSolutionsParams{
		ChallengeID: utils.Atoi(data.ChallengeID.ValueString()),
		Content:     data.Content.ValueString(),
		State:       data.State.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create solution of challenge %s, got error: %s", data.ChallengeID.ValueString(), err),
		)
		return
	}

	tflog.Trace(ctx, "created a solution")

	// Save computed attributes in state
	data.ID = types.StringValue(strconv.Itoa(res.ID))

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *solutionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data solutionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve solution
	res, err := r.client.GetSolutions(ctx, data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read solution of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}

	// Upsert values
	data.ChallengeID = types.StringValue(strconv.Itoa(res.ChallengeID))
	data.Content = types.StringValue(res.Content)
	data.State = types.StringValue(res.State)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *solutionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data solutionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update solution
	if _, err := r.client.PatchSolutions(ctx, data.ID.ValueString(), &api.PatchSolutionsParams{
		Content: data.Content.ValueString(),
		State:   data.State.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update solution of challenge %s, got error: %s", data.ChallengeID.ValueString(), err),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *solutionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data solutionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSolutions(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete solution of challenge %s, got error: %s", data.ChallengeID.ValueString(), err))
		return
	}
}

func (r *solutionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}
