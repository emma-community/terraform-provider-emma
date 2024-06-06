package emma

import (
	"context"
	"fmt"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &locationDataSource{}

func NewLocationDataSource() datasource.DataSource {
	return &locationDataSource{}
}

// locationDataSource defines the data source implementation.
type locationDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// locationDataSourceModel describes the data source data model.
type locationDataSourceModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *locationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

func (d *locationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Locations are cities or states (in the case of the USA) where providers have data centers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "ID of the geographical location",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the geographical location (city or state)",
				Computed:    false,
				Required:    true,
				Optional:    false,
			},
		},
	}
}

func (d *locationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *locationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data locationDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Read location")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	request := d.apiClient.LocationsAPI.GetLocations(auth)
	request = request.Name(data.Name.ValueString())
	locations, response, err := request.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read location, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}
	if len(locations) == 0 {
		resp.Diagnostics.AddError("Client Error", "Location not found")
		return
	}
	if len(locations) != 1 {
		resp.Diagnostics.AddError("Client Error", "More then one location was found")
		return
	}

	ConvertLocation(&data, &locations[0])

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ConvertLocation(locationModel *locationDataSourceModel, location *emmaSdk.Location) {
	locationModel.Id = types.Int64Value(int64(*location.Id))
	locationModel.Name = types.StringValue(*location.Name)
}
