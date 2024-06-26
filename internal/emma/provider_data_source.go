package emma

import (
	"context"
	"fmt"
	"github.com/emma-community/terraform-provider-emma/tools"

	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &providerDataSource{}

func NewProviderDataSource() datasource.DataSource {
	return &providerDataSource{}
}

// providerDataSource defines the data source implementation.
type providerDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// providerDataSourceModel describes the data source data model.
type providerDataSourceModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *providerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

func (d *providerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Cloud service provider is a third-party company that provides scalable computing resources that " +
			"users can access over emma platform. When you create any computing resources in emma you create them in " +
			"the provider's datacenters. Provider is necessary when you select a data center for creating compute instances.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "ID of the cloud provider",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the cloud provider",
				Computed:    false,
				Required:    true,
				Optional:    false,
			},
		},
	}
}

func (d *providerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData))
		return
	}
	d.apiClient = client.apiClient
	d.token = client.token
}

func (d *providerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data providerDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Read provider")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	request := d.apiClient.ProvidersAPI.GetProviders(auth)
	request = request.ProviderName(data.Name.ValueString())
	providers, response, err := request.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read provider, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}
	if len(providers) == 0 {
		resp.Diagnostics.AddError("Client Error", "Provider not found")
		return
	}
	if len(providers) != 1 {
		resp.Diagnostics.AddError("Client Error", "More then one provider was found")
		return
	}

	ConvertProvider(&data, &providers[0])

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ConvertProvider(providerModel *providerDataSourceModel, provider *emmaSdk.Provider) {
	providerModel.Id = types.Int64Value(int64(*provider.Id))
	providerModel.Name = types.StringValue(*provider.Name)
}
