package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/utils"
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
	URL      types.String `tfsdk:"url"`
	APIKey   types.String `tfsdk:"api_key"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *CTFdProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ctfd"
	resp.Version = p.version
}

func (p *CTFdProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Use the Terraform Provider to interact with a [CTFd](https://github.com/ctfd/ctfd).

## Why creating this ?

Terraform is used to manage resources that have lifecycles, configurations, to sum it up.

That is the case of CTFd: it handles challenges that could be created, modified and deleted.
With some work to leverage the unsteady CTFd's API, Terraform is now able to manage them as cloud resources bringing you to opportunity of CTF as Code.

With a paradigm-shifting vision of setting up CTFs, the Terraform Provider for CTFd avoid shitty scripts, ` + "`ctfcli`" + ` and other tools that does not solve the problem of reproductibility, ease of deployment and resiliency.

## Authentication

You must configure the provider with the proper credentials before you can use it.

If you are using the username/password configuration, remember that CTFd comes with a
ratelimiter on rare methods and endpoints, but ` + "`POST /login`" + ` is one of them.
This could lead to unexpected failures under intensive work.

!> **Warning:** Hard-coded credentials are not recommended in any Terraform
configuration and risks secret leakage should this file ever be committed to a
public version control system.
`,
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "CTFd base URL (e.g. `https://my-ctf.lan`). Could use `CTFD_URL` environment variable instead.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "User API key. Could use `CTFD_API_KEY` environment variable instead. Despite being the most convenient way to authenticate yourself, we do not recommend it as you will probably generate a long-live token without any rotation policy.",
				Sensitive:           true,
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: `The administrator or service account username to login with. Could use ` + "`CTFD_ADMIN_USERNAME`" + ` environment variable instead.`,
				Sensitive:           true,
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The administrator or service account password to login with. Could use `CTFD_ADMIN_PASSWORD` environment variable instead.",
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
	if config.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown CTFD url.",
			"The provider cannot guess where to reach the CTFd instance.",
		)
	}
	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown CTFd API key.",
			"The provider cannot create the CTFd API client as there is an unknown API key value.",
		)
	}
	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown CTFd admin or service account username.",
			"The provider cannot create the CTFd API client as there is an unknown username.",
		)
	}
	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown CTFd admin or service account password.",
			"The provider cannot create the CTFd API client as there is an unknown password.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract environment variables values
	url := os.Getenv("CTFD_URL")
	apiKey := os.Getenv("CTFD_API_KEY")
	username := os.Getenv("CTFD_ADMIN_USERNAME")
	password := os.Getenv("CTFD_ADMIN_PASSWORD")

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// Check there is enough content
	ak := apiKey != ""
	up := username != "" && password != ""
	if !ak && !up {
		resp.Diagnostics.AddError(
			"CTFd provider configuration error",
			"The provider cannot create the CTFd API client as there is an invalid configuration. Expected either an API key, a nonce and session, or a username and password.",
		)
		return
	}

	// Instantiate CTFd API client
	ctx = tflog.SetField(ctx, "ctfd_url", url)
	ctx = utils.AddSensitive(ctx, "ctfd_api_key", apiKey)
	ctx = utils.AddSensitive(ctx, "ctfd_username", username)
	ctx = utils.AddSensitive(ctx, "ctfd_password", password)
	tflog.Debug(ctx, "Creating CTFd API client")

	nonce, session, err := api.GetNonceAndSession(url, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"CTFd error",
			fmt.Sprintf("Failed to fetch nonce and session: %s", err),
		)
		return
	}

	client := api.NewClient(url, nonce, session, apiKey)
	if up {
		// XXX due to the CTFd ratelimiter on rare endpoint
		if _, ok := os.LookupEnv("TF_ACC"); ok {
			time.Sleep(5 * time.Second)
		}

		if err := client.Login(&api.LoginParams{
			Name:     username,
			Password: password,
		}, api.WithContext(ctx)); err != nil {
			resp.Diagnostics.AddError(
				"CTFd error",
				fmt.Sprintf("Failed to login: %s", err),
			)
			return
		}
	}

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configure CTFd API client", map[string]any{
		"success": true,
		"login":   up,
	})
}

func (p *CTFdProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewChallengeStandardResource,
		NewChallengeDynamicResource,
		NewHintResource,
		NewFlagResource,
		NewFileResource,
		NewUserResource,
		NewTeamResource,
	}
}

func (p *CTFdProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewChallengeStandardDataSource,
		NewChallengeDynamicDataSource,
		NewUserDataSource,
		NewTeamDataSource,
	}
}
