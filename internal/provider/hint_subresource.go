package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pandatix/go-ctfd/api"
)

// TODO requirements can be set manually, but can't be automatised. Hint may be a complete resource

type hintSubresourceModel struct {
	ID           types.String   `tfsdk:"id"`
	Content      types.String   `tfsdk:"content"`
	Cost         types.Int64    `tfsdk:"cost"`
	Requirements []types.String `tfsdk:"requirements"`
}

func hintSubresourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Identifier of the hint",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"content": schema.StringAttribute{
			Required: true,
		},
		"cost": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(0),
		},
		"requirements": schema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
		},
	}
}

func (data *hintSubresourceModel) Create(ctx context.Context, diags diag.Diagnostics, client *api.Client, challengeID string) {
	preq := make([]int, 0, len(data.Requirements))
	for _, req := range data.Requirements {
		hintID, _ := strconv.Atoi(req.ValueString())
		preq = append(preq, hintID)
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

func (data *hintSubresourceModel) Update(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	preq := make([]int, 0, len(data.Requirements))
	for _, req := range data.Requirements {
		hintID, _ := strconv.Atoi(req.ValueString())
		preq = append(preq, hintID)
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
	dPreq := make([]types.String, 0, len(res.Requirements.Prerequisites))
	for _, resPreq := range res.Requirements.Prerequisites {
		dPreq = append(dPreq, types.StringValue(strconv.Itoa(resPreq)))
	}
	data.Requirements = dPreq
}

func (data *hintSubresourceModel) Delete(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	if err := client.DeleteHint(data.ID.ValueString(), api.WithContext(ctx)); err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete hint, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a hint")
}
