package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*userDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*userDataSource)(nil)
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *Client
}

type usersDataSourceModel struct {
	ID    types.String        `tfsdk:"id"`
	Users []userResourceModel `tfsdk:"users"`
}

func (r *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (r *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Identifier of the user.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name or pseudo of the user.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "Email of the user, may be used to verify the account.",
							Computed:            true,
						},
						"password": schema.StringAttribute{
							MarkdownDescription: "Password of the user. Notice that during a CTF you may not want to update those to avoid defaulting user accesses.",
							Computed:            true,
						},
						"website": schema.StringAttribute{
							MarkdownDescription: "Website, blog, or anything similar (displayed to other participants).",
							Computed:            true,
						},
						"affiliation": schema.StringAttribute{
							MarkdownDescription: "Affiliation to a team, company or agency.",
							Computed:            true,
						},
						"country": schema.StringAttribute{
							MarkdownDescription: "Country the user represent or is native from.",
							Computed:            true,
						},
						"language": schema.StringAttribute{
							MarkdownDescription: "Language the user is fluent in.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Generic type for RBAC purposes.",
							Computed:            true,
						},
						"verified": schema.BoolAttribute{
							MarkdownDescription: "Is true if the user has verified its account by email, or if set by an admin.",
							Computed:            true,
						},
						"hidden": schema.BoolAttribute{
							MarkdownDescription: "Is true if the user is hidden to the participants.",
							Computed:            true,
						},
						"banned": schema.BoolAttribute{
							MarkdownDescription: "Is true if the user is banned from the CTF.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (r *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *github.com/ctfer-io/go-ctfd/api.Client, got: %T. Please open an issue at https://github.com/ctfer-io/terraform-provider-ctfd", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceModel

	users, err := r.client.GetUsers(ctx, &api.GetUsersParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read CTFd Users",
			err.Error(),
		)
		return
	}

	state.Users = make([]userResourceModel, 0, len(users))
	for _, u := range users {
		// Flatten response
		state.Users = append(state.Users, userResourceModel{
			ID:          types.StringValue(strconv.Itoa(u.ID)),
			Name:        types.StringValue(u.Name),
			Email:       types.StringPointerValue(u.Email),
			Password:    types.StringValue("placeholder"),
			Website:     types.StringPointerValue(u.Website),
			Affiliation: types.StringPointerValue(u.Affiliation),
			Country:     types.StringPointerValue(u.Country),
			Language:    types.StringPointerValue(u.Language),
			Type:        types.StringPointerValue(u.Type),
			Verified:    types.BoolPointerValue(u.Verified),
			Hidden:      types.BoolPointerValue(u.Hidden),
			Banned:      types.BoolPointerValue(u.Banned),
		})
	}

	state.ID = types.StringValue("placeholder")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
