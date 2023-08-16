package challenge

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pandatix/go-ctfd/api"
)

type flagSubresourceModel struct {
	ID      types.String `tfsdk:"id"`
	Content types.String `tfsdk:"content"`
	Data    types.String `tfsdk:"data"`
	Type    types.String `tfsdk:"type"`
}

func flagSubresourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Identifier of the flag",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"content": schema.StringAttribute{
			Required:  true,
			Sensitive: true,
		},
		"data": schema.StringAttribute{
			Optional: true,
			Computed: true,
			// default value is "" (empty string) according to Web UI
			Default: stringdefault.StaticString(""),
		},
		"type": schema.StringAttribute{
			Optional: true,
			Computed: true,
			// default value is "static" according to ctfcli
			Default: stringdefault.StaticString("static"),
		},
	}
}

func (data *flagSubresourceModel) Create(ctx context.Context, diags diag.Diagnostics, client *api.Client, challengeID string) {
	res, err := client.PostFlags(&api.PostFlagsParams{
		Challenge: challengeID,
		Content:   data.Content.ValueString(),
		Data:      data.Data.ValueString(),
		Type:      data.Type.ValueString(),
	}, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create flag, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a flag")

	data.ID = types.StringValue(strconv.Itoa(res.ID))
}

func (data *flagSubresourceModel) Update(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	res, err := client.PatchFlag(data.ID.ValueString(), &api.PatchFlagParams{
		Content: data.Content.ValueString(),
		Data:    data.Data.ValueString(),
		Type:    data.Type.ValueString(),
	}, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update flag, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "updated a flag")

	data.Content = types.StringValue(res.Content)
	// XXX this should be properly typed
	data.Data = types.StringValue(res.Data.(string))
	data.Type = types.StringValue(res.Type)
}

func (data *flagSubresourceModel) Delete(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	if err := client.DeleteFlag(data.ID.ValueString(), api.WithContext(ctx)); err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete flag, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a flag")
}
