package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/challenge"
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*challengeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*challengeDataSource)(nil)
)

func NewChallengeDataSource() datasource.DataSource {
	return &challengeDataSource{}
}

type challengeDataSource struct {
	client *api.Client
}

type challengesDataSourceModel struct {
	ID         types.String             `tfsdk:"id"`
	Challenges []challengeResourceModel `tfsdk:"challenges"`
}

func (ch *challengeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_challenges"
}

func (ch *challengeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"challenges": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"category": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"connection_info": schema.StringAttribute{
							Computed: true,
						},
						"value": schema.Int64Attribute{
							Computed: true,
						},
						"initial": schema.Int64Attribute{
							Computed: true,
						},
						"decay": schema.Int64Attribute{
							Computed: true,
						},
						"minimum": schema.Int64Attribute{
							Computed: true,
						},
						"max_attempts": schema.StringAttribute{
							Computed: true,
						},
						"function": schema.StringAttribute{
							Computed: true,
						},
						// TODO add support of next + requirements
						"state": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"flags": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: challenge.FlagSubdatasourceAttributes(),
							},
							Computed: true,
						},
						"tags": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"topics": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"hints": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: challenge.HintSubdatasourceAttributes(),
							},
							Computed: true,
						},
						"files": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: challenge.FileSubdatasourceAttributes(),
							},
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (ch *challengeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	ch.client = client
}

func (ch *challengeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state challengesDataSourceModel

	challs, err := ch.client.GetChallenges(&api.GetChallengesParams{}, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read CTFd Challenges",
			err.Error(),
		)
		return
	}

	for _, chall := range challs {
		// Fetch the challenge with all its information as the CTFd API is broken as fuck
		chall, err := ch.client.GetChallenge(chall.ID, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Read CTFd Challenge %d", chall.ID),
				err.Error(),
			)
			return
		}

		// => Files
		files, err := ch.client.GetChallengeFiles(chall.ID)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Read CTFd files of Challenge %d", chall.ID),
				err.Error(),
			)
			return
		}
		challFiles := make([]challenge.FileSubresourceModel, 0, len(files))
		for _, file := range files {
			f := challenge.FileSubresourceModel{
				ID:       types.StringValue(strconv.Itoa(file.ID)),
				Location: types.StringValue(file.Location),
			}
			f.Read(ctx, resp.Diagnostics, ch.client)

			challFiles = append(challFiles, f)
		}

		// => Flags
		flags, err := ch.client.GetChallengeFlags(chall.ID, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Read CTFd flags of Challenge %d", chall.ID),
				err.Error(),
			)
			return
		}
		challFlags := make([]challenge.FlagSubresourceModel, 0, len(flags))
		for _, flag := range flags {
			challFlags = append(challFlags, challenge.FlagSubresourceModel{
				ID:      types.StringValue(strconv.Itoa(flag.ID)),
				Content: types.StringValue(flag.Content),
				// XXX this should be properly typed
				Data: types.StringValue(flag.Data.(string)),
				Type: types.StringValue(flag.Type),
			})
		}

		// => Tags
		tags, err := ch.client.GetChallengeTags(chall.ID, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Read CTFd tags of Challenge %d", chall.ID),
				err.Error(),
			)
			return
		}
		challTags := make([]types.String, 0, len(tags))
		for _, tag := range tags {
			challTags = append(challTags, types.StringValue(tag.Value))
		}

		// => Topics
		topics, err := ch.client.GetChallengeTopics(chall.ID, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Read CTFd topics of Challenge %d", chall.ID),
				err.Error(),
			)
			return
		}
		challTopics := make([]types.String, 0, len(topics))
		for _, topic := range topics {
			challTopics = append(challTopics, types.StringValue(topic.Value))
		}

		// => Hints
		hints, err := ch.client.GetChallengeHints(chall.ID, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Reac CTFd hints of Challenge %d", chall.ID),
				err.Error(),
			)
			return
		}
		challHints := make([]challenge.HintSubresourceModel, 0, len(hints))
		for _, hint := range hints {
			hintReq := make([]attr.Value, 0, len(hint.Requirements.Prerequisites))
			for _, preq := range hint.Requirements.Prerequisites {
				hintReq = append(hintReq, types.StringValue(strconv.Itoa(preq)))
			}
			challHints = append(challHints, challenge.HintSubresourceModel{
				ID:           types.StringValue(strconv.Itoa(hint.ID)),
				Content:      types.StringValue(*hint.Content),
				Cost:         types.Int64Value(int64(hint.Cost)),
				Requirements: types.ListValueMust(types.StringType, hintReq),
			})
		}

		challState := challengeResourceModel{
			ID:             types.StringValue(strconv.Itoa(chall.ID)),
			Name:           types.StringValue(chall.Name),
			Category:       types.StringValue(chall.Category),
			Description:    types.StringValue(chall.Description),
			ConnectionInfo: utils.ToTFString(chall.ConnectionInfo),
			MaxAttempts:    utils.ToTFInt64(chall.MaxAttempts),
			Function:       types.StringValue(chall.Function),
			Value:          types.Int64Value(int64(chall.Value)),
			Initial:        utils.ToTFInt64(chall.Initial),
			Decay:          utils.ToTFInt64(chall.Decay),
			Minimum:        utils.ToTFInt64(chall.Minimum),
			State:          types.StringValue(chall.State),
			Type:           types.StringValue(chall.Type),
			Files:          challFiles,
			Flags:          challFlags,
			Tags:           challTags,
			Topics:         challTopics,
			Hints:          challHints,
		}

		state.ID = types.StringValue("placeholder")
		state.Challenges = append(state.Challenges, challState)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
