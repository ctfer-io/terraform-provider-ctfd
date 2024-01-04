package challenge

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type FlagSubresourceModel struct {
	ID      types.String `tfsdk:"id"`
	Content types.String `tfsdk:"content"`
	Data    types.String `tfsdk:"data"`
	Type    types.String `tfsdk:"type"`
}

func FlagSubresourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Identifier of the flag, used internally to handle the CTFd corresponding object.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"content": schema.StringAttribute{
			MarkdownDescription: "The actual flag to match. Consider using the convention `MYCTF{value}` with `MYCTF` being the shortcode of your event's name and `value` depending on each challenge.",
			Required:            true,
			Sensitive:           true,
		},
		"data": schema.StringAttribute{
			MarkdownDescription: "The flag sensitivity information, either case_sensitive or case_insensitive",
			Optional:            true,
			Computed:            true,
			// default value is "" (empty string) according to Web UI
			Default: stringdefault.StaticString("case_sensitive"),
			Validators: []validator.String{
				validators.NewStringEnumValidator([]basetypes.StringValue{
					types.StringValue("case_sensitive"),
					types.StringValue("case_insensitive"),
				}),
			},
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "The type of the flag, could be either static or regex",
			Optional:            true,
			Computed:            true,
			// default value is "static" according to ctfcli
			Default: stringdefault.StaticString("static"),
			Validators: []validator.String{
				validators.NewStringEnumValidator([]basetypes.StringValue{
					types.StringValue("static"),
					types.StringValue("regex"),
				}),
			},
		},
	}
}

func (data *FlagSubresourceModel) Create(ctx context.Context, diags diag.Diagnostics, client *api.Client, challengeID int) {
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

func (data *FlagSubresourceModel) Update(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
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
	data.Data = types.StringValue(res.Data)
	data.Type = types.StringValue(res.Type)
}

func (data *FlagSubresourceModel) Delete(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	if err := client.DeleteFlag(data.ID.ValueString(), api.WithContext(ctx)); err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete flag, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a flag")
}
