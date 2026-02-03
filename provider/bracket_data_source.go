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
	_ datasource.DataSource              = (*bracketDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*bracketDataSource)(nil)
)

func NewBracketSource() datasource.DataSource {
	return &bracketDataSource{}
}

type bracketDataSource struct {
	client *Client
}

type bracketsDataSourceModel struct {
	ID       types.String           `tfsdk:"id"`
	Brackets []bracketResourceModel `tfsdk:"brackets"`
}

func (data *bracketDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_brackets"
}

func (data *bracketDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
							MarkdownDescription: "Identifier of the bracket, used internally to handle the CTFd corresponding object.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name displayed to end-users (e.g. \"Students\", \"Interns\", \"Engineers\").",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description that explains the goal of this bracket.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of the bracket, either \"users\" or \"teams\".",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (data *bracketDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	data.client = client
}

func (data *bracketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx, span := StartTFSpan(ctx, data)
	defer span.End()

	var state bracketsDataSourceModel

	brackets, err := data.client.GetBrackets(ctx, &api.GetBracketsParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read CTFd Brackets",
			err.Error(),
		)
		return
	}

	state.Brackets = make([]bracketResourceModel, 0, len(brackets))
	for _, b := range brackets {
		// Flatten response
		state.Brackets = append(state.Brackets, bracketResourceModel{
			ID:          types.StringValue(strconv.Itoa(b.ID)),
			Name:        types.StringValue(b.Name),
			Description: types.StringValue(b.Description),
			Type:        types.StringValue(b.Type),
		})
	}

	state.ID = types.StringValue("placeholder")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
