package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = (*teamResource)(nil)
	_ resource.ResourceWithConfigure   = (*teamResource)(nil)
	_ resource.ResourceWithImportState = (*teamResource)(nil)
)

type teamResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	Email       types.String   `tfsdk:"email"`
	Password    types.String   `tfsdk:"password"`
	Website     types.String   `tfsdk:"website"`
	Affiliation types.String   `tfsdk:"affiliation"`
	Country     types.String   `tfsdk:"country"`
	Hidden      types.Bool     `tfsdk:"hidden"`
	Banned      types.Bool     `tfsdk:"banned"`
	Members     []types.String `tfsdk:"members"`
	Captain     types.String   `tfsdk:"captain"`
}

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

type teamResource struct {
	client *api.Client
}

func (r *teamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CTFd defines a Team as a group of Users who will attend the Capture The Flag event.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the user.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the team.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email of the team.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password of the team. Notice that during a CTF you may not want to update those to avoid defaulting team accesses.",
				Required:            true,
			},
			"website": schema.StringAttribute{
				MarkdownDescription: "Website, blog, or anything similar (displayed to other participants).",
				Optional:            true,
			},
			"affiliation": schema.StringAttribute{
				MarkdownDescription: "Affiliation to a company or agency.",
				Optional:            true,
			},
			"country": schema.StringAttribute{
				MarkdownDescription: "Country the team represent or is hail from.",
				Optional:            true,
			},
			"hidden": schema.BoolAttribute{
				MarkdownDescription: "Is true if the team is hidden to the participants.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.Bool(booldefault.StaticBool(false)),
			},
			"banned": schema.BoolAttribute{
				MarkdownDescription: "Is true if the team is banned from the CTF.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.Bool(booldefault.StaticBool(false)),
			},
			"members": schema.ListAttribute{
				MarkdownDescription: "List of members (User), defined by their IDs.",
				ElementType:         types.StringType,
				Required:            true,
			},
			"captain": schema.StringAttribute{
				MarkdownDescription: "Member who is captain of the team. Must be part of the members too.",
				Required:            true,
			},
		},
	}
}

func (r *teamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.PostTeams(&api.PostTeamsParams{
		Name:        data.Name.ValueString(),
		Email:       data.Email.ValueString(),
		Password:    data.Password.ValueString(),
		Website:     data.Website.ValueStringPointer(),
		Affiliation: data.Affiliation.ValueStringPointer(),
		Country:     data.Country.ValueStringPointer(),
		Hidden:      data.Hidden.ValueBool(),
		Banned:      data.Banned.ValueBool(),
		Fields:      []api.Field{},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create team, got error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(strconv.Itoa(res.ID))

	// => Members
	for _, mem := range data.Members {
		_, err := r.client.PostTeamMembers(res.ID, &api.PostTeamsMembersParams{
			UserID: utils.Atoi(mem.ValueString()),
		}, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to add user to team %d, got error: %s", res.ID, err),
			)
			return
		}
	}
	// => Captain
	cap := utils.Atoi(data.Captain.ValueString())
	if _, err := r.client.PatchTeam(res.ID, &api.PatchTeamsParams{
		CaptainID: &cap,
		Fields:    []api.Field{},
	}); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to set user %d as team %d captain, got error: %s", cap, res.ID, err),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := utils.Atoi(data.ID.ValueString())
	res, err := r.client.GetTeam(teamId, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read team %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}

	data.Name = types.StringValue(res.Name)
	data.Email = types.StringPointerValue(res.Email)
	data.Website = types.StringPointerValue(res.Website)
	data.Affiliation = types.StringPointerValue(res.Affiliation)
	data.Country = types.StringPointerValue(res.Country)
	data.Hidden = types.BoolValue(res.Hidden)
	data.Banned = types.BoolValue(res.Banned)
	// password is not returned, which is good :)

	// => Members
	mems, err := r.client.GetTeamMembers(teamId, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read team %s members, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	data.Members = make([]basetypes.StringValue, 0, len(mems))
	for _, mem := range mems {
		data.Members = append(data.Members, types.StringValue(strconv.Itoa(mem)))
	}
	// => Captain
	data.Captain = types.StringValue(strconv.Itoa(*res.CaptainID))

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamId := utils.Atoi(data.ID.ValueString())
	_, err := r.client.PatchTeam(teamId, &api.PatchTeamsParams{
		Name:        data.Name.ValueStringPointer(),
		Email:       data.Email.ValueStringPointer(),
		Password:    data.Password.ValueStringPointer(),
		Website:     data.Website.ValueStringPointer(),
		Affiliation: data.Affiliation.ValueStringPointer(),
		Country:     data.Country.ValueStringPointer(),
		Hidden:      data.Hidden.ValueBoolPointer(),
		Banned:      data.Banned.ValueBoolPointer(),
		Fields:      []api.Field{},
	}, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update team, got error: %s", err),
		)
		return
	}

	// => Members
	currentMembers, err := r.client.GetTeamMembers(teamId, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get team's %d members, got error: %s", teamId, err),
		)
		return
	}
	members := []basetypes.StringValue{}
	for _, tfMember := range data.Members {
		exists := false
		for _, currentMember := range currentMembers {
			if tfMember.ValueString() == strconv.Itoa(currentMember) {
				exists = true
				break
			}
		}
		if !exists {
			if _, err := r.client.PostTeamMembers(teamId, &api.PostTeamsMembersParams{
				UserID: utils.Atoi(tfMember.ValueString()),
			}, api.WithContext(ctx)); err != nil {
				resp.Diagnostics.AddError(
					"Client Error",
					fmt.Sprintf("Unable to post team's %d member %s, got error: %s", teamId, tfMember.ValueString(), err),
				)
				return
			}
			members = append(members, tfMember)
		}
	}
	for _, currentMember := range currentMembers {
		exists := false
		for _, tfMember := range data.Members {
			if tfMember.ValueString() == strconv.Itoa(currentMember) {
				exists = true
				members = append(members, tfMember)
				break
			}
		}
		if !exists {
			if _, err := r.client.DeleteTeamMembers(teamId, &api.DeleteTeamMembersParams{
				UserID: currentMember,
			}, api.WithContext(ctx)); err != nil {
				resp.Diagnostics.AddError(
					"Client Error",
					fmt.Sprintf("Unable to delete team's %d member %d, got error: %s", teamId, currentMember, err),
				)
				return
			}
		}
	}
	if data.Members != nil {
		data.Members = members
	}
	// => Captain
	cap := utils.Ptr(utils.Atoi(data.Captain.ValueString()))
	if _, err := r.client.PatchTeam(teamId, &api.PatchTeamsParams{
		CaptainID: cap,
		Fields:    []api.Field{},
	}); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to set user %d as team %d captain, got error: %s", cap, teamId, err),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTeam(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete team %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}
