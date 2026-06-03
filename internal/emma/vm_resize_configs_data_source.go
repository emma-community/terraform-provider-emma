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

var _ datasource.DataSource = &vmResizeConfigsDataSource{}

func NewVmResizeConfigsDataSource() datasource.DataSource {
	return &vmResizeConfigsDataSource{}
}

type vmResizeConfigsDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

type vmResizeConfigsDataSourceModel struct {
	VmId           types.Int64   `tfsdk:"vm_id"`
	VCpuType       types.String  `tfsdk:"vcpu_type"`
	VCpuMin        types.Int64   `tfsdk:"vcpu_min"`
	VCpuMax        types.Int64   `tfsdk:"vcpu_max"`
	RamGbMin       types.Int64   `tfsdk:"ram_gb_min"`
	RamGbMax       types.Int64   `tfsdk:"ram_gb_max"`
	PriceMin       types.Float64 `tfsdk:"price_min"`
	PriceMax       types.Float64 `tfsdk:"price_max"`
	Configurations types.List    `tfsdk:"configurations"`
}

type vmResizeConfigurationModel struct {
	VCpuType            types.String  `tfsdk:"vcpu_type"`
	VCpu                types.Int64   `tfsdk:"vcpu"`
	RamGb               types.Int64   `tfsdk:"ram_gb"`
	ProviderComputeType types.String  `tfsdk:"provider_compute_type"`
	AcceleratorTypeId   types.String  `tfsdk:"accelerator_type_id"`
	AcceleratorType     types.String  `tfsdk:"accelerator_type"`
	Accelerators        types.Float64 `tfsdk:"accelerators"`
	PricePerUnit        types.Float64 `tfsdk:"price_per_unit"`
	PriceCurrency       types.String  `tfsdk:"price_currency"`
	PriceUnit           types.String  `tfsdk:"price_unit"`
}

func (d *vmResizeConfigsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_resize_configs"
}

func (d *vmResizeConfigsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source retrieves available hardware configurations for virtual machine resizing in the Emma platform.\n\n" +
			"Use this data source to discover valid resize configurations for an existing VM before calling the VM edit endpoint.",
		Attributes: map[string]schema.Attribute{
			"vm_id": schema.Int64Attribute{
				Description: "ID of the virtual machine to retrieve resize configurations for",
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
			"price_min": schema.Float64Attribute{
				Description: "Minimum price to filter configurations",
				Optional:    true,
			},
			"price_max": schema.Float64Attribute{
				Description: "Maximum price to filter configurations",
				Optional:    true,
			},
			"configurations": schema.ListNestedAttribute{
				Description: "List of available VM resize configurations",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
						"provider_compute_type": schema.StringAttribute{
							Description: "Provider-specific compute type identifier",
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

func (d *vmResizeConfigsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vmResizeConfigsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vmResizeConfigsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	apiReq := d.apiClient.ComputeInstancesConfigurationsAPI.GetVmResizeConfigs(auth)

	if !data.VmId.IsNull() && !data.VmId.IsUnknown() {
		apiReq = apiReq.VmId(int32(data.VmId.ValueInt64()))
	}
	if !data.VCpuType.IsNull() && !data.VCpuType.IsUnknown() {
		apiReq = apiReq.VCpuType(data.VCpuType.ValueString())
	}
	if !data.VCpuMin.IsNull() && !data.VCpuMin.IsUnknown() {
		apiReq = apiReq.VCpuMin(int32(data.VCpuMin.ValueInt64()))
	}
	if !data.VCpuMax.IsNull() && !data.VCpuMax.IsUnknown() {
		apiReq = apiReq.VCpuMax(int32(data.VCpuMax.ValueInt64()))
	}
	if !data.RamGbMin.IsNull() && !data.RamGbMin.IsUnknown() {
		apiReq = apiReq.RamGbMin(int32(data.RamGbMin.ValueInt64()))
	}
	if !data.RamGbMax.IsNull() && !data.RamGbMax.IsUnknown() {
		apiReq = apiReq.RamGbMax(int32(data.RamGbMax.ValueInt64()))
	}
	if !data.PriceMin.IsNull() && !data.PriceMin.IsUnknown() {
		apiReq = apiReq.PriceMin(float32(data.PriceMin.ValueFloat64()))
	}
	if !data.PriceMax.IsNull() && !data.PriceMax.IsUnknown() {
		apiReq = apiReq.PriceMax(float32(data.PriceMax.ValueFloat64()))
	}
	apiReq = apiReq.Size(100)
	configs, response, err := apiReq.Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read VM resize configurations, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	configList, diags := convertVmResizeConfigsToList(ctx, configs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Configurations = configList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (o vmResizeConfigurationModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"vcpu_type":             types.StringType,
		"vcpu":                  types.Int64Type,
		"ram_gb":                types.Int64Type,
		"provider_compute_type": types.StringType,
		"accelerator_type_id":   types.StringType,
		"accelerator_type":      types.StringType,
		"accelerators":          types.Float64Type,
		"price_per_unit":        types.Float64Type,
		"price_currency":        types.StringType,
		"price_unit":            types.StringType,
	}
}

func convertVmResizeConfigsToList(ctx context.Context, configs *emmaSdk.GetVmResizeConfigs200Response) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var configModels []vmResizeConfigurationModel

	if configs == nil || configs.Content == nil {
		emptyList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmResizeConfigurationModel{}.attrTypes()}, []vmResizeConfigurationModel{})
		diags.Append(listDiags...)
		return emptyList, diags
	}

	for _, config := range configs.Content {
		configModel := vmResizeConfigurationModel{}

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

		if config.ProviderComputeType != nil {
			configModel.ProviderComputeType = types.StringValue(*config.ProviderComputeType)
		} else {
			configModel.ProviderComputeType = types.StringNull()
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

	configList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmResizeConfigurationModel{}.attrTypes()}, configModels)
	diags.Append(listDiags...)

	return configList, diags
}
