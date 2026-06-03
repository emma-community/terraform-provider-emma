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

var _ datasource.DataSource = &systemVolumeConfigsDataSource{}

func NewSystemVolumeConfigsDataSource() datasource.DataSource {
	return &systemVolumeConfigsDataSource{}
}

type systemVolumeConfigsDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

type systemVolumeConfigsDataSourceModel struct {
	AttachedToId   types.Int64  `tfsdk:"attached_to_id"`
	VolumeGbMin    types.Int64  `tfsdk:"volume_gb_min"`
	VolumeType     types.String `tfsdk:"volume_type"`
	Configurations types.List   `tfsdk:"configurations"`
}

type systemVolumeConfigurationModel struct {
	ProviderId     types.Int64   `tfsdk:"provider_id"`
	ProviderName   types.String  `tfsdk:"provider_name"`
	LocationId     types.Int64   `tfsdk:"location_id"`
	LocationName   types.String  `tfsdk:"location_name"`
	DataCenterId   types.String  `tfsdk:"data_center_id"`
	DataCenterName types.String  `tfsdk:"data_center_name"`
	VolumeGb       types.Int64   `tfsdk:"volume_gb"`
	VolumeType     types.String  `tfsdk:"volume_type"`
	PricePerUnit   types.Float64 `tfsdk:"price_per_unit"`
	PriceCurrency  types.String  `tfsdk:"price_currency"`
	PriceUnit      types.String  `tfsdk:"price_unit"`
}

func (d *systemVolumeConfigsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_volume_configs"
}

func (d *systemVolumeConfigsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source retrieves available configurations for system volumes in the Emma platform.\n\n" +
			"Use this data source to discover valid system volume sizes and types when planning VM deployments or system volume upsizing.",
		Attributes: map[string]schema.Attribute{
			"attached_to_id": schema.Int64Attribute{
				Description: "ID of the instance the system volume is attached to",
				Optional:    true,
			},
			"volume_gb_min": schema.Int64Attribute{
				Description: "Minimum volume size in gigabytes to filter configurations",
				Optional:    true,
			},
			"volume_type": schema.StringAttribute{
				Description: "Volume type to filter configurations (e.g., ssd, hdd)",
				Optional:    true,
			},
			"configurations": schema.ListNestedAttribute{
				Description: "List of available system volume configurations",
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
						"location_id": schema.Int64Attribute{
							Description: "Location ID",
							Computed:    true,
						},
						"location_name": schema.StringAttribute{
							Description: "Location name (city or state)",
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
						"volume_gb": schema.Int64Attribute{
							Description: "Volume size in gigabytes",
							Computed:    true,
						},
						"volume_type": schema.StringAttribute{
							Description: "Volume type (e.g., ssd, hdd)",
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

func (d *systemVolumeConfigsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *systemVolumeConfigsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data systemVolumeConfigsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	apiReq := d.apiClient.VolumesConfigurationsAPI.GetSystemVolumeConfigs(auth)

	if !data.AttachedToId.IsNull() && !data.AttachedToId.IsUnknown() {
		apiReq = apiReq.AttachedToId(int32(data.AttachedToId.ValueInt64()))
	}
	if !data.VolumeGbMin.IsNull() && !data.VolumeGbMin.IsUnknown() {
		apiReq = apiReq.VolumeGbMin(int32(data.VolumeGbMin.ValueInt64()))
	}
	if !data.VolumeType.IsNull() && !data.VolumeType.IsUnknown() {
		apiReq = apiReq.VolumeType(data.VolumeType.ValueString())
	}
	apiReq = apiReq.Size(100)
	configs, response, err := apiReq.Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read system volume configurations, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	configList, diags := convertSystemVolumeConfigsToList(ctx, configs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Configurations = configList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (o systemVolumeConfigurationModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"provider_id":      types.Int64Type,
		"provider_name":    types.StringType,
		"location_id":      types.Int64Type,
		"location_name":    types.StringType,
		"data_center_id":   types.StringType,
		"data_center_name": types.StringType,
		"volume_gb":        types.Int64Type,
		"volume_type":      types.StringType,
		"price_per_unit":   types.Float64Type,
		"price_currency":   types.StringType,
		"price_unit":       types.StringType,
	}
}

func convertSystemVolumeConfigsToList(ctx context.Context, configs *emmaSdk.GetSystemVolumeConfigs200Response) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var configModels []systemVolumeConfigurationModel

	if configs == nil || configs.Content == nil {
		emptyList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: systemVolumeConfigurationModel{}.attrTypes()}, []systemVolumeConfigurationModel{})
		diags.Append(listDiags...)
		return emptyList, diags
	}

	for _, config := range configs.Content {
		configModel := systemVolumeConfigurationModel{}

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

		if config.LocationId != nil {
			configModel.LocationId = types.Int64Value(int64(*config.LocationId))
		} else {
			configModel.LocationId = types.Int64Null()
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

	configList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: systemVolumeConfigurationModel{}.attrTypes()}, configModels)
	diags.Append(listDiags...)

	return configList, diags
}
