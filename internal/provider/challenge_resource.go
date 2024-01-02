package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/challenge"
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/utils"
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

type challengeResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Category       types.String `tfsdk:"category"`
	Description    types.String `tfsdk:"description"`
	ConnectionInfo types.String `tfsdk:"connection_info"`
	MaxAttempts    types.Int64  `tfsdk:"max_attempts"`
	Function       types.String `tfsdk:"function"`
	Value          types.Int64  `tfsdk:"value"`
	Initial        types.Int64  `tfsdk:"initial"`
	Decay          types.Int64  `tfsdk:"decay"`
	Minimum        types.Int64  `tfsdk:"minimum"`
	State          types.String `tfsdk:"state"`
	Type           types.String `tfsdk:"type"`
	// TODO add support of Next challenges
	Requirements *challenge.RequirementsSubresourceModel `tfsdk:"requirements"`
	Flags        []challenge.FlagSubresourceModel        `tfsdk:"flags"`
	Tags         []types.String                          `tfsdk:"tags"`
	Topics       []types.String                          `tfsdk:"topics"`
	Hints        []challenge.HintSubresourceModel        `tfsdk:"hints"`
	Files        []challenge.FileSubresourceModel        `tfsdk:"files"`
}

func (r *challengeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_challenge"
}

func (r *challengeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the challenge",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the challenge, displayed as it",
				Required:            true,
			},
			"category": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Required: true,
			},
			"connection_info": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"max_attempts": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"function": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             defaults.String(stringdefault.StaticString("linear")),
				Description:         "Decay function to define how the challenge's value will change through time.",
				MarkdownDescription: "Decay function to define how the challenge's value will change through time.",
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						challenge.FunctionLinear,
						challenge.FunctionLogarithmic,
					}),
				},
			},
			// TODO value can't be set side to <initial,decay,minimum>, depends on .type value (respectively standard and dynamic)
			"value": schema.Int64Attribute{
				Optional: true,
			},
			"initial": schema.Int64Attribute{
				Optional: true,
			},
			"decay": schema.Int64Attribute{
				Optional: true,
			},
			"minimum": schema.Int64Attribute{
				Optional: true,
			},
			"state": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("hidden"),
			},
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("dynamic"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"requirements": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"behavior": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Behavior if not unlocked.",
						MarkdownDescription: "Behavior if not unlocked.",
						Default:             stringdefault.StaticString("hidden"),
						Validators: []validator.String{
							validators.NewStringEnumValidator([]basetypes.StringValue{
								challenge.BehaviorHidden,
								challenge.BehaviorAnonymized,
							}),
						},
					},
					"prerequisites": schema.ListAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"flags": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: challenge.FlagSubresourceAttributes(),
				},
				Optional: true,
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"topics": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"hints": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: challenge.HintSubresourceAttributes(),
				},
				Optional: true,
			},
			"files": schema.ListNestedAttribute{
				// TODO find why modifying other fields requires updating those
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
		Initial:        utils.ToInt(data.Initial),
		Decay:          utils.ToInt(data.Decay),
		Minimum:        utils.ToInt(data.Minimum),
		State:          data.State.ValueString(),
		Type:           data.Type.ValueString(),
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

	// Retrieve challenge
	res, err := r.client.GetChallenge(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read challenge %s, got error: %s", data.ID.ValueString(), err))
		return
	}
	data.Name = types.StringValue(res.Name)
	data.Category = types.StringValue(res.Category)
	data.Description = types.StringValue(res.Description)
	data.ConnectionInfo = utils.ToTFString(res.ConnectionInfo)
	data.MaxAttempts = utils.ToTFInt64(res.MaxAttempts)
	data.Function = types.StringValue("linear") // XXX CTFd does not return the `function` attribute
	data.Value = types.Int64Value(int64(res.Value))
	data.Initial = utils.ToTFInt64(res.Initial)
	data.Decay = utils.ToTFInt64(res.Decay)
	data.Minimum = utils.ToTFInt64(res.Minimum)
	data.State = types.StringValue(res.State)
	data.Type = types.StringValue(res.Type)

	id := utils.Atoi(data.ID.ValueString())

	// Get subresources
	// => Requirements
	resReqs, err := r.client.GetChallengeRequirements(id, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d requirements, got error: %s", id, err),
		)
	}
	reqs := (*challenge.RequirementsSubresourceModel)(nil)
	if resReqs != nil {
		challPreqs := make([]types.String, 0, len(resReqs.Prerequisites))
		for _, req := range resReqs.Prerequisites {
			challPreqs = append(challPreqs, types.StringValue(strconv.Itoa(req)))
		}
		reqs = &challenge.RequirementsSubresourceModel{
			Behavior:      challenge.FromAnon(resReqs.Anonymize),
			Prerequisites: challPreqs,
		}
	}
	data.Requirements = reqs

	// => Files
	resFiles, err := r.client.GetChallengeFiles(id, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d files, got error: %s", id, err),
		)
		return
	}
	data.Files = make([]challenge.FileSubresourceModel, 0, len(resFiles))
	for _, file := range resFiles {
		c, err := r.client.GetFileContent(file, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read file content at %s, got error: %s", file.Location, err),
			)
			continue
		}
		nf := challenge.FileSubresourceModel{
			ID:       types.StringValue(strconv.Itoa(file.ID)),
			Name:     types.StringValue(utils.Filename(file.Location)),
			Location: types.StringValue(file.Location),
			Content:  types.StringValue(string(c)),
		}
		nf.PropagateContent(ctx, resp.Diagnostics)
		data.Files = append(data.Files, nf)
	}

	// => Flags
	resFlags, err := r.client.GetChallengeFlags(id, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d flags, got error: %s", id, err),
		)
		return
	}
	data.Flags = make([]challenge.FlagSubresourceModel, 0, len(resFlags))
	for _, flag := range resFlags {
		data.Flags = append(data.Flags, challenge.FlagSubresourceModel{
			ID:      types.StringValue(strconv.Itoa(flag.ID)),
			Content: types.StringValue(flag.Content),
			// XXX this should be typed properly
			Data: types.StringValue(flag.Data.(string)),
			Type: types.StringValue(flag.Type),
		})
	}

	// => Hints
	resHints, err := r.client.GetChallengeHints(id, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d hints, got error: %s", id, err),
		)
		return
	}
	data.Hints = make([]challenge.HintSubresourceModel, 0, len(resHints))
	for _, hint := range resHints {
		reqs := []attr.Value{}
		if hint.Requirements != nil {
			reqs = make([]attr.Value, 0, len(hint.Requirements.Prerequisites))
			for _, req := range hint.Requirements.Prerequisites {
				reqs = append(reqs, types.StringValue(strconv.Itoa(req)))
			}
		}
		data.Hints = append(data.Hints, challenge.HintSubresourceModel{
			ID:           types.StringValue(strconv.Itoa(hint.ID)),
			Content:      types.StringValue(*hint.Content),
			Cost:         types.Int64Value(int64(hint.Cost)),
			Requirements: types.ListValueMust(types.StringType, reqs),
		})
	}

	// => Tags
	resTags, err := r.client.GetChallengeTags(id, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d tags, got error: %s", id, err),
		)
		return
	}
	data.Tags = make([]basetypes.StringValue, 0, len(resTags))
	for _, tag := range resTags {
		data.Tags = append(data.Tags, types.StringValue(tag.Value))
	}

	// => Topics
	resTopics, err := r.client.GetChallengeTopics(id, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d topics, got error: %s", id, err),
		)
		return
	}
	data.Topics = make([]basetypes.StringValue, 0, len(resTopics))
	for _, topic := range resTopics {
		data.Topics = append(data.Topics, types.StringValue(topic.Value))
	}

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
		// Value:          int(data.Value.ToInt64Value()), // TODO add support of .value in PATCH /challenges
		Initial:      utils.ToInt(data.Initial),
		Decay:        utils.ToInt(data.Decay),
		Minimum:      utils.ToInt(data.Minimum),
		State:        data.State.ValueString(),
		Requirements: reqs,
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
