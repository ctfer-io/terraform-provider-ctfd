package provider

import (
	"context"
	"os"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = (*CTFdProvider)(nil)

type CTFdProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CTFdProvider{
			version: version,
		}
	}
}

type CTFdProviderModel struct {
	URL     types.String `tfsdk:"url"`
	Session types.String `tfsdk:"session"`
	Nonce   types.String `tfsdk:"nonce"`
	APIKey  types.String `tfsdk:"api_key"`
}

func (p *CTFdProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ctfd"
	resp.Version = p.version
}

func (p *CTFdProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Interact with a [CTFd](https://github.com/ctfd/ctfd).",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "CTFd base URL. Could use `CTFD_URL` environment variable.",
				Optional:            true,
			},
			"session": schema.StringAttribute{
				MarkdownDescription: "User session token, comes with nonce. Could use `CTFD_SESSION` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"nonce": schema.StringAttribute{
				MarkdownDescription: "User session nonce, comes with session. Could use `CTFD_NONCE` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "User API key. Could use `CTFD_API_KEY` environment variable. Despite being the most convenient way to authenticate yourself, we do not recommend it as you will probably generate a long-live token without any rotation policy.",
				Sensitive:           true,
				Optional:            true,
			},
		},
	}
}

func (p *CTFdProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config CTFdProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check configuration values are known
	if config.Session.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("session"),
			"Unknown CTFd session.",
			"The provider cannot create the CTFd API client as there is an unknown session value.",
		)
	}
	if config.Nonce.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("nonce"),
			"Unknown CTFd nonce.",
			"The provider cannot create the CTFd API client as there is an unknown nonce value.",
		)
	}
	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown CTFd API key.",
			"The provider cannot create the CTFd API client as there is an unknown API key value.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract environment variables values
	url := os.Getenv("CTFD_URL")
	session := os.Getenv("CTFD_SESSION")
	nonce := os.Getenv("CTFD_NONCE")
	apiKey := os.Getenv("CTFD_API_KEY")

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}
	if !config.Session.IsNull() {
		session = config.Session.ValueString()
	}
	if !config.Nonce.IsNull() {
		nonce = config.Nonce.ValueString()
	}
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	// Check there is enough content
	if apiKey == "" {
		if session == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("session"),
				"Missing CTFd session",
				"The provider cannot create the CTFd API client as there is a missing value for the CTFd API session, as the API key is not defined.",
			)
		}
		if nonce == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("nonce"),
				"Missing CTFd nonce",
				"The provider cannot create the CTFd API client as there is a missing value for the CTFd API nonce, as the API key is not defined.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Instantiate CTFd API client
	ctx = tflog.SetField(ctx, "ctfd_url", url)
	ctx = addSensitive(ctx, "ctfd_session", session)
	ctx = addSensitive(ctx, "ctfd_nonce", nonce)
	ctx = addSensitive(ctx, "ctfd_api_key", apiKey)
	tflog.Debug(ctx, "Creating CTFd API client")

	client := api.NewClient(url, session, nonce, apiKey)
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configure CTFd API client", map[string]any{
		"success": true,
	})
}

func (p *CTFdProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewChallengeResource,
	}
}

func (p *CTFdProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewChallengeDataSource,
	}
}
