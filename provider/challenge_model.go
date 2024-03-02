package provider

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/provider/challenge"
	"github.com/ctfer-io/terraform-provider-ctfd/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type challengeResourceModel struct {
	ID             types.String                            `tfsdk:"id"`
	Name           types.String                            `tfsdk:"name"`
	Category       types.String                            `tfsdk:"category"`
	Description    types.String                            `tfsdk:"description"`
	ConnectionInfo types.String                            `tfsdk:"connection_info"`
	MaxAttempts    types.Int64                             `tfsdk:"max_attempts"`
	Function       types.String                            `tfsdk:"function"`
	Value          types.Int64                             `tfsdk:"value"`
	Decay          types.Int64                             `tfsdk:"decay"`
	Minimum        types.Int64                             `tfsdk:"minimum"`
	State          types.String                            `tfsdk:"state"`
	Type           types.String                            `tfsdk:"type"`
	Next           types.Int64                             `tfsdk:"next"`
	Requirements   *challenge.RequirementsSubresourceModel `tfsdk:"requirements"`
	Flags          []challenge.FlagSubresourceModel        `tfsdk:"flags"`
	Tags           []types.String                          `tfsdk:"tags"`
	Topics         []types.String                          `tfsdk:"topics"`
	Hints          []challenge.HintSubresourceModel        `tfsdk:"hints"`
	Files          []challenge.FileSubresourceModel        `tfsdk:"files"`
}

func (chall *challengeResourceModel) Read(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	// Retrieve challenge
	res, err := client.GetChallenge(utils.Atoi(chall.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read challenge %s, got error: %s", chall.ID.ValueString(), err))
		return
	}
	chall.Name = types.StringValue(res.Name)
	chall.Category = types.StringValue(res.Category)
	chall.Description = types.StringValue(res.Description)
	chall.ConnectionInfo = utils.ToTFString(res.ConnectionInfo)
	chall.MaxAttempts = utils.ToTFInt64(res.MaxAttempts)
	chall.Function = types.StringValue("linear") // XXX CTFd does not return the `function` attribute
	chall.Decay = utils.ToTFInt64(res.Decay)
	chall.Minimum = utils.ToTFInt64(res.Minimum)
	chall.State = types.StringValue(res.State)
	chall.Type = types.StringValue(res.Type)
	chall.Next = utils.ToTFInt64(res.NextID)

	switch res.Type {
	case "standard":
		chall.Value = types.Int64Value(int64(res.Value))
	case "dynamic":
		chall.Value = utils.ToTFInt64(res.Initial)
	}

	id := utils.Atoi(chall.ID.ValueString())

	// Get subresources
	// => Requirements
	resReqs, err := client.GetChallengeRequirements(id, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
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
	chall.Requirements = reqs

	// => Files
	resFiles, err := client.GetChallengeFiles(id, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d files, got error: %s", id, err),
		)
		return
	}
	chall.Files = make([]challenge.FileSubresourceModel, 0, len(resFiles))
	for _, file := range resFiles {
		c, err := client.GetFileContent(file, api.WithContext(ctx))
		if err != nil {
			diags.AddError(
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
		nf.PropagateContent(ctx, diags)
		h := sha1.New()
		_, err = h.Write(c)
		if err != nil {
			diags.AddError(
				"Internal Error",
				fmt.Sprintf("Failed to compute SHA1 sum, got error: %s", err),
			)
		}
		sum := h.Sum(nil)
		nf.SHA1Sum = types.StringValue(hex.EncodeToString(sum))
		chall.Files = append(chall.Files, nf)
	}

	// => Flags
	resFlags, err := client.GetChallengeFlags(id, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d flags, got error: %s", id, err),
		)
		return
	}
	chall.Flags = make([]challenge.FlagSubresourceModel, 0, len(resFlags))
	for _, flag := range resFlags {
		chall.Flags = append(chall.Flags, challenge.FlagSubresourceModel{
			ID:      types.StringValue(strconv.Itoa(flag.ID)),
			Content: types.StringValue(flag.Content),
			Data:    types.StringValue(flag.Data),
			Type:    types.StringValue(flag.Type),
		})
	}

	// => Hints
	resHints, err := client.GetChallengeHints(id, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d hints, got error: %s", id, err),
		)
		return
	}
	chall.Hints = make([]challenge.HintSubresourceModel, 0, len(resHints))
	for _, hint := range resHints {
		reqs := []attr.Value{}
		if hint.Requirements != nil {
			reqs = make([]attr.Value, 0, len(hint.Requirements.Prerequisites))
			for _, req := range hint.Requirements.Prerequisites {
				reqs = append(reqs, types.StringValue(strconv.Itoa(req)))
			}
		}
		chall.Hints = append(chall.Hints, challenge.HintSubresourceModel{
			ID:           types.StringValue(strconv.Itoa(hint.ID)),
			Content:      types.StringValue(*hint.Content),
			Cost:         types.Int64Value(int64(hint.Cost)),
			Requirements: types.ListValueMust(types.StringType, reqs),
		})
	}

	// => Tags
	resTags, err := client.GetChallengeTags(id, api.WithContext(ctx))
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
	resTopics, err := client.GetChallengeTopics(id, api.WithContext(ctx))
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
