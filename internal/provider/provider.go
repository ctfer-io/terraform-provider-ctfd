package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pandatix/go-ctfd/api"
	"github.com/pandatix/terraform-provider-ctfd/internal/provider/challenge"
)

var _ provider.Provider = &CTFdProvider{}

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
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "CTFd base URL.",
				Required:            true,
			},
			"session": schema.StringAttribute{
				MarkdownDescription: "User session token, comes with nonce. Could use CTFD_SESSION environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"nonce": schema.StringAttribute{
				MarkdownDescription: "User session nonce, comes with session. Could use CTFD_NONCE environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "User API key. Could use CTFD_API_KEY environment variable.",
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
	session := os.Getenv("CTFD_SESSION")
	nonce := os.Getenv("CTFD_NONCE")
	apiKey := os.Getenv("CTFD_API_KEY")

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
	ctx = tflog.SetField(ctx, "ctfd_url", config.URL.ValueString())
	ctx = addSensitive(ctx, "ctfd_session", session)
	ctx = addSensitive(ctx, "ctfd_nonce", nonce)
	ctx = addSensitive(ctx, "ctfd_api_key", apiKey)
	tflog.Debug(ctx, "Creating CTFd API client")

	client := api.NewClient(config.URL.ValueString(), session, nonce, apiKey)
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configure CTFd API client", map[string]any{
		"success": true,
	})
}

func (p *CTFdProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		challenge.NewChallengeResource,
	}
}

func (p *CTFdProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		challenge.NewChallengeDataSource,
	}
}
