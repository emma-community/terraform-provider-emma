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

var _ datasource.DataSource = &vmsDataSource{}

func NewVmsDataSource() datasource.DataSource {
	return &vmsDataSource{}
}

// vmsDataSource defines the data source implementation.
type vmsDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// vmsDataSourceModel describes the data source data model.
type vmsDataSourceModel struct {
	ProjectId types.Int64 `tfsdk:"project_id"`
	Vms       types.List  `tfsdk:"vms"`
}

// vmItemModel describes individual VM items in the list.
type vmItemModel struct {
	Id               types.Int64   `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Status           types.String  `tfsdk:"status"`
	ProjectId        types.Int64   `tfsdk:"project_id"`
	ProviderId       types.Int64   `tfsdk:"provider_id"`
	ProviderName     types.String  `tfsdk:"provider_name"`
	LocationId       types.Int64   `tfsdk:"location_id"`
	LocationName     types.String  `tfsdk:"location_name"`
	DataCenterId     types.String  `tfsdk:"data_center_id"`
	DataCenterName   types.String  `tfsdk:"data_center_name"`
	OsId             types.Int64   `tfsdk:"os_id"`
	OsType           types.String  `tfsdk:"os_type"`
	VCpu             types.Int64   `tfsdk:"vcpu"`
	VCpuType         types.String  `tfsdk:"vcpu_type"`
	RamGb            types.Int64   `tfsdk:"ram_gb"`
	CloudNetworkType types.String  `tfsdk:"cloud_network_type"`
	CreatedAt        types.String  `tfsdk:"created_at"`
}

func (d *vmsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vms"
}

func (d *vmsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of virtual machines in the Emma platform. Optionally filter by project ID.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.Int64Attribute{
				Description: "Project ID to filter virtual machines",
				Optional:    true,
			},
			"vms": schema.ListNestedAttribute{
				Description: "List of virtual machines",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "ID of the virtual machine",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the virtual machine",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the virtual machine",
							Computed:    true,
						},
						"project_id": schema.Int64Attribute{
							Description: "Project ID the virtual machine belongs to",
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
							Description: "Cloud network type of the virtual machine",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Date and time when the virtual machine was created",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *vmsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vmsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vmsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	apiReq := d.apiClient.VirtualMachinesAPI.GetVms(auth)
	if !data.ProjectId.IsNull() {
		apiReq = apiReq.ProjectId(int32(data.ProjectId.ValueInt64()))
	}
	vms, response, err := apiReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read virtual machines, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	vmList, diags := convertVmsToList(ctx, vms)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Vms = vmList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to get attribute types for VM nested object
func (o vmItemModel) attrTypes() map[string]attr.Type {
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
		"created_at":         types.StringType,
	}
}

// convertVmsToList converts Emma API VMs response to Terraform list
func convertVmsToList(ctx context.Context, vms []emmaSdk.Vm) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	itemModels := make([]vmItemModel, 0, len(vms))

	for _, vm := range vms {
		item := vmItemModel{}

		if vm.Id != nil {
			item.Id = types.Int64Value(int64(*vm.Id))
		} else {
			item.Id = types.Int64Null()
		}

		if vm.Name != nil {
			item.Name = types.StringValue(*vm.Name)
		} else {
			item.Name = types.StringNull()
		}

		if vm.Status != nil {
			item.Status = types.StringValue(*vm.Status)
		} else {
			item.Status = types.StringNull()
		}

		if vm.ProjectId != nil {
			item.ProjectId = types.Int64Value(int64(*vm.ProjectId))
		} else {
			item.ProjectId = types.Int64Null()
		}

		if vm.Provider != nil {
			if vm.Provider.Id != nil {
				item.ProviderId = types.Int64Value(int64(*vm.Provider.Id))
			} else {
				item.ProviderId = types.Int64Null()
			}
			if vm.Provider.Name != nil {
				item.ProviderName = types.StringValue(*vm.Provider.Name)
			} else {
				item.ProviderName = types.StringNull()
			}
		} else {
			item.ProviderId = types.Int64Null()
			item.ProviderName = types.StringNull()
		}

		if vm.Location != nil {
			if vm.Location.Id != nil {
				item.LocationId = types.Int64Value(int64(*vm.Location.Id))
			} else {
				item.LocationId = types.Int64Null()
			}
			if vm.Location.Name != nil {
				item.LocationName = types.StringValue(*vm.Location.Name)
			} else {
				item.LocationName = types.StringNull()
			}
		} else {
			item.LocationId = types.Int64Null()
			item.LocationName = types.StringNull()
		}

		if vm.DataCenter != nil {
			if vm.DataCenter.Id != nil {
				item.DataCenterId = types.StringValue(*vm.DataCenter.Id)
			} else {
				item.DataCenterId = types.StringNull()
			}
			if vm.DataCenter.Name != nil {
				item.DataCenterName = types.StringValue(*vm.DataCenter.Name)
			} else {
				item.DataCenterName = types.StringNull()
			}
		} else {
			item.DataCenterId = types.StringNull()
			item.DataCenterName = types.StringNull()
		}

		if vm.Os != nil {
			if vm.Os.Id != nil {
				item.OsId = types.Int64Value(int64(*vm.Os.Id))
			} else {
				item.OsId = types.Int64Null()
			}
			if vm.Os.Type != nil {
				item.OsType = types.StringValue(*vm.Os.Type)
			} else {
				item.OsType = types.StringNull()
			}
		} else {
			item.OsId = types.Int64Null()
			item.OsType = types.StringNull()
		}

		if vm.VCpu != nil {
			item.VCpu = types.Int64Value(int64(*vm.VCpu))
		} else {
			item.VCpu = types.Int64Null()
		}

		if vm.VCpuType != nil {
			item.VCpuType = types.StringValue(*vm.VCpuType)
		} else {
			item.VCpuType = types.StringNull()
		}

		if vm.RamGb != nil {
			item.RamGb = types.Int64Value(int64(*vm.RamGb))
		} else {
			item.RamGb = types.Int64Null()
		}

		if vm.CloudNetworkType != nil {
			item.CloudNetworkType = types.StringValue(*vm.CloudNetworkType)
		} else {
			item.CloudNetworkType = types.StringNull()
		}

		if vm.CreatedAt != nil {
			item.CreatedAt = types.StringValue(*vm.CreatedAt)
		} else {
			item.CreatedAt = types.StringNull()
		}

		itemModels = append(itemModels, item)
	}

	vmList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmItemModel{}.attrTypes()}, itemModels)
	diags.Append(listDiags...)

	return vmList, diags
}
