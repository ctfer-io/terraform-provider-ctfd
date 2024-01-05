package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/challenge"
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/utils"
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*challengeResource)(nil)
	_ resource.ResourceWithConfigure   = (*challengeResource)(nil)
	_ resource.ResourceWithImportState = (*challengeResource)(nil)
)

func NewChallengeResource() resource.Resource {
	return &challengeResource{}
}

type challengeResource struct {
	client *api.Client
}

func (r *challengeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_challenge"
}

func (r *challengeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			"function": schema.StringAttribute{
				MarkdownDescription: "Decay function to define how the challenge value evolve through solves, either linear or logarithmic.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.String(stringdefault.StaticString("linear")),
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						challenge.FunctionLinear,
						challenge.FunctionLogarithmic,
					}),
				},
			},
			"value": schema.Int64Attribute{
				MarkdownDescription: "The value (points) of the challenge once solved. Internally, the provider will handle what target is legitimate depending on the `.type` value, i.e. either `value` for \"standard\" or `initial` for \"dynamic\".",
				Optional:            true,
			},
			// XXX decay and minimum are only required if .type == "dynamic"
			"decay": schema.Int64Attribute{
				MarkdownDescription: "The decay defines from each number of solves does the decay function triggers until reaching minimum. This function is defined by CTFd and could be configured through `.function`.",
				Optional:            true,
			},
			"minimum": schema.Int64Attribute{
				MarkdownDescription: "The minimum points for a dynamic-score challenge to reach with the decay function. Once there, no solve could have more value.",
				Optional:            true,
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
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the challenge defining its layout/behavior, either standard or dynamic (default).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("dynamic"),
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						types.StringValue("standard"),
						types.StringValue("dynamic"),
					}),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
								challenge.BehaviorHidden,
								challenge.BehaviorAnonymized,
							}),
						},
					},
					"prerequisites": schema.ListAttribute{
						MarkdownDescription: "List of the challenges ID.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"flags": schema.ListNestedAttribute{
				MarkdownDescription: "List of challenge flags that solves it.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: challenge.FlagSubresourceAttributes(),
				},
				Optional: true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "List of challenge tags that will be displayed to the end-user. You could use them to give some quick insights of what a challenge involves.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"topics": schema.ListAttribute{
				MarkdownDescription: "List of challenge topics that are displayed to the administrators for maintenance and planification.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"hints": schema.ListNestedAttribute{
				MarkdownDescription: "List of hints about the challenge displayed to the end-user.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: challenge.HintSubresourceAttributes(),
				},
				Optional: true,
			},
			"files": schema.ListNestedAttribute{
				MarkdownDescription: "List of files given to players to flag the challenge.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: challenge.FileSubresourceAttributes(),
				},
				Optional: true,
			},
		},
	}
}

func (r *challengeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *challengeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data challengeResourceModel
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
			Anonymize:     challenge.GetAnon(data.Requirements.Behavior),
			Prerequisites: preqs,
		}
	}
	res, err := r.client.PostChallenges(&api.PostChallengesParams{
		Name:           data.Name.ValueString(),
		Category:       data.Category.ValueString(),
		Description:    data.Description.ValueString(),
		ConnectionInfo: data.ConnectionInfo.ValueStringPointer(),
		MaxAttempts:    utils.ToInt(data.MaxAttempts),
		Function:       data.Function.ValueString(),
		Value:          int(data.Value.ValueInt64()),
		Initial:        utils.ToInt(data.Value),
		Decay:          utils.ToInt(data.Decay),
		Minimum:        utils.ToInt(data.Minimum),
		State:          data.State.ValueString(),
		Type:           data.Type.ValueString(),
		NextID:         utils.ToInt(data.Next),
		Requirements:   reqs,
	}, api.WithContext(ctx))
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

	// Create files
	challFiles := make([]challenge.FileSubresourceModel, 0, len(data.Files))
	for _, file := range data.Files {
		file.Create(ctx, resp.Diagnostics, r.client, utils.Atoi(data.ID.ValueString()))
		challFiles = append(challFiles, file)
	}
	if data.Files != nil {
		data.Files = challFiles
	}

	// Create flags
	challFlags := make([]challenge.FlagSubresourceModel, 0, len(data.Flags))
	for _, flag := range data.Flags {
		flag.Create(ctx, resp.Diagnostics, r.client, utils.Atoi(data.ID.ValueString()))
		challFlags = append(challFlags, flag)
	}
	if data.Flags != nil {
		data.Flags = challFlags
	}

	// Create tags
	challTags := make([]types.String, 0, len(data.Tags))
	for _, tag := range data.Tags {
		_, err := r.client.PostTags(&api.PostTagsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Value:     tag.ValueString(),
		}, api.WithContext(ctx))
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
		_, err := r.client.PostTopics(&api.PostTopicsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Type:      "challenge",
			Value:     topic.ValueString(),
		}, api.WithContext(ctx))
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

	// Create hints
	challHints := make([]challenge.HintSubresourceModel, 0, len(data.Hints))
	for _, hint := range data.Hints {
		hint.Create(ctx, resp.Diagnostics, r.client, utils.Atoi(data.ID.ValueString()))
		challHints = append(challHints, hint)
	}
	if data.Hints != nil {
		data.Hints = challHints
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data challengeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Read(ctx, resp.Diagnostics, r.client)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data challengeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var dataState challengeResourceModel
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
			Anonymize:     challenge.GetAnon(data.Requirements.Behavior),
			Prerequisites: preqs,
		}
	}
	_, err := r.client.PatchChallenge(utils.Atoi(data.ID.ValueString()), &api.PatchChallengeParams{
		Name:           data.Name.ValueString(),
		Category:       data.Category.ValueString(),
		Description:    data.Description.ValueString(),
		ConnectionInfo: data.ConnectionInfo.ValueStringPointer(),
		MaxAttempts:    utils.ToInt(data.MaxAttempts),
		Function:       data.Function.ValueString(),
		Value:          utils.ToInt(data.Value),
		Initial:        utils.ToInt(data.Value),
		Decay:          utils.ToInt(data.Decay),
		Minimum:        utils.ToInt(data.Minimum),
		State:          data.State.ValueString(),
		NextID:         utils.ToInt(data.Next),
		Requirements:   reqs,
	}, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update challenge, got error: %s", err),
		)
		return
	}

	// Update its files
	currentFiles, err := r.client.GetChallengeFiles(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get challenge's files, got error: %s", err),
		)
	}
	files := []challenge.FileSubresourceModel{}
	for _, file := range data.Files {
		exists := false
		for _, currentFile := range currentFiles {
			if file.ID.ValueString() == strconv.Itoa(currentFile.ID) {
				exists = true

				// Get corresponding file from state
				var corFile challenge.FileSubresourceModel
				for _, fState := range dataState.Files {
					if file.ID.Equal(fState.ID) {
						corFile = fState
						break
					}
				}

				// => Drop and replace iif content changed
				update := !corFile.Content.Equal(file.Content)
				if update {
					file.Delete(ctx, resp.Diagnostics, r.client)
					file.Create(ctx, resp.Diagnostics, r.client, utils.Atoi(data.ID.ValueString()))
				}

				files = append(files, file)
				break
			}
		}
		if !exists {
			file.Create(ctx, resp.Diagnostics, r.client, utils.Atoi(data.ID.ValueString()))
			files = append(files, file)
		}
	}
	for _, currentFile := range currentFiles {
		exists := false
		for _, tfFile := range data.Files {
			if tfFile.ID.ValueString() == strconv.Itoa(currentFile.ID) {
				exists = true
				break
			}
		}
		if !exists {
			f := challenge.FileSubresourceModel{
				ID:       types.StringValue(strconv.Itoa(currentFile.ID)),
				Location: types.StringValue(currentFile.Location),
			}
			f.Delete(ctx, resp.Diagnostics, r.client)
		}
	}
	if data.Files != nil {
		data.Files = files
	}

	// Update its flags
	currentFlags, err := r.client.GetChallengeFlags(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get challenge's flags, got error: %s", err),
		)
		return
	}
	flags := []challenge.FlagSubresourceModel{}
	for _, tfFlag := range data.Flags {
		exists := false
		for _, currentFlag := range currentFlags {
			if tfFlag.ID.ValueString() == strconv.Itoa(currentFlag.ID) {
				exists = true
				tfFlag.Update(ctx, resp.Diagnostics, r.client)
				flags = append(flags, tfFlag)
				break
			}
		}
		if !exists {
			tfFlag.Create(ctx, resp.Diagnostics, r.client, utils.Atoi(data.ID.ValueString()))
			flags = append(flags, tfFlag)
		}
	}
	for _, currentFlag := range currentFlags {
		exists := false
		for _, tfFlag := range data.Flags {
			if tfFlag.ID.ValueString() == strconv.Itoa(currentFlag.ID) {
				exists = true
				break
			}
		}
		if !exists {
			f := &challenge.FlagSubresourceModel{
				ID: types.StringValue(strconv.Itoa(currentFlag.ID)),
			}
			f.Delete(ctx, resp.Diagnostics, r.client)
		}
	}
	if data.Flags != nil {
		data.Flags = flags
	}

	// Update its tags (drop them all, create new ones)
	challTags, err := r.client.GetChallengeTags(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all tags of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, tag := range challTags {
		if err := r.client.DeleteTag(strconv.Itoa(tag.ID), api.WithContext(ctx)); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to delete tag %d of challenge %s, got error: %s", tag.ID, data.ID.ValueString(), err),
			)
			return
		}
	}
	tags := make([]types.String, 0, len(data.Tags))
	for _, tag := range data.Tags {
		_, err := r.client.PostTags(&api.PostTagsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Value:     tag.ValueString(),
		}, api.WithContext(ctx))
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
	challTopics, err := r.client.GetChallengeTopics(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all topics of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, topic := range challTopics {
		if err := r.client.DeleteTopic(&api.DeleteTopicArgs{
			ID:   strconv.Itoa(topic.ID),
			Type: "challenge",
		}, api.WithContext(ctx)); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to delete topic %d of challenge %s, got error: %s", topic.ID, data.ID.ValueString(), err),
			)
			return
		}
	}
	topics := make([]types.String, 0, len(data.Topics))
	for _, topic := range data.Topics {
		_, err := r.client.PostTopics(&api.PostTopicsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Type:      "challenge",
			Value:     topic.ValueString(),
		}, api.WithContext(ctx))
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

	// Update its hints
	currentHints, err := r.client.GetChallengeHints(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get challenge's hints, got error: %s", err),
		)
		return
	}
	hints := []challenge.HintSubresourceModel{}
	for _, tfHint := range data.Hints {
		exists := false
		for _, currentHint := range currentHints {
			if tfHint.ID.ValueString() == strconv.Itoa(currentHint.ID) {
				exists = true
				tfHint.Update(ctx, resp.Diagnostics, r.client)
				hints = append(hints, tfHint)
				break
			}
		}
		if !exists {
			tfHint.Create(ctx, resp.Diagnostics, r.client, utils.Atoi(data.ID.ValueString()))
			hints = append(hints, tfHint)
		}
	}
	for _, currentHint := range currentHints {
		exists := false
		for _, tfHint := range data.Hints {
			if tfHint.ID.ValueString() == strconv.Itoa(currentHint.ID) {
				exists = true
				break
			}
		}
		if !exists {
			h := &challenge.HintSubresourceModel{
				ID: types.StringValue(strconv.Itoa(currentHint.ID)),
			}
			h.Delete(ctx, resp.Diagnostics, r.client)
		}
	}
	if data.Hints != nil {
		data.Hints = hints
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data challengeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteChallenge(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete challenge, got error: %s", err))
		return
	}

	// ... don't need to delete nested objects, this is handled by CTFd
}

func (r *challengeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}
