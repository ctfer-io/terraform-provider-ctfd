package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*fileResource)(nil)
	_ resource.ResourceWithConfigure   = (*fileResource)(nil)
	_ resource.ResourceWithImportState = (*fileResource)(nil)
)

func NewFileResource() resource.Resource {
	return &fileResource{}
}

type fileResource struct {
	client *api.Client
}

type fileResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ChallengeID types.String `tfsdk:"challenge_id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
	SHA1Sum     types.String `tfsdk:"sha1sum"`
	ContentB64  types.String `tfsdk:"contentb64"`
}

func (r *fileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

func (r *fileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A CTFd file for a challenge.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the file, used internally to handle the CTFd corresponding object. WARNING: updating this file does not work, requires full replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"challenge_id": schema.StringAttribute{
				MarkdownDescription: "Challenge of the file.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the file as displayed to end-users.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"location": schema.StringAttribute{
				MarkdownDescription: "Location where the file is stored on the CTFd instance, for download purposes.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sha1sum": schema.StringAttribute{
				MarkdownDescription: "The sha1 sum of the file.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"contentb64": schema.StringAttribute{
				MarkdownDescription: "Base 64 content of the file, perfectly fit the use-cases of complex binaries. You could provide it from the file-system using `filebase64(\"${path.module}/...\")`.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true, // define as sensitive, because content could be + avoid printing it
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *fileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *github.com/ctfer-io/go-ctfd/api.Client, got: %T. Please open an issue at https://github.com/ctfer-io/terraform-provider-ctfd", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *fileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create file
	content, err := base64.StdEncoding.DecodeString(data.ContentB64.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Content Error",
			fmt.Sprintf("base64 content is invalid: %s", err),
		)
		return
	}
	params := &api.PostFilesParams{
		Files: []*api.InputFile{
			{
				Name:    data.Name.ValueString(),
				Content: content,
			},
		},
		Location: data.Location.ValueStringPointer(),
	}
	if !data.ChallengeID.IsNull() {
		params.Challenge = utils.Ptr(utils.Atoi(data.ChallengeID.ValueString()))
	}
	res, err := r.client.PostFiles(params, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create file, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a file")

	// Save computed attributes in state
	data.ID = types.StringValue(strconv.Itoa(res[0].ID))
	data.SHA1Sum = types.StringValue(res[0].SHA1sum)
	data.Location = types.StringValue(res[0].Location)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data fileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetFile(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"CTFd Error",
			fmt.Sprintf("Unable to retrieve file %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}

	data.Name = types.StringValue(filepath.Base(res.Location))
	data.Location = types.StringValue(res.Location)
	data.SHA1Sum = types.StringValue(res.SHA1sum)
	data.ChallengeID = lookForChallengeId(ctx, r.client, res.ID, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	content, err := r.client.GetFileContent(&api.File{
		Location: res.Location,
	}, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"CTFd Error",
			fmt.Sprintf("Unable to read file at location %s, got error: %s", res.Location, err),
		)
		return
	}

	data.ContentB64 = types.StringValue(base64.StdEncoding.EncodeToString(content))

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *fileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data fileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddError("Provider Error", "CTFd does not permit update of file-related information thus this provider cannot do so. This operation should not have been possible.")
}

func (r *fileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data fileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFile(data.ID.ValueString(), api.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete file %s, got error: %s", data.ID.ValueString(), err))
		return
	}
}

func (r *fileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}

// XXX this helper only exist because CTFd does not return the challenge id of a file if it exist...
func lookForChallengeId(ctx context.Context, client *api.Client, fileID int, diags diag.Diagnostics) types.String {
	challs, err := client.GetChallenges(&api.GetChallengesParams{
		View: utils.Ptr("admin"), // required, else CTFd only returns the "visible" challenges
	}, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"CTFd Error",
			fmt.Sprintf("Unable to query challenges, got error: %s", err),
		)
		return types.StringNull()
	}

	for _, chall := range challs {
		files, err := client.GetChallengeFiles(chall.ID, api.WithContext(ctx))
		if err != nil {
			diags.AddError(
				"CTFd Error",
				fmt.Sprintf("Unable to query challenge %d files, got error: %s", chall.ID, err),
			)
			return types.StringNull()
		}
		for _, file := range files {
			if file.ID == fileID {
				return types.StringValue(strconv.Itoa(chall.ID))
			}
		}
	}
	return types.StringNull()
}
