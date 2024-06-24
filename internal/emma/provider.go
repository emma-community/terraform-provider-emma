package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"

	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &Provider{}
)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &Provider{}
	}
}

type providerModel struct {
	Host         types.String `tfsdk:"host"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// Provider is the provider implementation.
type Provider struct {
}

// Metadata returns the provider type name.
func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "emma"
}

// Schema defines the provider-level schema for configuration data.
func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This Terraform Provider Emma " +
			"allows you to manage multi-cloud resources. The [emma platform](https://www.emma.ms/) empowers you to " +
			"effortlessly deploy and manage cloud resources across diverse environments, spanning on-premises, " +
			"private, and public clouds. Whether you're a seasoned cloud professional honing your multi-cloud setup " +
			"or diving into cloud management for the first time, our cloud-agnostic approach guarantees freedom to " +
			"leverage the right cloud services you need.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
				Required: false,
			},
			"client_id": schema.StringAttribute{
				Optional:    false,
				Required:    true,
				Description: "Client ID from the Service application in the project",
			},
			"client_secret": schema.StringAttribute{
				Optional:    false,
				Required:    true,
				Sensitive:   true,
				Description: "Client secret from the Service application in the project",
			},
		},
	}
}

// Configure prepares a EMMA API client for data sources and resources.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	tflog.Info(ctx, "Configuring EMMA client")
	var config providerModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.ClientId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("clientId"),
			"Unknown EMMA API ClientId",
			"The provider cannot create the EMMA API client as there is an unknown configuration value for the EMMA API clientId. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the EMMA_CLIENT_ID environment variable.")
	}

	if config.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("clientSecret"),
			"Unknown EMMA API ClientSecret",
			"The provider cannot create the EMMA API client as there is an unknown configuration value for the EMMA API clientSecret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the EMMA_CLIENT_SECRET environment variable.")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	clientId := os.Getenv("EMMA_CLIENT_ID")
	clientSecret := os.Getenv("EMMA_CLIENT_SECRET")

	if !config.Host.IsNull() {

	}

	if !config.ClientId.IsNull() {
		clientId = config.ClientId.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if clientId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("clientId"),
			"Missing EMMA API Username",
			"The provider cannot create the EMMA API client as there is a missing or empty value for the EMMA API clientId. "+
				"Set the clientId value in the configuration or use the EMMA_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.")
	}

	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("clientSecret"),
			"Missing EMMA API Password",
			"The provider cannot create the EMMA API client as there is a missing or empty value for the EMMA API clientSecret. "+
				"Set the clientSecret value in the configuration or use the EMMA_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.")
	}
	host := os.Getenv("EMMA_HOST")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	if host == "" {
		host = emmaSdk.NewConfiguration().Servers[0].URL
	}

	if resp.Diagnostics.HasError() {
		return
	}

	configuration := &emmaSdk.Configuration{
		DefaultHeader: make(map[string]string),
		UserAgent:     "OpenAPI-Generator/0.0.1/go",
		Debug:         false,
		Servers: emmaSdk.ServerConfigurations{
			{
				URL:         host,
				Description: "Public EMMA API",
			},
		},
		OperationServers: map[string]emmaSdk.ServerConfigurations{},
	}
	apiClient := emmaSdk.NewAPIClient(configuration)
	credentials := emmaSdk.Credentials{ClientId: clientId, ClientSecret: clientSecret}
	// Create a new EMMA client using the configuration values
	token, _, err := apiClient.AuthenticationAPI.IssueToken(ctx).Credentials(credentials).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to authenticate EMMA API Client",
			"An unexpected error occurred when creating the EMMA API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"EMMA Client Error: "+err.Error())
		return
	}
	providerClient := Client{apiClient: apiClient, token: token}
	tflog.Info(ctx, "Configured EMMA client")
	// Make the EMMA client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = &providerClient
	resp.ResourceData = &providerClient
}

// DataSources defines the data sources implemented in the provider.
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDataCenterDataSource,
		NewLocationDataSource,
		NewOperatingSystemDataSource,
		NewProviderDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVmResource,
		NewSshKeyResource,
		NewSecurityGroupResource,
		NewSpotInstanceResource,
	}
}

type Client struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}
