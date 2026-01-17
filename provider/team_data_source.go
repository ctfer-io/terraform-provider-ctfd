package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	_ datasource.DataSource              = (*teamDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*teamDataSource)(nil)
)

func NewTeamDataSource() datasource.DataSource {
	return &teamDataSource{}
}

type teamDataSource struct {
	client *api.Client
}

type teamsDataSourceModel struct {
	ID    types.String        `tfsdk:"id"`
	Teams []teamResourceModel `tfsdk:"teams"`
}

func (team *teamDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (team *teamDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"teams": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Identifier of the user.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the team.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "Email of the team.",
							Computed:            true,
						},
						"password": schema.StringAttribute{
							MarkdownDescription: "Password of the team. Notice that during a CTF you may not want to update those to avoid defaulting team accesses.",
							Computed:            true,
						},
						"website": schema.StringAttribute{
							MarkdownDescription: "Website, blog, or anything similar (displayed to other participants).",
							Computed:            true,
						},
						"affiliation": schema.StringAttribute{
							MarkdownDescription: "Affiliation to a company or agency.",
							Computed:            true,
						},
						"country": schema.StringAttribute{
							MarkdownDescription: "Country the team represent or is hail from.",
							Computed:            true,
						},
						"hidden": schema.BoolAttribute{
							MarkdownDescription: "Is true if the team is hidden to the participants.",
							Computed:            true,
						},
						"banned": schema.BoolAttribute{
							MarkdownDescription: "Is true if the team is banned from the CTF.",
							Computed:            true,
						},
						"members": schema.SetAttribute{
							MarkdownDescription: "List of members (User), defined by their IDs.",
							ElementType:         types.StringType,
							Computed:            true,
						},
						"captain": schema.StringAttribute{
							MarkdownDescription: "Member who is captain of the team. Must be part of the members too. Note it could cause a fatal error in case of resource import with an inconsistent CTFd configuration i.e. if a team has no captain yet (should not be possible).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (team *teamDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *github.com/ctfer-io/go-ctfd/api.Client, got: %T. Please open an issue at https://github.com/ctfer-io/terraform-provider-ctfd", req.ProviderData),
		)
		return
	}

	team.client = client
}

func (team *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state teamsDataSourceModel

	teams, err := team.client.GetTeams(&api.GetTeamsParams{}, api.WithContext(ctx), api.WithTransport(otelhttp.NewTransport(nil)))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read CTFd Teams",
			err.Error(),
		)
		return
	}

	state.Teams = make([]teamResourceModel, 0, len(teams))
	for _, t := range teams {
		// Flatten response
		members := make([]basetypes.StringValue, 0, len(t.Members))
		for _, tm := range t.Members {
			members = append(members, types.StringValue(strconv.Itoa(tm)))
		}
		state.Teams = append(state.Teams, teamResourceModel{
			ID:          types.StringValue(strconv.Itoa(t.ID)),
			Name:        types.StringValue(t.Name),
			Email:       types.StringValue(*t.Email),
			Password:    types.StringValue("placeholder"),
			Website:     types.StringValue(*t.Website),
			Affiliation: types.StringValue(*t.Affiliation),
			Country:     types.StringValue(*t.Country),
			Hidden:      types.BoolValue(t.Hidden),
			Banned:      types.BoolValue(t.Banned),
			Members:     members,
			Captain:     types.StringValue(strconv.Itoa(*t.CaptainID)),
		})
	}

	state.ID = types.StringValue("placeholder")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
