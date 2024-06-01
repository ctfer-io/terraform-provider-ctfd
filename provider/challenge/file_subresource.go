package challenge

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type FileSubresourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Location   types.String `tfsdk:"location"`
	SHA1Sum    types.String `tfsdk:"sha1sum"`
	Content    types.String `tfsdk:"content"`
	ContentB64 types.String `tfsdk:"contentb64"`
}

func FileSubresourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Identifier of the file, used internally to handle the CTFd corresponding object.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the file as displayed to end-users.",
			Required:            true,
		},
		"location": schema.StringAttribute{
			MarkdownDescription: "Location where the file is stored on the CTFd instance, for download purposes.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"sha1sum": schema.StringAttribute{
			MarkdownDescription: "The sha1 sum of the file.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"content": schema.StringAttribute{
			MarkdownDescription: "Raw content of the file, perfectly fit the use-cases of a .txt document or anything with a simple binary content. You could provide it from the file-system using `file(\"${path.module}/...\")`.",
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"contentb64": schema.StringAttribute{
			MarkdownDescription: "Base 64 content of the file, perfectly fit the use-cases of complex binaries. You could provide it from the file-system using `filebase64(\"${path.module}/...\")`.",
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// Read fetches all the file's information, only requiring the ID to be set.
func (file *FileSubresourceModel) Read(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	content, err := client.GetFileContent(&api.File{
		Location: file.Location.ValueString(),
	}, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"CTFd Error",
			fmt.Sprintf("Unable to read files at location %s, got error: %s", file.Location, err),
		)
	}

	file.Content = types.StringValue(string(content))
	file.PropagateContent(ctx, diags)

	h := sha1.New()
	_, err = h.Write(content)
	if err != nil {
		diags.AddError(
			"Internal Error",
			fmt.Sprintf("Failed to compute SHA1 sum, got error: %s", err),
		)
	}
	sum := h.Sum(nil)
	file.SHA1Sum = types.StringValue(hex.EncodeToString(sum))
}

func (data *FileSubresourceModel) Create(ctx context.Context, diags diag.Diagnostics, client *api.Client, challengeID int) {
	// Fetch raw or base64 content prior to creating it with raw
	data.PropagateContent(ctx, diags)
	if diags.HasError() {
		return
	}

	res, err := client.PostFiles(&api.PostFilesParams{
		Challenge: &challengeID,
		Files: []*api.InputFile{
			{
				Name:    data.Name.ValueString(),
				Content: []byte(data.Content.ValueString()),
			},
		},
		Location: data.Location.ValueStringPointer(),
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
	data.SHA1Sum = types.StringValue(res[0].SHA1sum)
	data.Location = types.StringValue(res[0].Location)
}

func (data *FileSubresourceModel) Delete(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
	if err := client.DeleteFile(data.ID.ValueString(), api.WithContext(ctx)); err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete file %s, got error: %s", data.Name, err),
		)
		return
	}

	tflog.Trace(ctx, "deleted a file")
}

func (data *FileSubresourceModel) PropagateContent(ctx context.Context, diags diag.Diagnostics) {
	// If the other content source is set, get the other from it
	if len(data.Content.ValueString()) != 0 {
		cb64 := base64.StdEncoding.EncodeToString([]byte(data.Content.ValueString()))
		data.ContentB64 = types.StringValue(cb64)
		return
	}
	if len(data.ContentB64.ValueString()) != 0 {
		c, err := base64.StdEncoding.DecodeString(data.ContentB64.ValueString())
		diags.AddError(
			"File Error",
			fmt.Sprintf("Base64 file content failed at decoding: %s", err),
		)
		data.Content = types.StringValue(string(c))
		return
	}
	// If no content seems to be set, set them both empty
	data.Content = types.StringValue("")
	data.ContentB64 = types.StringValue("")
}
