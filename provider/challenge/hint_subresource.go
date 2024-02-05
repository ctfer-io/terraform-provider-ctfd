package challenge

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type HintSubresourceModel struct {
	ID           types.String `tfsdk:"id"`
	Content      types.String `tfsdk:"content"`
	Cost         types.Int64  `tfsdk:"cost"`
	Requirements types.List   `tfsdk:"requirements"`
}

func HintSubresourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Identifier of the hint, used internally to handle the CTFd corresponding object.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"content": schema.StringAttribute{
			MarkdownDescription: "Content of the hint as displayed to the end-user.",
			Required:            true,
		},
		"cost": schema.Int64Attribute{
			MarkdownDescription: "Cost of the hint, and if any specified, the end-user will consume its own (or team) points to get it.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(0),
		},
		"requirements": schema.ListAttribute{
			MarkdownDescription: "Other hints required to be consumed before getting this one. Useful for cost-increasing hint strategies with more and more help.",
			ElementType:         types.StringType,
			Default:             listdefault.StaticValue(basetypes.ListValue{}),
			Computed:            true,
			Optional:            true,
		},
	}
}

func (data *HintSubresourceModel) Create(ctx context.Context, diags diag.Diagnostics, client *api.Client, challengeID int) {
	preq := make([]int, 0, len(data.Requirements.Elements()))
	for _, req := range data.Requirements.Elements() {
		// TODO use strconv.Atoi and handle error properly
		reqid := utils.Atoi(req.(types.String).ValueString())
		preq = append(preq, reqid)
	}

	res, err := client.PostHints(&api.PostHintsParams{
		ChallengeID: challengeID,
		Content:     data.Content.ValueString(),
		Cost:        int(data.Cost.ValueInt64()),
		Requirements: api.Requirements{
			Prerequisites: preq,
		},
	}, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create hint, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a hint")

	data.ID = types.StringValue(strconv.Itoa(res.ID))
}

func (data *HintSubresourceModel) Update(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	preq := make([]int, 0, len(data.Requirements.Elements()))
	for _, req := range data.Requirements.Elements() {
		// TODO use strconv.Atoi and handle error properly
		reqid := utils.Atoi(req.(types.String).ValueString())
		preq = append(preq, reqid)
	}

	res, err := client.PatchHint(data.ID.ValueString(), &api.PatchHintsParams{
		Content: data.Content.ValueString(),
		Cost:    int(data.Cost.ValueInt64()),
		Requirements: api.Requirements{
			Prerequisites: preq,
		},
	}, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update hint, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "updated a hint")

	data.Content = types.StringValue(*res.Content)
	data.Cost = types.Int64Value(int64(res.Cost))
	dPreq := make([]attr.Value, 0, len(res.Requirements.Prerequisites))
	for _, resPreq := range res.Requirements.Prerequisites {
		dPreq = append(dPreq, types.StringValue(strconv.Itoa(resPreq)))
	}
	data.Requirements = types.ListValueMust(types.StringType, dPreq)
}

func (data *HintSubresourceModel) Delete(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	if err := client.DeleteHint(data.ID.ValueString(), api.WithContext(ctx)); err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete hint, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a hint")
}
