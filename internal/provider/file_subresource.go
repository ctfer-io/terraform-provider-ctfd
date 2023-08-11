package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pandatix/go-ctfd/api"
)

type fileSubresourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	// XXX may use sha256 of file to avoid fetching it each time (fasten large-files cases, e.g. forensic dump)
	Content    types.String `tfsdk:"content"`
	ContentB64 types.String `tfsdk:"contentb64"`
}

func fileSubresourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"content": schema.StringAttribute{
			Optional: true,
		},
		"contentb64": schema.StringAttribute{
			Optional: true,
		},
	}
}

func (data *fileSubresourceModel) Create(ctx context.Context, diags diag.Diagnostics, client *api.Client, challengeID string) {
	content := data.GetContent(diags)
	if diags.HasError() {
		return
	}
	res, err := client.PostFiles(&api.PostFilesParams{
		Challenge: challengeID,
		File: &api.InputFile{
			Name:    data.Name.ValueString(),
			Content: content,
		},
	}, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create file, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a file")

	data.ID = types.StringValue(strconv.Itoa(res[0].ID))
}

func (data *fileSubresourceModel) Delete(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	if err := client.DeleteFile(data.ID.ValueString(), api.WithContext(ctx)); err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete file %s, got error: %s", data.Name, err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a file")
}

func (data *fileSubresourceModel) GetContent(diags diag.Diagnostics) []byte {
	switch {
	case !data.Content.IsNull():
		return []byte(data.Content.ValueString())
	case !data.Content.IsNull():
		content, err := base64.RawStdEncoding.DecodeString(data.ContentB64.ValueString())
		diags.AddError(
			"File base64 Error",
			fmt.Sprintf("Base64 file content failed at decoding: %s", err),
		)
		return content
	default:
		diags.AddError(
			"File datamodel Error",
			"Either .content or .contentb64 should be defined.",
		)
		return nil
	}
}
