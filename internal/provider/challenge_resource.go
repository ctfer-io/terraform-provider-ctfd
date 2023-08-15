package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	api "github.com/pandatix/go-ctfd/api"
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

// TODO support maxAttempts attribute
type challengeResourceModel struct {
	ID             types.String                 `tfsdk:"id"`
	Name           types.String                 `tfsdk:"name"`
	Category       types.String                 `tfsdk:"category"`
	Description    types.String                 `tfsdk:"description"`
	ConnectionInfo types.String                 `tfsdk:"connection_info"`
	MaxAttempts    types.Int64                  `tfsdk:"max_attempts"`
	Value          types.Int64                  `tfsdk:"value"`
	Initial        types.Int64                  `tfsdk:"initial"`
	Decay          types.Int64                  `tfsdk:"decay"`
	Minimum        types.Int64                  `tfsdk:"minimum"`
	State          types.String                 `tfsdk:"state"`
	Type           types.String                 `tfsdk:"type"`
	Requirements   requirementsSubresourceModel `tfsdk:"requirements"`
	Flags          []flagSubresourceModel       `tfsdk:"flags"`
	Tags           []types.String               `tfsdk:"tags"`
	Topics         []types.String               `tfsdk:"topics"`
	Hints          []hintSubresourceModel       `tfsdk:"hints"`
	Files          []fileSubresourceModel       `tfsdk:"files"`
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
			},
			"max_attempts": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
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
							NewStringEnumValidator([]basetypes.StringValue{
								behaviorHidden,
								behaviorAnonymized,
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
					Attributes: flagSubresourceAttributes(),
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
					Attributes: hintSubresourceAttributes(),
				},
				Optional: true,
			},
			"files": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: fileSubresourceAttributes(),
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
			fmt.Sprintf("Expected *github.com/pandatix/go-ctfd/api.Client, got: %T. Please open an issue at https://github.com/pandatix/terraform-provider-ctfd", req.ProviderData),
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
	preqs := make([]int, 0, len(data.Requirements.Prerequisites))
	for _, preq := range data.Requirements.Prerequisites {
		id, _ := strconv.Atoi(preq.ValueString())
		preqs = append(preqs, id)
	}
	res, err := r.client.PostChallenges(&api.PostChallengesParams{
		Name:           data.Name.ValueString(),
		Category:       data.Category.ValueString(),
		Description:    data.Description.ValueString(),
		ConnectionInfo: data.Category.ValueStringPointer(),
		MaxAttempts:    toInt(data.MaxAttempts),
		Value:          toInt(data.Value),
		Initial:        toInt(data.Initial),
		Decay:          toInt(data.Decay),
		Minimum:        toInt(data.Minimum),
		State:          data.State.ValueString(),
		Type:           data.Type.ValueString(),
		Requirements: &api.Requirements{
			Anonymize:     getAnon(data.Requirements.Behavior),
			Prerequisites: preqs,
		},
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
	challFiles := make([]fileSubresourceModel, 0, len(data.Files))
	for _, file := range data.Files {
		file.Create(ctx, resp.Diagnostics, r.client, data.ID.ValueString())
		challFiles = append(challFiles, file)
	}
	if data.Files != nil {
		data.Files = challFiles
	}

	// Create flags
	challFlags := make([]flagSubresourceModel, 0, len(data.Flags))
	for _, flag := range data.Flags {
		flag.Create(ctx, resp.Diagnostics, r.client, data.ID.ValueString())
		challFlags = append(challFlags, flag)
	}
	if data.Flags != nil {
		data.Flags = challFlags
	}

	// Create tags
	challTags := make([]types.String, 0, len(data.Tags))
	for _, tag := range data.Tags {
		_, err := r.client.PostTags(&api.PostTagsParams{
			Challenge: data.ID.ValueString(),
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
			Challenge: data.ID.ValueString(),
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
	challHints := make([]hintSubresourceModel, 0, len(data.Hints))
	for _, hint := range data.Hints {
		hint.Create(ctx, resp.Diagnostics, r.client, data.ID.ValueString())
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
	res, err := r.client.GetChallenge(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read challenge %s, got error: %s", data.ID.ValueString(), err))
		return
	}
	data.Name = types.StringValue(res.Name)
	data.Category = types.StringValue(res.Category)
	data.Description = types.StringValue(res.Description)
	data.ConnectionInfo = types.StringValue(res.ConnectionInfo)
	data.MaxAttempts = types.Int64Value(int64(res.MaxAttempts))
	data.Value = toTFInt64(res.Value)
	data.Initial = toTFInt64(res.Initial)
	data.Decay = toTFInt64(res.Decay)
	data.Minimum = toTFInt64(res.Minimum)
	data.State = types.StringValue(res.State)
	data.Type = types.StringValue(res.Type)

	// Read its requirements
	resReqs, err := r.client.GetChallengeRequirements(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read requirements of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
	}
	challReqs := make([]types.String, 0, len(resReqs.Prerequisites))
	for _, req := range resReqs.Prerequisites {
		challReqs = append(challReqs, types.StringValue(strconv.Itoa(req)))
	}
	data.Requirements = requirementsSubresourceModel{
		Behavior:      fromAnon(resReqs.Anonymize),
		Prerequisites: challReqs,
	}

	// Read its files
	resFiles, err := r.client.GetChallengeFiles(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read files of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
	}
	challFiles := make([]fileSubresourceModel, 0, len(resFiles))
	for _, file := range resFiles {
		f := fileSubresourceModel{
			ID:       types.StringValue(strconv.Itoa(file.ID)),
			Location: types.StringValue(file.Location),
			Name:     types.StringValue(filename(file.Location)),
		}
		f.Read(ctx, resp.Diagnostics, r.client)
		challFiles = append(challFiles, f)
	}
	if data.Files != nil {
		data.Files = challFiles
	}

	// Read its flags
	resFlags, err := r.client.GetChallengeFlags(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read flags of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	challFlags := make([]flagSubresourceModel, 0, len(resFlags))
	for _, flag := range resFlags {
		challFlags = append(challFlags, flagSubresourceModel{
			ID:      types.StringValue(strconv.Itoa(flag.ID)),
			Content: types.StringValue(flag.Content),
			// XXX this should be typed properly
			Data: types.StringValue(flag.Data.(string)),
			Type: types.StringValue(flag.Type),
		})
	}
	if data.Flags != nil {
		data.Flags = challFlags
	}

	// Read its tags
	resTags, err := r.client.GetChallengeTags(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read tags of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	challTags := make([]types.String, 0, len(resTags))
	for _, tag := range resTags {
		challTags = append(challTags, types.StringValue(tag.Value))
	}
	if data.Tags != nil {
		data.Tags = challTags
	}

	// Read its topics
	resTopics, err := r.client.GetChallengeTopics(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read topics of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	challTopics := make([]types.String, 0, len(resTopics))
	for _, topic := range resTopics {
		challTopics = append(challTopics, types.StringValue(topic.Value))
	}
	if data.Topics != nil {
		data.Topics = challTopics
	}

	// Read its hints
	resHints, err := r.client.GetChallengeHints(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read hints of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	challHints := make([]hintSubresourceModel, 0, len(resHints))
	for _, hint := range resHints {
		hintReqs := make([]types.String, 0, len(hint.Requirements.Prerequisites))
		for _, req := range hint.Requirements.Prerequisites {
			hintReqs = append(hintReqs, types.StringValue(strconv.Itoa(req)))
		}
		if len(hint.Requirements.Prerequisites) == 0 {
			hintReqs = nil
		}
		challHints = append(challHints, hintSubresourceModel{
			ID:           types.StringValue(strconv.Itoa(hint.ID)),
			Content:      types.StringValue(*hint.Content),
			Cost:         types.Int64Value(int64(hint.Cost)),
			Requirements: hintReqs,
		})
	}
	if data.Hints != nil {
		data.Hints = challHints
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

	// Patch direct attributes
	preqs := make([]int, 0, len(data.Requirements.Prerequisites))
	for _, preq := range data.Requirements.Prerequisites {
		id, _ := strconv.Atoi(preq.ValueString())
		preqs = append(preqs, id)
	}
	_, err := r.client.PatchChallenge(data.ID.ValueString(), &api.PatchChallengeParams{
		Name:           data.Name.ValueStringPointer(),
		Category:       data.Category.ValueStringPointer(),
		Description:    data.Description.ValueStringPointer(),
		ConnectionInfo: data.ConnectionInfo.ValueStringPointer(),
		MaxAttempts:    toInt(data.MaxAttempts),
		Value:          toInt(data.Value),
		Initial:        toInt(data.Initial),
		Decay:          toInt(data.Decay),
		Minimum:        toInt(data.Minimum),
		State:          data.State.ValueStringPointer(),
		Requirements: &api.Requirements{
			Anonymize:     getAnon(data.Requirements.Behavior),
			Prerequisites: preqs,
		},
	}, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update challenge, got error: %s", err),
		)
		return
	}

	// Update its files
	currentFiles, err := r.client.GetChallengeFiles(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get challenge's files, got error: %s", err),
		)
	}
	files := []fileSubresourceModel{}
	for _, file := range data.Files {
		exists := false
		for _, currentFile := range currentFiles {
			if file.ID.ValueString() == strconv.Itoa(currentFile.ID) {
				exists = true

				// => Drop and replace iif content changed
				// TODO do it only iif content changed
				file.Delete(ctx, resp.Diagnostics, r.client)
				file.Create(ctx, resp.Diagnostics, r.client, data.ID.ValueString())

				files = append(files, file)
				break
			}
		}
		if !exists {
			file.Create(ctx, resp.Diagnostics, r.client, data.ID.ValueString())
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
			f := fileSubresourceModel{
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
	currentFlags, err := r.client.GetChallengeFlags(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get challenge's flags, got error: %s", err),
		)
		return
	}
	flags := []flagSubresourceModel{}
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
			tfFlag.Create(ctx, resp.Diagnostics, r.client, data.ID.ValueString())
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
			f := &flagSubresourceModel{
				ID: types.StringValue(strconv.Itoa(currentFlag.ID)),
			}
			f.Delete(ctx, resp.Diagnostics, r.client)
		}
	}
	if data.Flags != nil {
		data.Flags = flags
	}

	// Update its tags (drop them all, create new ones)
	challTags, err := r.client.GetChallengeTags(data.ID.ValueString(), api.WithContext(ctx))
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
			Challenge: data.ID.ValueString(),
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
	challTopics, err := r.client.GetChallengeTopics(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all topics of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, topic := range challTopics {
		if err := r.client.DeleteTopic(strconv.Itoa(topic.ID), api.WithContext(ctx)); err != nil {
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
			Challenge: data.ID.ValueString(),
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
	currentHints, err := r.client.GetChallengeHints(data.ID.ValueString(), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get challenge's hints, got error: %s", err),
		)
		return
	}
	hints := []hintSubresourceModel{}
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
			tfHint.Create(ctx, resp.Diagnostics, r.client, data.ID.ValueString())
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
			h := &hintSubresourceModel{
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

	if err := r.client.DeleteChallenge(data.ID.ValueString(), api.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete challenge, got error: %s", err))
		return
	}

	// ... don't need to delete nested objects, this is handled by CTFd
}

func (r *challengeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
