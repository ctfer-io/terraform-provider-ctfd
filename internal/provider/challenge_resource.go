package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

type challengeResourceModel struct {
	ID          types.String           `tfsdk:"id"`
	Name        types.String           `tfsdk:"name"`
	Category    types.String           `tfsdk:"category"`
	Description types.String           `tfsdk:"description"`
	Value       types.Int64            `tfsdk:"value"`
	Initial     types.Int64            `tfsdk:"initial"`
	Decay       types.Int64            `tfsdk:"decay"`
	Minimum     types.Int64            `tfsdk:"minimum"`
	State       types.String           `tfsdk:"state"`
	Type        types.String           `tfsdk:"type"`
	Flags       []flagSubresourceModel `tfsdk:"flags"`
	Tags        []types.String         `tfsdk:"tags"`
	Topics      []types.String         `tfsdk:"topics"`
	Hints       []hintSubresourceModel `tfsdk:"hints"`
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
			// TODO value can't be set side to <initial,decay,minimum>
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
			"flags": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: flagSubresourceAttributes(),
				},
				Optional: true,
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"topics": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"hints": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: hintSubresourceAttributes(),
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
	res, err := r.client.PostChallenges(&api.PostChallengesParams{
		Name:        data.Name.ValueString(),
		Category:    data.Category.ValueString(),
		Description: data.Description.ValueString(),
		Value:       toInt(data.Value),
		Initial:     toInt(data.Initial),
		Decay:       toInt(data.Decay),
		Minimum:     toInt(data.Minimum),
		State:       data.State.ValueString(),
		Type:        data.Type.ValueString(),
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
	data.Value = toTFInt64(res.Value)

	// Create flags
	challFlags := make([]flagSubresourceModel, 0, len(data.Flags))
	for _, flag := range data.Flags {
		flag.Create(ctx, resp.Diagnostics, r.client, data.ID.ValueString())
		challFlags = append(challFlags, flag)
	}
	data.Flags = challFlags

	// Create tags
	challTags := make([]types.String, 0, len(data.Tags))
	for _, tag := range data.Tags {
		_, err := r.client.PostTags(&api.PostTagsParams{
			Challenge: data.ID.ValueString(),
			Value:     tag.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create topic, got error: %s", err),
			)
			return
		}
		challTags = append(challTags, tag)
	}
	data.Tags = challTags

	// Create topics
	challTopics := make([]types.String, 0, len(data.Topics))
	for _, topic := range data.Topics {
		_, err := r.client.PostTopics(&api.PostTopicsParams{
			Challenge: data.ID.ValueString(),
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
	data.Topics = challTopics

	// Create hints
	challHints := make([]hintSubresourceModel, 0, len(data.Hints))
	for _, hint := range data.Hints {
		hint.Create(ctx, resp.Diagnostics, r.client, data.ID.ValueString())
		challHints = append(challHints, hint)
	}
	data.Hints = challHints

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
	res, err := r.client.GetChallenge(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read challenge %s, got error: %s", data.ID.ValueString(), err))
		return
	}
	data.Name = types.StringValue(res.Name)
	data.Category = types.StringValue(res.Category)
	data.Description = types.StringValue(res.Description)
	data.Value = toTFInt64(res.Value)
	data.Initial = toTFInt64(res.Initial)
	data.Decay = toTFInt64(res.Decay)
	data.Minimum = toTFInt64(res.Minimum)
	data.State = types.StringValue(res.State)
	data.Type = types.StringValue(res.Type)

	// Read its flags
	resFlags, err := r.client.GetChallengeFlags(data.ID.ValueString())
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
	data.Flags = challFlags

	// Read its tags
	resTags, err := r.client.GetChallengeTags(data.ID.ValueString())
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
	data.Tags = challTags

	// Read its topics
	resTopics, err := r.client.GetChallengeTopics(data.ID.ValueString())
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
	data.Topics = challTopics

	// Read its hints
	resHints, err := r.client.GetChallengeHints(data.ID.ValueString())
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
		challHints = append(challHints, hintSubresourceModel{
			ID:           types.StringValue(strconv.Itoa(hint.ID)),
			Content:      types.StringValue(*hint.Content),
			Cost:         types.Int64Value(int64(hint.Cost)),
			Requirements: hintReqs,
		})
	}
	data.Hints = challHints

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

	// XXX Type can't be modified after creation, needs to delete the challenge
	res, err := r.client.PatchChallenge(data.ID.ValueString(), &api.PatchChallengeParams{
		Name:        data.Name.ValueStringPointer(),
		Category:    data.Category.ValueStringPointer(),
		Description: data.Description.ValueStringPointer(),
		Initial:     toInt(data.Initial),
		Decay:       toInt(data.Decay),
		Minimum:     toInt(data.Minimum),
		State:       data.State.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update challenge, got error: %s", err))
		return
	}

	data.Name = types.StringValue(res.Name)
	data.Category = types.StringValue(res.Category)
	data.Description = types.StringValue(res.Description)
	data.Value = toTFInt64(res.Value)
	data.Initial = toTFInt64(res.Initial)
	data.Decay = toTFInt64(res.Decay)
	data.Minimum = toTFInt64(res.Minimum)
	data.State = types.StringValue(res.State)
	data.Type = types.StringValue(res.Type)

	// Update its flags
	currentFlags, err := r.client.GetFlags(&api.GetFlagsParams{
		ChallengeID: data.ID.ValueStringPointer(),
	})
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
	data.Flags = flags

	// Update its tags (drop them all, create new ones)
	challTags, err := r.client.GetChallengeTags(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all tags of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, tag := range challTags {
		if err := r.client.DeleteTag(strconv.Itoa(tag.ID)); err != nil {
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
	data.Tags = tags

	// Update its topics (drop them all, create new ones)
	challTopics, err := r.client.GetChallengeTopics(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all topics of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, topic := range challTopics {
		if err := r.client.DeleteTopic(strconv.Itoa(topic.ID)); err != nil {
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
	data.Topics = topics

	// Update its hints
	currentHints, err := r.client.GetChallengeHints(data.ID.ValueString())
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
	data.Hints = hints

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

	if err := r.client.DeleteChallenge(data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete challenge, got error: %s", err))
		return
	}

	// ... don't need to delete nested objects, this is handled by CTFd
}

func (r *challengeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
