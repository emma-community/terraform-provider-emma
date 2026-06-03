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

var _ datasource.DataSource = &spotsDataSource{}

func NewSpotsDataSource() datasource.DataSource {
	return &spotsDataSource{}
}

// spotsDataSource defines the data source implementation.
type spotsDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// spotsDataSourceModel describes the data source data model.
type spotsDataSourceModel struct {
	ProjectId types.Int64 `tfsdk:"project_id"`
	Spots     types.List  `tfsdk:"spots"`
}

// spotItemModel describes individual spot instance items in the list.
type spotItemModel struct {
	Id                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Status            types.String `tfsdk:"status"`
	ProjectId         types.Int64  `tfsdk:"project_id"`
	ProviderId        types.Int64  `tfsdk:"provider_id"`
	ProviderName      types.String `tfsdk:"provider_name"`
	LocationId        types.Int64  `tfsdk:"location_id"`
	LocationName      types.String `tfsdk:"location_name"`
	DataCenterId      types.String `tfsdk:"data_center_id"`
	DataCenterName    types.String `tfsdk:"data_center_name"`
	OsId              types.Int64  `tfsdk:"os_id"`
	OsType            types.String `tfsdk:"os_type"`
	VCpu              types.Int64  `tfsdk:"vcpu"`
	VCpuType          types.String `tfsdk:"vcpu_type"`
	RamGb             types.Int64  `tfsdk:"ram_gb"`
	CloudNetworkType  types.String `tfsdk:"cloud_network_type"`
	AcceleratorType   types.String `tfsdk:"accelerator_type"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

func (d *spotsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spots"
}

func (d *spotsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of spot instances in the Emma platform. Optionally filter by project ID.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.Int64Attribute{
				Description: "Project ID to filter spot instances",
				Optional:    true,
			},
			"spots": schema.ListNestedAttribute{
				Description: "List of spot instances",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "ID of the spot instance",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the spot instance",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the spot instance",
							Computed:    true,
						},
						"project_id": schema.Int64Attribute{
							Description: "Project ID the spot instance belongs to",
							Computed:    true,
						},
						"provider_id": schema.Int64Attribute{
							Description: "ID of the cloud provider",
							Computed:    true,
						},
						"provider_name": schema.StringAttribute{
							Description: "Name of the cloud provider",
							Computed:    true,
						},
						"location_id": schema.Int64Attribute{
							Description: "ID of the data center location",
							Computed:    true,
						},
						"location_name": schema.StringAttribute{
							Description: "Name of the data center location",
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
						"vcpu": schema.Int64Attribute{
							Description: "Number of virtual CPUs",
							Computed:    true,
						},
						"vcpu_type": schema.StringAttribute{
							Description: "Type of virtual CPUs (shared, standard, hpc)",
							Computed:    true,
						},
						"ram_gb": schema.Int64Attribute{
							Description: "RAM in gigabytes",
							Computed:    true,
						},
						"cloud_network_type": schema.StringAttribute{
							Description: "Cloud network type of the spot instance",
							Computed:    true,
						},
						"accelerator_type": schema.StringAttribute{
							Description: "GPU accelerator type name, if applicable",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Date and time when the spot instance was created",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *spotsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *spotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data spotsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	apiReq := d.apiClient.SpotInstancesAPI.GetSpots(auth)
	if !data.ProjectId.IsNull() {
		apiReq = apiReq.ProjectId(int32(data.ProjectId.ValueInt64()))
	}
	spots, response, err := apiReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read spot instances, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	spotList, diags := convertSpotsToList(ctx, spots)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Spots = spotList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to get attribute types for spot instance nested object
func (o spotItemModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                 types.Int64Type,
		"name":               types.StringType,
		"status":             types.StringType,
		"project_id":         types.Int64Type,
		"provider_id":        types.Int64Type,
		"provider_name":      types.StringType,
		"location_id":        types.Int64Type,
		"location_name":      types.StringType,
		"data_center_id":     types.StringType,
		"data_center_name":   types.StringType,
		"os_id":              types.Int64Type,
		"os_type":            types.StringType,
		"vcpu":               types.Int64Type,
		"vcpu_type":          types.StringType,
		"ram_gb":             types.Int64Type,
		"cloud_network_type": types.StringType,
		"accelerator_type":   types.StringType,
		"created_at":         types.StringType,
	}
}

// convertSpotsToList converts Emma API spot instances response to Terraform list
func convertSpotsToList(ctx context.Context, spots []emmaSdk.SpotVm) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var itemModels []spotItemModel

	for _, spot := range spots {
		item := spotItemModel{}

		if spot.Id != nil {
			item.Id = types.Int64Value(int64(*spot.Id))
		} else {
			item.Id = types.Int64Null()
		}

		if spot.Name != nil {
			item.Name = types.StringValue(*spot.Name)
		} else {
			item.Name = types.StringNull()
		}

		if spot.Status != nil {
			item.Status = types.StringValue(*spot.Status)
		} else {
			item.Status = types.StringNull()
		}

		if spot.ProjectId != nil {
			item.ProjectId = types.Int64Value(int64(*spot.ProjectId))
		} else {
			item.ProjectId = types.Int64Null()
		}

		if spot.Provider != nil {
			if spot.Provider.Id != nil {
				item.ProviderId = types.Int64Value(int64(*spot.Provider.Id))
			} else {
				item.ProviderId = types.Int64Null()
			}
			if spot.Provider.Name != nil {
				item.ProviderName = types.StringValue(*spot.Provider.Name)
			} else {
				item.ProviderName = types.StringNull()
			}
		} else {
			item.ProviderId = types.Int64Null()
			item.ProviderName = types.StringNull()
		}

		if spot.Location != nil {
			if spot.Location.Id != nil {
				item.LocationId = types.Int64Value(int64(*spot.Location.Id))
			} else {
				item.LocationId = types.Int64Null()
			}
			if spot.Location.Name != nil {
				item.LocationName = types.StringValue(*spot.Location.Name)
			} else {
				item.LocationName = types.StringNull()
			}
		} else {
			item.LocationId = types.Int64Null()
			item.LocationName = types.StringNull()
		}

		if spot.DataCenter != nil {
			if spot.DataCenter.Id != nil {
				item.DataCenterId = types.StringValue(*spot.DataCenter.Id)
			} else {
				item.DataCenterId = types.StringNull()
			}
			if spot.DataCenter.Name != nil {
				item.DataCenterName = types.StringValue(*spot.DataCenter.Name)
			} else {
				item.DataCenterName = types.StringNull()
			}
		} else {
			item.DataCenterId = types.StringNull()
			item.DataCenterName = types.StringNull()
		}

		if spot.Os != nil {
			if spot.Os.Id != nil {
				item.OsId = types.Int64Value(int64(*spot.Os.Id))
			} else {
				item.OsId = types.Int64Null()
			}
			if spot.Os.Type != nil {
				item.OsType = types.StringValue(*spot.Os.Type)
			} else {
				item.OsType = types.StringNull()
			}
		} else {
			item.OsId = types.Int64Null()
			item.OsType = types.StringNull()
		}

		if spot.VCpu != nil {
			item.VCpu = types.Int64Value(int64(*spot.VCpu))
		} else {
			item.VCpu = types.Int64Null()
		}

		if spot.VCpuType != nil {
			item.VCpuType = types.StringValue(*spot.VCpuType)
		} else {
			item.VCpuType = types.StringNull()
		}

		if spot.RamGb != nil {
			item.RamGb = types.Int64Value(int64(*spot.RamGb))
		} else {
			item.RamGb = types.Int64Null()
		}

		if spot.CloudNetworkType != nil {
			item.CloudNetworkType = types.StringValue(*spot.CloudNetworkType)
		} else {
			item.CloudNetworkType = types.StringNull()
		}

		if spot.Accelerator != nil && spot.Accelerator.AcceleratorType != nil {
			item.AcceleratorType = types.StringValue(*spot.Accelerator.AcceleratorType)
		} else {
			item.AcceleratorType = types.StringNull()
		}

		if spot.CreatedAt != nil {
			item.CreatedAt = types.StringValue(*spot.CreatedAt)
		} else {
			item.CreatedAt = types.StringNull()
		}

		itemModels = append(itemModels, item)
	}

	spotList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: spotItemModel{}.attrTypes()}, itemModels)
	diags.Append(listDiags...)

	return spotList, diags
}
