package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/utils"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*challengeDynamicResource)(nil)
	_ resource.ResourceWithConfigure   = (*challengeDynamicResource)(nil)
	_ resource.ResourceWithImportState = (*challengeDynamicResource)(nil)
)

func NewChallengeDynamicResource() resource.Resource {
	return &challengeDynamicResource{}
}

type challengeDynamicResource struct {
	client *Client
}

// ChallengeDynamicResourceModel is exported for ease of extending
// CTFd through a plugin. Under normal circumpstances, you should
// not use it.
type ChallengeDynamicResourceModel struct {
	ChallengeStandardResourceModel

	Function types.String `tfsdk:"function"`
	Decay    types.Int64  `tfsdk:"decay"`
	Minimum  types.Int64  `tfsdk:"minimum"`
}

func (r *challengeDynamicResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_challenge_dynamic"
}

func (r *challengeDynamicResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CTFd is built around the Challenge resource, which contains all the attributes to define a part of the Capture The Flag event.\n\nThis implementation has support of a more dynamic behavior for its scoring through time/solves thus is different from a standard challenge.",
		Attributes:          ChallengeDynamicResourceAttributes,
	}
}

func (r *challengeDynamicResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *challengeDynamicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, span := StartTFSpan(ctx, r)
	defer span.End()

	var data ChallengeDynamicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create Challenge
	reqs := (*api.Requirements)(nil)
	if data.Requirements != nil {
		preqs := make([]int, 0, len(data.Requirements.Prerequisites))
		for _, preq := range data.Requirements.Prerequisites {
			id, _ := strconv.Atoi(preq.ValueString())
			preqs = append(preqs, id)
		}
		reqs = &api.Requirements{
			Anonymize:     GetAnon(data.Requirements.Behavior),
			Prerequisites: preqs,
		}
	}
	res, err := r.client.PostChallenges(ctx, &api.PostChallengesParams{
		Name:           data.Name.ValueString(),
		Category:       data.Category.ValueString(),
		Description:    data.Description.ValueString(),
		Attribution:    data.Attribution.ValueStringPointer(),
		ConnectionInfo: data.ConnectionInfo.ValueStringPointer(),
		MaxAttempts:    utils.ToInt(data.MaxAttempts),
		Function:       data.Function.ValueStringPointer(),
		Initial:        utils.ToInt(data.Value),
		Decay:          utils.ToInt(data.Decay),
		Minimum:        utils.ToInt(data.Minimum),
		Logic:          data.Logic.ValueString(),
		State:          data.State.ValueString(),
		Type:           "dynamic",
		NextID:         utils.ToInt(data.Next),
		Requirements:   reqs,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create challenge, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a challenge")

	// Save computed attributes in state
	data.ID = types.StringValue(strconv.Itoa(res.ID))

	// Create tags
	challTags := make([]types.String, 0, len(data.Tags))
	for _, tag := range data.Tags {
		_, err := r.client.PostTags(ctx, &api.PostTagsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Value:     tag.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create tags, got error: %s", err),
			)
			return
		}
		challTags = append(challTags, tag)
	}
	if data.Tags != nil {
		data.Tags = challTags
	}

	// Create topics
	challTopics := make([]types.String, 0, len(data.Topics))
	for _, topic := range data.Topics {
		_, err := r.client.PostTopics(ctx, &api.PostTopicsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Type:      "challenge",
			Value:     topic.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create topic, got error: %s", err),
			)
			return
		}
		challTopics = append(challTopics, topic)
	}
	if data.Topics != nil {
		data.Topics = challTopics
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeDynamicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, span := StartTFSpan(ctx, r)
	defer span.End()

	var data ChallengeDynamicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Read(ctx, r.client, resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeDynamicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, span := StartTFSpan(ctx, r)
	defer span.End()

	var data ChallengeDynamicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var dataState ChallengeDynamicResourceModel
	req.State.Get(ctx, &dataState)

	// Patch direct attributes
	reqs := (*api.Requirements)(nil)
	if data.Requirements != nil {
		preqs := make([]int, 0, len(data.Requirements.Prerequisites))
		for _, preq := range data.Requirements.Prerequisites {
			id, _ := strconv.Atoi(preq.ValueString())
			preqs = append(preqs, id)
		}
		reqs = &api.Requirements{
			Anonymize:     GetAnon(data.Requirements.Behavior),
			Prerequisites: preqs,
		}
	}
	_, err := r.client.PatchChallenge(ctx, data.ID.ValueString(), &api.PatchChallengeParams{
		Name:           data.Name.ValueString(),
		Category:       data.Category.ValueString(),
		Description:    data.Description.ValueString(),
		Attribution:    data.Attribution.ValueStringPointer(),
		ConnectionInfo: data.ConnectionInfo.ValueStringPointer(),
		MaxAttempts:    utils.ToInt(data.MaxAttempts),
		Function:       data.Function.ValueStringPointer(),
		Initial:        utils.ToInt(data.Value),
		Decay:          utils.ToInt(data.Decay),
		Minimum:        utils.ToInt(data.Minimum),
		Logic:          data.Logic.ValueStringPointer(),
		State:          data.State.ValueString(),
		NextID:         utils.ToInt(data.Next),
		Requirements:   reqs,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update challenge, got error: %s", err),
		)
		return
	}

	// Update its tags (drop them all, create new ones)
	challTags, err := r.client.GetChallengeTags(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all tags of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, tag := range challTags {
		if err := r.client.DeleteTag(ctx, strconv.Itoa(tag.ID)); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to delete tag %d of challenge %s, got error: %s", tag.ID, data.ID.ValueString(), err),
			)
			return
		}
	}
	tags := make([]types.String, 0, len(data.Tags))
	for _, tag := range data.Tags {
		_, err := r.client.PostTags(ctx, &api.PostTagsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Value:     tag.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create tag of challenge %s, got error: %s", data.ID.ValueString(), err),
			)
			return
		}
		tags = append(tags, tag)
	}
	if data.Tags != nil {
		data.Tags = tags
	}

	// Update its topics (drop them all, create new ones)
	challTopics, err := r.client.GetChallengeTopics(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all topics of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, topic := range challTopics {
		if err := r.client.DeleteTopic(ctx, &api.DeleteTopicArgs{
			ID:   strconv.Itoa(topic.ID),
			Type: "challenge",
		}); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to delete topic %d of challenge %s, got error: %s", topic.ID, data.ID.ValueString(), err),
			)
			return
		}
	}
	topics := make([]types.String, 0, len(data.Topics))
	for _, topic := range data.Topics {
		_, err := r.client.PostTopics(ctx, &api.PostTopicsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Type:      "challenge",
			Value:     topic.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create topic of challenge %s, got error: %s", data.ID.ValueString(), err),
			)
			return
		}
		topics = append(topics, topic)
	}
	if data.Topics != nil {
		data.Topics = topics
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeDynamicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, span := StartTFSpan(ctx, r)
	defer span.End()

	var data ChallengeDynamicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteChallenge(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete challenge, got error: %s", err))
		return
	}

	// ... don't need to delete nested objects, this is handled by CTFd
}

func (r *challengeDynamicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}

//
// Starting from this are helper or types-specific code related to the ctfd_challenge_dynamic resource
//

func (chall *ChallengeDynamicResourceModel) Read(ctx context.Context, client *Client, diags diag.Diagnostics) {
	res, err := client.GetChallenge(ctx, chall.ID.ValueString())
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read challenge %s, got error: %s", chall.ID.ValueString(), err))
		return
	}
	chall.Name = types.StringValue(res.Name)
	chall.Category = types.StringValue(res.Category)
	chall.Description = types.StringValue(res.Description)
	chall.Attribution = types.StringPointerValue(res.Attribution)
	chall.ConnectionInfo = utils.ToTFString(res.ConnectionInfo)
	chall.MaxAttempts = utils.ToTFInt64(res.MaxAttempts)
	chall.Function = utils.ToTFString(res.Function)
	chall.Value = utils.ToTFInt64(res.Initial)
	chall.Decay = utils.ToTFInt64(res.Decay)
	chall.Minimum = utils.ToTFInt64(res.Minimum)
	chall.Logic = types.StringValue(res.Logic)
	chall.State = types.StringValue(res.State)
	chall.Next = utils.ToTFInt64(res.NextID)

	id := utils.Atoi(chall.ID.ValueString())

	// Get subresources
	// => Requirements
	resReqs, err := client.GetChallengeRequirements(ctx, strconv.Itoa(id))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d requirements, got error: %s", id, err),
		)
		return
	}
	reqs := (*RequirementsSubresourceModel)(nil)
	if resReqs != nil {
		challPreqs := make([]types.String, 0, len(resReqs.Prerequisites))
		for _, req := range resReqs.Prerequisites {
			challPreqs = append(challPreqs, types.StringValue(strconv.Itoa(req)))
		}
		reqs = &RequirementsSubresourceModel{
			Behavior:      FromAnon(resReqs.Anonymize),
			Prerequisites: challPreqs,
		}
	}
	chall.Requirements = reqs

	// => Tags
	resTags, err := client.GetChallengeTags(ctx, strconv.Itoa(id))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d tags, got error: %s", id, err),
		)
		return
	}
	chall.Tags = make([]basetypes.StringValue, 0, len(resTags))
	for _, tag := range resTags {
		chall.Tags = append(chall.Tags, types.StringValue(tag.Value))
	}

	// => Topics
	resTopics, err := client.GetChallengeTopics(ctx, strconv.Itoa(id))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d topics, got error: %s", id, err),
		)
		return
	}
	chall.Topics = make([]basetypes.StringValue, 0, len(resTopics))
	for _, topic := range resTopics {
		chall.Topics = append(chall.Topics, types.StringValue(topic.Value))
	}
}

var (
	// ChallengeDynamicResourceAttributes is exported for ease of extending
	// CTFd through a plugin. Under normal circumpstances, you should
	// not use it.
	ChallengeDynamicResourceAttributes = utils.BlindMerge(ChallengeStandardResourceAttributes, map[string]schema.Attribute{
		"function": schema.StringAttribute{
			MarkdownDescription: "Decay function to define how the challenge value evolve through solves, either linear or logarithmic.",
			Optional:            true,
			Computed:            true,
			Default:             defaults.String(stringdefault.StaticString("linear")),
			Validators: []validator.String{
				validators.NewStringEnumValidator([]basetypes.StringValue{
					FunctionLinear,
					FunctionLogarithmic,
				}),
			},
		},
		// value is overwritten with a specific description
		"value": schema.Int64Attribute{
			MarkdownDescription: "The value (points) of the challenge once solved. It is mapped to `initial` under the hood, but displayed as `value` for consistency with the standard challenge.",
			Required:            true,
		},
		"decay": schema.Int64Attribute{
			MarkdownDescription: "The decay defines from each number of solves does the decay function triggers until reaching minimum. This function is defined by CTFd and could be configured through `.function`.",
			Required:            true,
		},
		"minimum": schema.Int64Attribute{
			MarkdownDescription: "The minimum points for a dynamic-score challenge to reach with the decay function. Once there, no solve could have more value.",
			Required:            true,
		},
	})
)
