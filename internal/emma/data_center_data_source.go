package emma

import (
	"context"
	"fmt"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &dataCenterDataSource{}

func NewDataCenterDataSource() datasource.DataSource {
	return &dataCenterDataSource{}
}

// dataCenterDataSource defines the data source implementation.
type dataCenterDataSource struct {
	apiClient  *emmaSdk.APIClient
	token      *emmaSdk.Token
	LocationID *int64
}

// dataCenterDataSourceModel describes the data source data model.
type dataCenterDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProviderName types.String `tfsdk:"provider_name"`
	ProviderId   types.Int64  `tfsdk:"provider_id"`
	LocationId   types.Int64  `tfsdk:"location_id"`
	LocationName types.String `tfsdk:"location_name"`
}

func (d *dataCenterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_center"
}

func (d *dataCenterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Data center data sources",
		Attributes: map[string]schema.Attribute{

			"id": schema.StringAttribute{
				MarkdownDescription: "Data center id",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Data center name",
				Computed:            true,
				Required:            false,
				Optional:            true,
			},
			"provider_name": schema.StringAttribute{
				MarkdownDescription: "Data center provider_name",
				Computed:            true,
				Required:            false,
				Optional:            true,
			},
			"provider_id": schema.Int64Attribute{
				MarkdownDescription: "Data center provider_id",
				Computed:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "Data center location_id",
				Computed:            true,
				Required:            false,
				Optional:            true,
			},
			"location_name": schema.StringAttribute{
				MarkdownDescription: "Data center location_name",
				Computed:            true,
			},
		},
	}
}

func (d *dataCenterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataCenterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataCenterDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Read data center")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	request := d.apiClient.DataCentersAPI.GetDataCenters(auth)
	if !data.LocationId.IsUnknown() && !data.LocationId.IsNull() {
		request = request.LocationId(int32(data.LocationId.ValueInt64()))
	}
	if !data.ProviderName.IsUnknown() && !data.ProviderName.IsNull() {
		request = request.ProviderName(data.ProviderName.ValueString())
	}
	if !data.Name.IsUnknown() && !data.Name.IsNull() {
		request = request.DataCenterName(data.Name.ValueString())
	}
	dataCenters, response, err := request.Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read data center, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	if len(dataCenters) == 0 {
		resp.Diagnostics.AddError("Client Error", "Data center not found")
		return
	}
	if len(dataCenters) != 1 {
		resp.Diagnostics.AddError("Client Error", "More then one data center was found")
		return
	}

	ConvertDataCenter(&data, &dataCenters[0])

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ConvertDataCenter(dataCenterModel *dataCenterDataSourceModel, dataCenter *emmaSdk.DataCenter) {
	dataCenterModel.Id = types.StringValue(*dataCenter.Id)
	dataCenterModel.Name = types.StringValue(*dataCenter.Name)
	dataCenterModel.ProviderName = types.StringValue(*dataCenter.ProviderName)
	dataCenterModel.ProviderId = types.Int64Value(int64(*dataCenter.ProviderId))
	dataCenterModel.LocationId = types.Int64Value(int64(*dataCenter.LocationId))
	dataCenterModel.LocationName = types.StringValue(*dataCenter.LocationName)
}
