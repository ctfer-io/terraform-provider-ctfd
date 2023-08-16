package challenge

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pandatix/go-ctfd/api"
)

type fileSubresourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Location types.String `tfsdk:"location"`
	// XXX may use sha256 of file to avoid fetching it each time (fasten large-files cases, e.g. forensic dump)
	Content    types.String `tfsdk:"content"`
	ContentB64 types.String `tfsdk:"contentb64"`
}

func fileSubresourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"location": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"content": schema.StringAttribute{
			Optional:  true,
			Computed:  true,
			Sensitive: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"contentb64": schema.StringAttribute{
			Optional:  true,
			Computed:  true,
			Sensitive: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}
}

// Read fetches all the file's information, only requiring the ID to be set.
func (file *fileSubresourceModel) Read(ctx context.Context, diags diag.Diagnostics, client *api.Client) {
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
}

func (data *fileSubresourceModel) Create(ctx context.Context, diags diag.Diagnostics, client *api.Client, challengeID string) {
	// Fetch raw or base64 content prior to creating it with raw
	data.PropagateContent(ctx, diags)
	if diags.HasError() {
		return
	}

	res, err := client.PostFiles(&api.PostFilesParams{
		Challenge: challengeID,
		File: &api.InputFile{
			Name:    data.Name.ValueString(),
			Content: []byte(data.Content.ValueString()),
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
	data.Location = types.StringValue(res[0].Location)
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

func (data *fileSubresourceModel) PropagateContent(ctx context.Context, diags diag.Diagnostics) {
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
