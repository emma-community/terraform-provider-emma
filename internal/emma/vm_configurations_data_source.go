package emma

import (
	"context"
	"fmt"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &vmConfigurationsDataSource{}

func NewVmConfigurationsDataSource() datasource.DataSource {
	return &vmConfigurationsDataSource{}
}

// vmConfigurationsDataSource defines the data source implementation.
type vmConfigurationsDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// vmConfigurationsDataSourceModel describes the data source data model.
type vmConfigurationsDataSourceModel struct {
	AcceleratorTypeId types.String  `tfsdk:"accelerator_type_id"`
	DataCenterId      types.String  `tfsdk:"data_center_id"`
	ProviderId        types.Int64   `tfsdk:"provider_id"`
	VCpuType          types.String  `tfsdk:"vcpu_type"`
	VCpuMin           types.Int64   `tfsdk:"vcpu_min"`
	VCpuMax           types.Int64   `tfsdk:"vcpu_max"`
	RamGbMin          types.Int64   `tfsdk:"ram_gb_min"`
	RamGbMax          types.Int64   `tfsdk:"ram_gb_max"`
	VolumeGbMin       types.Int64   `tfsdk:"volume_gb_min"`
	VolumeGbMax       types.Int64   `tfsdk:"volume_gb_max"`
	PriceMin          types.Float64 `tfsdk:"price_min"`
	PriceMax          types.Float64 `tfsdk:"price_max"`
	Configurations    types.List    `tfsdk:"configurations"`
}

// vmConfigurationModel describes individual configuration items in the list.
type vmConfigurationModel struct {
	ProviderId        types.Int64   `tfsdk:"provider_id"`
	ProviderName      types.String  `tfsdk:"provider_name"`
	LocationName      types.String  `tfsdk:"location_name"`
	DataCenterId      types.String  `tfsdk:"data_center_id"`
	DataCenterName    types.String  `tfsdk:"data_center_name"`
	OsId              types.Int64   `tfsdk:"os_id"`
	OsType            types.String  `tfsdk:"os_type"`
	CloudNetworkTypes types.List    `tfsdk:"cloud_network_types"`
	VCpuType          types.String  `tfsdk:"vcpu_type"`
	VCpu              types.Int64   `tfsdk:"vcpu"`
	RamGb             types.Int64   `tfsdk:"ram_gb"`
	VolumeGb          types.Int64   `tfsdk:"volume_gb"`
	VolumeType        types.String  `tfsdk:"volume_type"`
	AcceleratorTypeId types.String  `tfsdk:"accelerator_type_id"`
	AcceleratorType   types.String  `tfsdk:"accelerator_type"`
	Accelerators      types.Float64 `tfsdk:"accelerators"`
	PricePerUnit      types.Float64 `tfsdk:"price_per_unit"`
	PriceCurrency     types.String  `tfsdk:"price_currency"`
	PriceUnit         types.String  `tfsdk:"price_unit"`
}

func (d *vmConfigurationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_configurations"
}

func (d *vmConfigurationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source retrieves available VM configurations in the Emma platform.\n\n" +
			"Use this data source to discover valid VM types, CPU/RAM/storage options, GPU configurations, and pricing information when planning VM deployments.",
		Attributes: map[string]schema.Attribute{
			"accelerator_type_id": schema.StringAttribute{
				Description: "GPU accelerator type ID to filter configurations",
				Optional:    true,
			},
			"data_center_id": schema.StringAttribute{
				Description: "ID of the data center to filter configurations",
				Optional:    true,
			},
			"provider_id": schema.Int64Attribute{
				Description: "ID of the cloud provider to filter configurations",
				Optional:    true,
			},
			"vcpu_type": schema.StringAttribute{
				Description: "vCPU type to filter configurations (shared, standard, hpc)",
				Optional:    true,
			},
			"vcpu_min": schema.Int64Attribute{
				Description: "Minimum number of vCPUs to filter configurations",
				Optional:    true,
			},
			"vcpu_max": schema.Int64Attribute{
				Description: "Maximum number of vCPUs to filter configurations",
				Optional:    true,
			},
			"ram_gb_min": schema.Int64Attribute{
				Description: "Minimum RAM in gigabytes to filter configurations",
				Optional:    true,
			},
			"ram_gb_max": schema.Int64Attribute{
				Description: "Maximum RAM in gigabytes to filter configurations",
				Optional:    true,
			},
			"volume_gb_min": schema.Int64Attribute{
				Description: "Minimum volume size in gigabytes to filter configurations",
				Optional:    true,
			},
			"volume_gb_max": schema.Int64Attribute{
				Description: "Maximum volume size in gigabytes to filter configurations",
				Optional:    true,
			},
			"price_min": schema.Float64Attribute{
				Description: "Minimum price to filter configurations",
				Optional:    true,
			},
			"price_max": schema.Float64Attribute{
				Description: "Maximum price to filter configurations",
				Optional:    true,
			},
			"configurations": schema.ListNestedAttribute{
				Description: "List of available VM configurations",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"provider_id": schema.Int64Attribute{
							Description: "ID of the cloud provider",
							Computed:    true,
						},
						"provider_name": schema.StringAttribute{
							Description: "Name of the cloud provider",
							Computed:    true,
						},
						"location_name": schema.StringAttribute{
							Description: "Location name (city or region)",
							Computed:    true,
						},
						"data_center_id": schema.StringAttribute{
							Description: "ID of the data center",
							Computed:    true,
						},
						"data_center_name": schema.StringAttribute{
							Description: "Name of the data center",
							Computed:    true,
						},
						"os_id": schema.Int64Attribute{
							Description: "ID of the operating system",
							Computed:    true,
						},
						"os_type": schema.StringAttribute{
							Description: "Type of the operating system",
							Computed:    true,
						},
						"cloud_network_types": schema.ListAttribute{
							Description: "List of supported cloud network types",
							Computed:    true,
							ElementType: types.StringType,
						},
						"vcpu_type": schema.StringAttribute{
							Description: "vCPU type (shared, standard, hpc)",
							Computed:    true,
						},
						"vcpu": schema.Int64Attribute{
							Description: "Number of virtual CPUs",
							Computed:    true,
						},
						"ram_gb": schema.Int64Attribute{
							Description: "RAM in gigabytes",
							Computed:    true,
						},
						"volume_gb": schema.Int64Attribute{
							Description: "Volume size in gigabytes",
							Computed:    true,
						},
						"volume_type": schema.StringAttribute{
							Description: "Volume type (e.g., ssd, hdd)",
							Computed:    true,
						},
						"accelerator_type_id": schema.StringAttribute{
							Description: "GPU accelerator type ID",
							Computed:    true,
						},
						"accelerator_type": schema.StringAttribute{
							Description: "GPU accelerator type name",
							Computed:    true,
						},
						"accelerators": schema.Float64Attribute{
							Description: "Number of GPU accelerators",
							Computed:    true,
						},
						"price_per_unit": schema.Float64Attribute{
							Description: "Price per unit for this configuration",
							Computed:    true,
						},
						"price_currency": schema.StringAttribute{
							Description: "Currency of the price",
							Computed:    true,
						},
						"price_unit": schema.StringAttribute{
							Description: "Billing unit for the price",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *vmConfigurationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vmConfigurationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vmConfigurationsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	apiReq := d.apiClient.ComputeInstancesConfigurationsAPI.GetVmConfigs(auth)
	if !data.AcceleratorTypeId.IsNull() {
		apiReq = apiReq.AcceleratorTypeId(data.AcceleratorTypeId.ValueString())
	}
	if !data.DataCenterId.IsNull() {
		apiReq = apiReq.DataCenterId(data.DataCenterId.ValueString())
	}
	if !data.ProviderId.IsNull() {
		apiReq = apiReq.ProviderId(int32(data.ProviderId.ValueInt64()))
	}
	if !data.VCpuType.IsNull() {
		apiReq = apiReq.VCpuType(data.VCpuType.ValueString())
	}
	if !data.VCpuMin.IsNull() {
		apiReq = apiReq.VCpuMin(int32(data.VCpuMin.ValueInt64()))
	}
	if !data.VCpuMax.IsNull() {
		apiReq = apiReq.VCpuMax(int32(data.VCpuMax.ValueInt64()))
	}
	if !data.RamGbMin.IsNull() {
		apiReq = apiReq.RamGbMin(int32(data.RamGbMin.ValueInt64()))
	}
	if !data.RamGbMax.IsNull() {
		apiReq = apiReq.RamGbMax(int32(data.RamGbMax.ValueInt64()))
	}
	if !data.VolumeGbMin.IsNull() {
		apiReq = apiReq.VolumeGbMin(int32(data.VolumeGbMin.ValueInt64()))
	}
	if !data.VolumeGbMax.IsNull() {
		apiReq = apiReq.VolumeGbMax(int32(data.VolumeGbMax.ValueInt64()))
	}
	if !data.PriceMin.IsNull() {
		apiReq = apiReq.PriceMin(float32(data.PriceMin.ValueFloat64()))
	}
	if !data.PriceMax.IsNull() {
		apiReq = apiReq.PriceMax(float32(data.PriceMax.ValueFloat64()))
	}
	apiReq = apiReq.Size(100)
	configs, response, err := apiReq.Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read VM configurations, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	configList, diags := convertVmConfigsToList(ctx, configs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Configurations = configList

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to get attribute types for configuration nested object
func (o vmConfigurationModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"provider_id":         types.Int64Type,
		"provider_name":       types.StringType,
		"location_name":       types.StringType,
		"data_center_id":      types.StringType,
		"data_center_name":    types.StringType,
		"os_id":               types.Int64Type,
		"os_type":             types.StringType,
		"cloud_network_types": types.ListType{ElemType: types.StringType},
		"vcpu_type":           types.StringType,
		"vcpu":                types.Int64Type,
		"ram_gb":              types.Int64Type,
		"volume_gb":           types.Int64Type,
		"volume_type":         types.StringType,
		"accelerator_type_id": types.StringType,
		"accelerator_type":    types.StringType,
		"accelerators":        types.Float64Type,
		"price_per_unit":      types.Float64Type,
		"price_currency":      types.StringType,
		"price_unit":          types.StringType,
	}
}

// convertVmConfigsToList converts Emma API VM configurations response to Terraform list
func convertVmConfigsToList(ctx context.Context, configs *emmaSdk.GetVmConfigs200Response) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var configModels []vmConfigurationModel

	// Check if configs is nil or has no content
	if configs == nil || configs.Content == nil {
		emptyList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmConfigurationModel{}.attrTypes()}, []vmConfigurationModel{})
		diags.Append(listDiags...)
		return emptyList, diags
	}

	for _, config := range configs.Content {
		configModel := vmConfigurationModel{}

		if config.ProviderId != nil {
			configModel.ProviderId = types.Int64Value(int64(*config.ProviderId))
		} else {
			configModel.ProviderId = types.Int64Null()
		}

		if config.ProviderName != nil {
			configModel.ProviderName = types.StringValue(*config.ProviderName)
		} else {
			configModel.ProviderName = types.StringNull()
		}

		if config.LocationName != nil {
			configModel.LocationName = types.StringValue(*config.LocationName)
		} else {
			configModel.LocationName = types.StringNull()
		}

		if config.DataCenterId != nil {
			configModel.DataCenterId = types.StringValue(*config.DataCenterId)
		} else {
			configModel.DataCenterId = types.StringNull()
		}

		if config.DataCenterName != nil {
			configModel.DataCenterName = types.StringValue(*config.DataCenterName)
		} else {
			configModel.DataCenterName = types.StringNull()
		}

		if config.OsId != nil {
			configModel.OsId = types.Int64Value(int64(*config.OsId))
		} else {
			configModel.OsId = types.Int64Null()
		}

		if config.OsType != nil {
			configModel.OsType = types.StringValue(*config.OsType)
		} else {
			configModel.OsType = types.StringNull()
		}

		cloudNetworkTypesList, listDiags := types.ListValueFrom(ctx, types.StringType, config.CloudNetworkTypes)
		diags.Append(listDiags...)
		configModel.CloudNetworkTypes = cloudNetworkTypesList

		if config.VCpuType != nil {
			configModel.VCpuType = types.StringValue(*config.VCpuType)
		} else {
			configModel.VCpuType = types.StringNull()
		}

		if config.VCpu != nil {
			configModel.VCpu = types.Int64Value(int64(*config.VCpu))
		} else {
			configModel.VCpu = types.Int64Null()
		}

		if config.RamGb != nil {
			configModel.RamGb = types.Int64Value(int64(*config.RamGb))
		} else {
			configModel.RamGb = types.Int64Null()
		}

		if config.VolumeGb != nil {
			configModel.VolumeGb = types.Int64Value(int64(*config.VolumeGb))
		} else {
			configModel.VolumeGb = types.Int64Null()
		}

		if config.VolumeType != nil {
			configModel.VolumeType = types.StringValue(*config.VolumeType)
		} else {
			configModel.VolumeType = types.StringNull()
		}

		if config.Accelerator != nil {
			if config.Accelerator.AcceleratorTypeId != nil {
				configModel.AcceleratorTypeId = types.StringValue(*config.Accelerator.AcceleratorTypeId)
			} else {
				configModel.AcceleratorTypeId = types.StringNull()
			}
			if config.Accelerator.AcceleratorType != nil {
				configModel.AcceleratorType = types.StringValue(*config.Accelerator.AcceleratorType)
			} else {
				configModel.AcceleratorType = types.StringNull()
			}
			if config.Accelerator.Accelerators != nil {
				configModel.Accelerators = types.Float64Value(float64(*config.Accelerator.Accelerators))
			} else {
				configModel.Accelerators = types.Float64Null()
			}
		} else {
			configModel.AcceleratorTypeId = types.StringNull()
			configModel.AcceleratorType = types.StringNull()
			configModel.Accelerators = types.Float64Null()
		}

		if config.Cost != nil {
			if config.Cost.PricePerUnit != nil {
				configModel.PricePerUnit = types.Float64Value(float64(*config.Cost.PricePerUnit))
			} else {
				configModel.PricePerUnit = types.Float64Null()
			}
			if config.Cost.Currency != nil {
				configModel.PriceCurrency = types.StringValue(*config.Cost.Currency)
			} else {
				configModel.PriceCurrency = types.StringNull()
			}
			if config.Cost.Unit != nil {
				configModel.PriceUnit = types.StringValue(*config.Cost.Unit)
			} else {
				configModel.PriceUnit = types.StringNull()
			}
		} else {
			configModel.PricePerUnit = types.Float64Null()
			configModel.PriceCurrency = types.StringNull()
			configModel.PriceUnit = types.StringNull()
		}

		configModels = append(configModels, configModel)
	}

	// Convert to Terraform list
	configList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmConfigurationModel{}.attrTypes()}, configModels)
	diags.Append(listDiags...)

	return configList, diags
}
