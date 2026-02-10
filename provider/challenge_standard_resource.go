package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/utils"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*challengeStandardResource)(nil)
	_ resource.ResourceWithConfigure   = (*challengeStandardResource)(nil)
	_ resource.ResourceWithImportState = (*challengeStandardResource)(nil)
)

func NewChallengeStandardResource() resource.Resource {
	return &challengeStandardResource{}
}

type challengeStandardResource struct {
	fm *Framework
}

// ChallengeStandardResourceModel is exported for ease of extending
// CTFd through a plugin. Under normal circumpstances, you should
// not use it.
type ChallengeStandardResourceModel struct {
	ID             types.String                  `tfsdk:"id"`
	Name           types.String                  `tfsdk:"name"`
	Category       types.String                  `tfsdk:"category"`
	Description    types.String                  `tfsdk:"description"`
	Attribution    types.String                  `tfsdk:"attribution"`
	ConnectionInfo types.String                  `tfsdk:"connection_info"`
	MaxAttempts    types.Int64                   `tfsdk:"max_attempts"`
	Value          types.Int64                   `tfsdk:"value"`
	Logic          types.String                  `tfsdk:"logic"`
	State          types.String                  `tfsdk:"state"`
	Next           types.Int64                   `tfsdk:"next"`
	Requirements   *RequirementsSubresourceModel `tfsdk:"requirements"`
	Tags           []types.String                `tfsdk:"tags"`
	Topics         []types.String                `tfsdk:"topics"`
}

func (r *challengeStandardResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_challenge_standard"
}

func (r *challengeStandardResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CTFd is built around the Challenge resource, which contains all the attributes to define a part of the Capture The Flag event.\n\nIt is the first historic implementation of its kind, with basic functionalities.",
		Attributes:          ChallengeStandardResourceAttributes,
	}
}

func (r *challengeStandardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *challengeStandardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data ChallengeStandardResourceModel
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
	res, err := r.fm.Client.PostChallenges(ctx, &api.PostChallengesParams{
		Name:           data.Name.ValueString(),
		Category:       data.Category.ValueString(),
		Description:    data.Description.ValueString(),
		Attribution:    data.Attribution.ValueStringPointer(),
		ConnectionInfo: data.ConnectionInfo.ValueStringPointer(),
		MaxAttempts:    utils.ToInt(data.MaxAttempts),
		Value:          int(data.Value.ValueInt64()),
		Logic:          data.Logic.ValueString(),
		State:          data.State.ValueString(),
		Type:           "standard",
		NextID:         utils.ToInt(data.Next),
		Requirements:   reqs,
	}, WithTracerProvider(r.fm.Tp))
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
		_, err := r.fm.Client.PostTags(ctx, &api.PostTagsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Value:     tag.ValueString(),
		}, WithTracerProvider(r.fm.Tp))
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
		_, err := r.fm.Client.PostTopics(ctx, &api.PostTopicsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Type:      "challenge",
			Value:     topic.ValueString(),
		}, WithTracerProvider(r.fm.Tp))
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

func (r *challengeStandardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data ChallengeStandardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Read(ctx, r.fm.Client, resp.Diagnostics, WithTracerProvider(r.fm.Tp))

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeStandardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data ChallengeStandardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var dataState ChallengeStandardResourceModel
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
	_, err := r.fm.Client.PatchChallenge(ctx, data.ID.ValueString(), &api.PatchChallengeParams{
		Name:           data.Name.ValueString(),
		Category:       data.Category.ValueString(),
		Description:    data.Description.ValueString(),
		Attribution:    data.Attribution.ValueStringPointer(),
		ConnectionInfo: data.ConnectionInfo.ValueStringPointer(),
		MaxAttempts:    utils.ToInt(data.MaxAttempts),
		Value:          utils.ToInt(data.Value),
		Logic:          data.Logic.ValueStringPointer(),
		State:          data.State.ValueString(),
		NextID:         utils.ToInt(data.Next),
		Requirements:   reqs,
	}, WithTracerProvider(r.fm.Tp))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update challenge, got error: %s", err),
		)
		return
	}

	// Update its tags (drop them all, create new ones)
	challTags, err := r.fm.Client.GetChallengeTags(ctx, data.ID.ValueString(), WithTracerProvider(r.fm.Tp))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all tags of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, tag := range challTags {
		if err := r.fm.Client.DeleteTag(ctx, strconv.Itoa(tag.ID), WithTracerProvider(r.fm.Tp)); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to delete tag %d of challenge %s, got error: %s", tag.ID, data.ID.ValueString(), err),
			)
			return
		}
	}
	tags := make([]types.String, 0, len(data.Tags))
	for _, tag := range data.Tags {
		_, err := r.fm.Client.PostTags(ctx, &api.PostTagsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Value:     tag.ValueString(),
		}, WithTracerProvider(r.fm.Tp))
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
	challTopics, err := r.fm.Client.GetChallengeTopics(ctx, data.ID.ValueString(), WithTracerProvider(r.fm.Tp))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all topics of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, topic := range challTopics {
		if err := r.fm.Client.DeleteTopic(ctx, &api.DeleteTopicArgs{
			ID:   strconv.Itoa(topic.ID),
			Type: "challenge",
		}, WithTracerProvider(r.fm.Tp)); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to delete topic %d of challenge %s, got error: %s", topic.ID, data.ID.ValueString(), err),
			)
			return
		}
	}
	topics := make([]types.String, 0, len(data.Topics))
	for _, topic := range data.Topics {
		_, err := r.fm.Client.PostTopics(ctx, &api.PostTopicsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Type:      "challenge",
			Value:     topic.ValueString(),
		}, WithTracerProvider(r.fm.Tp))
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

func (r *challengeStandardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, span := StartTFSpan(ctx, r.fm.Tp.Tracer(serviceName), r)
	defer span.End()

	var data ChallengeStandardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.fm.Client.DeleteChallenge(ctx, data.ID.ValueString(), WithTracerProvider(r.fm.Tp)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete challenge, got error: %s", err))
		return
	}

	// ... don't need to delete nested objects, this is handled by CTFd
}

func (r *challengeStandardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}

//
// Starting from this are helper or types-specific code related to the ctfd_challenge_standard resource
//

func (chall *ChallengeStandardResourceModel) Read(ctx context.Context, client *Client, diags diag.Diagnostics, opts ...Option) {
	res, err := client.GetChallenge(ctx, chall.ID.ValueString(), opts...)
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
	chall.Value = types.Int64Value(int64(res.Value))
	chall.Logic = types.StringValue(res.Logic)
	chall.State = types.StringValue(res.State)
	chall.Next = utils.ToTFInt64(res.NextID)

	id := utils.Atoi(chall.ID.ValueString())

	// Get subresources
	// => Requirements
	resReqs, err := client.GetChallengeRequirements(ctx, strconv.Itoa(id), opts...)
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
	resTags, err := client.GetChallengeTags(ctx, strconv.Itoa(id), opts...)
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
	resTopics, err := client.GetChallengeTopics(ctx, strconv.Itoa(id), opts...)
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
	// ChallengeStandardResourceAttributes is exported for ease of extending
	// CTFd through a plugin. Under normal circumpstances, you should
	// not use it.
	ChallengeStandardResourceAttributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Identifier of the challenge.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the challenge, displayed as it.",
			Required:            true,
		},
		"category": schema.StringAttribute{
			MarkdownDescription: "Category of the challenge that CTFd groups by on the web UI.",
			Required:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description of the challenge, consider using multiline descriptions for better style.",
			Required:            true,
		},
		"attribution": schema.StringAttribute{
			MarkdownDescription: "Attribution to the creator(s) of the challenge.",
			Optional:            true,
		},
		"connection_info": schema.StringAttribute{
			MarkdownDescription: "Connection Information to connect to the challenge instance, useful for pwn, web and infrastructure pentests.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
		},
		"max_attempts": schema.Int64Attribute{
			MarkdownDescription: "Maximum amount of attempts before being unable to flag the challenge.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(0),
		},
		"value": schema.Int64Attribute{
			MarkdownDescription: "The value (points) of the challenge once solved.",
			Required:            true,
		},
		"logic": schema.StringAttribute{
			MarkdownDescription: "The flag validation logic.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("any"),
			Validators: []validator.String{
				validators.NewStringEnumValidator([]basetypes.StringValue{
					types.StringValue("any"),
					types.StringValue("all"),
					types.StringValue("team"),
				}),
			},
		},
		"state": schema.StringAttribute{
			MarkdownDescription: "State of the challenge, either hidden or visible.",
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
		"next": schema.Int64Attribute{
			MarkdownDescription: "Suggestion for the end-user as next challenge to work on.",
			Optional:            true,
		},
		"requirements": schema.SingleNestedAttribute{
			MarkdownDescription: "List of required challenges that needs to get flagged before this one being accessible. Useful for skill-trees-like strategy CTF.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"behavior": schema.StringAttribute{
					MarkdownDescription: "Behavior if not unlocked, either hidden or anonymized.",
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString("hidden"),
					Validators: []validator.String{
						validators.NewStringEnumValidator([]basetypes.StringValue{
							BehaviorHidden,
							BehaviorAnonymized,
						}),
					},
				},
				"prerequisites": schema.SetAttribute{
					MarkdownDescription: "List of the challenges ID.",
					ElementType:         types.StringType,
					Optional:            true,
				},
			},
		},
		"tags": schema.SetAttribute{
			MarkdownDescription: "List of challenge tags that will be displayed to the end-user. You could use them to give some quick insights of what a challenge involves.",
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
			Default:             setdefault.StaticValue(basetypes.NewSetValueMust(types.StringType, []attr.Value{})),
		},
		"topics": schema.SetAttribute{
			MarkdownDescription: "List of challenge topics that are displayed to the administrators for maintenance and planification.",
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
			Default:             setdefault.StaticValue(basetypes.NewSetValueMust(types.StringType, []attr.Value{})),
		},
	}
)
