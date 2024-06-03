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

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &operatingSystemDataSource{}

func NewOperatingSystemDataSource() datasource.DataSource {
	return &operatingSystemDataSource{}
}

// operatingSystemDataSource defines the data source implementation.
type operatingSystemDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// operatingSystemDataSourceModel describes the data source data model.
type operatingSystemDataSourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Family       types.String `tfsdk:"family"`
	Type         types.String `tfsdk:"type"`
	Architecture types.String `tfsdk:"architecture"`
	Version      types.String `tfsdk:"version"`
}

func (d *operatingSystemDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_operating_system"
}

func (d *operatingSystemDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "All compute instances are created with operating system. The operating system ID is necessary for creating any compute instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Operating system id",
				Computed:    true,
			},
			"family": schema.StringAttribute{
				Description: "Operating system family",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Operating system type",
				Computed:    false,
				Required:    true,
				Optional:    false,
			},
			"architecture": schema.StringAttribute{
				Description: "Operating system architecture",
				Computed:    false,
				Required:    true,
				Optional:    false,
			},
			"version": schema.StringAttribute{
				Description: "Operating system version",
				Computed:    false,
				Required:    true,
				Optional:    false,
			},
		},
	}
}

func (d *operatingSystemDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *operatingSystemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data operatingSystemDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Read operating system")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	request := d.apiClient.OperatingSystemsAPI.GetOperatingSystems(auth)
	request = request.Version(data.Version.ValueString())
	request = request.Type_(data.Type.ValueString())
	request = request.Architecture(data.Architecture.ValueString())
	operatingSystems, response, err := request.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read operating system, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}
	if len(operatingSystems) == 0 {
		resp.Diagnostics.AddError("Client Error", "Operating system not found")
		return
	}
	if len(operatingSystems) != 1 {
		resp.Diagnostics.AddError("Client Error", "More then one operating system was found")
		return
	}

	ConvertOperatingSystem(&data, &operatingSystems[0])

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func ConvertOperatingSystem(operatingSystemModel *operatingSystemDataSourceModel, operatingSystem *emmaSdk.OperatingSystem) {
	operatingSystemModel.Id = types.Int64Value(int64(*operatingSystem.Id))
	operatingSystemModel.Family = types.StringValue(*operatingSystem.Family)
	operatingSystemModel.Type = types.StringValue(*operatingSystem.Type)
	operatingSystemModel.Architecture = types.StringValue(*operatingSystem.Architecture)
	operatingSystemModel.Version = types.StringValue(*operatingSystem.Version)
}
