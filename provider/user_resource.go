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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                = (*userResource)(nil)
	_ resource.ResourceWithConfigure   = (*userResource)(nil)
	_ resource.ResourceWithImportState = (*userResource)(nil)
)

type userResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Email       types.String `tfsdk:"email"`
	Password    types.String `tfsdk:"password"`
	Website     types.String `tfsdk:"website"`
	Affiliation types.String `tfsdk:"affiliation"`
	Country     types.String `tfsdk:"country"`
	Language    types.String `tfsdk:"language"`
	Type        types.String `tfsdk:"type"`
	Verified    types.Bool   `tfsdk:"verified"`
	Hidden      types.Bool   `tfsdk:"hidden"`
	Banned      types.Bool   `tfsdk:"banned"`
	BracketID   types.String `tfsdk:"bracket_id"`
}

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *Client
}

func (r *userResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CTFd defines a User as someone who will either play or administrate the Capture The Flag event.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Identifier of the user.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name or pseudo of the user.",
				Required:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email of the user, may be used to verify the account.",
				Required:            true,
				Sensitive:           true, // Sensitive as PII => GDPR
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password of the user. Notice than during a CTF you may not want to update those to avoid defaulting user accesses.",
				Required:            true,
				Sensitive:           true,
			},
			"website": schema.StringAttribute{
				MarkdownDescription: "Website, blog, or anything similar (displayed to other participants).",
				Optional:            true,
			},
			"affiliation": schema.StringAttribute{
				MarkdownDescription: "Affiliation to a team, company or agency.",
				Optional:            true,
			},
			"country": schema.StringAttribute{
				MarkdownDescription: "Country the user represent or is native from.",
				Optional:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Language the user is fluent in.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Generic type for RBAC purposes.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.String(stringdefault.StaticString("user")),
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						types.StringValue("user"),
						types.StringValue("admin"),
					}),
				},
			},
			"verified": schema.BoolAttribute{
				MarkdownDescription: "Is true if the user has verified its account by email, or if set by an admin.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.Bool(booldefault.StaticBool(false)),
			},
			"hidden": schema.BoolAttribute{
				MarkdownDescription: "Is true if the user is hidden to the participants.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.Bool(booldefault.StaticBool(false)),
			},
			"banned": schema.BoolAttribute{
				MarkdownDescription: "Is true if the user is banned from the CTF.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.Bool(booldefault.StaticBool(false)),
			},
			"bracket_id": schema.StringAttribute{
				MarkdownDescription: "The bracket id the user plays in.",
				Optional:            true,
			},
		},
	}
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.PostUsers(ctx, &api.PostUsersParams{
		Name:        data.Name.ValueString(),
		Email:       data.Email.ValueString(),
		Password:    data.Password.ValueString(),
		Website:     data.Website.ValueStringPointer(),
		Language:    data.Language.ValueStringPointer(),
		Affiliation: data.Affiliation.ValueStringPointer(),
		Country:     data.Country.ValueStringPointer(),
		Type:        data.Type.ValueString(),
		Verified:    data.Verified.ValueBool(),
		Hidden:      data.Hidden.ValueBool(),
		Banned:      data.Banned.ValueBool(),
		Fields:      []api.Field{},
		BracketID:   data.BracketID.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create user, got error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(strconv.Itoa(res.ID))

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.GetUser(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read user %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}

	data.Name = types.StringValue(res.Name)
	data.Email = types.StringPointerValue(res.Email)
	data.Website = types.StringPointerValue(res.Website)
	data.Affiliation = types.StringPointerValue(res.Affiliation)
	data.Country = types.StringPointerValue(res.Country)
	data.Language = types.StringPointerValue(res.Language)
	data.Type = types.StringPointerValue(res.Type)
	data.Verified = types.BoolPointerValue(res.Verified)
	data.Hidden = types.BoolPointerValue(res.Hidden)
	data.Banned = types.BoolPointerValue(res.Banned)
	if res.BracketID != nil {
		data.BracketID = types.StringValue(strconv.Itoa(*res.BracketID))
	}
	// password is not returned, which is good :)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.PatchUser(ctx, data.ID.ValueString(), &api.PatchUsersParams{
		Name:        data.Name.ValueString(),
		Email:       data.Email.ValueString(),
		Password:    data.Password.ValueStringPointer(),
		Website:     data.Website.ValueStringPointer(),
		Affiliation: data.Affiliation.ValueStringPointer(),
		Language:    data.Language.ValueStringPointer(),
		Country:     data.Country.ValueStringPointer(),
		Type:        data.Type.ValueStringPointer(),
		Verified:    data.Verified.ValueBoolPointer(),
		Hidden:      data.Hidden.ValueBoolPointer(),
		Banned:      data.Banned.ValueBoolPointer(),
		Fields:      []api.Field{},
		BracketID:   data.BracketID.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update user, got error: %s", err),
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteUser(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete user %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}
