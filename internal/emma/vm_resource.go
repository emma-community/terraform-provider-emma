package emma

import (
	"context"
	"fmt"
	emmaSdk "github.com/MandarinSolutions/emma-go-sdk"
	"github.com/MandarinSolutions/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &vmResource{}

func NewVmResource() resource.Resource {
	return &vmResource{}
}

// vmResource defines the resource implementation.
type vmResource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// vmResourceModel describes the resource data model.
type vmResourceModel struct {
	Id               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DataCenterId     types.String `tfsdk:"data_center_id"`
	OsId             types.Int64  `tfsdk:"os_id"`
	CloudNetworkType types.String `tfsdk:"cloud_network_type"`
	VcpuType         types.String `tfsdk:"vcpu_type"`
	Vcpu             types.Int64  `tfsdk:"vcpu"`
	RamGb            types.Int64  `tfsdk:"ram_gb"`
	VolumeType       types.String `tfsdk:"volume_type"`
	VolumeGb         types.Int64  `tfsdk:"volume_gb"`
	SshKeyId         types.Int64  `tfsdk:"ssh_key_id"`
	//SecurityGroupId  types.Int64  `tfsdk:"security_group_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedById    types.Int64  `tfsdk:"created_by_id"`
	CreatedByName  types.String `tfsdk:"created_by_name"`
	ModifiedAt     types.String `tfsdk:"modified_at"`
	ModifiedByName types.String `tfsdk:"modified_by_name"`
	ModifiedById   types.Int64  `tfsdk:"modified_by_id"`
	ProjectId      types.Int64  `tfsdk:"project_id"`
	Status         types.String `tfsdk:"status"`
	Cpu            types.Int64  `tfsdk:"cpu"`
	UserName       types.String `tfsdk:"user_name"`
	Provider_      types.Object `tfsdk:"provider_"`
	Location       types.Object `tfsdk:"location"`
	DataCenter     types.Object `tfsdk:"data_center"`
	Os             types.Object `tfsdk:"os"`
	Disks          types.List   `tfsdk:"disks"`
	Networks       types.List   `tfsdk:"networks"`
	Cost           types.Object `tfsdk:"cost"`
}

type vmResourceProviderModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type vmResourceLocationModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type vmResourceDataCenterModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProviderName types.String `tfsdk:"provider_name"`
	ProviderId   types.Int64  `tfsdk:"provider_id"`
	LocationId   types.Int64  `tfsdk:"location_id"`
	LocationName types.String `tfsdk:"location_name"`
}

type vmResourceOsModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Family       types.String `tfsdk:"family"`
	Type         types.String `tfsdk:"type"`
	Architecture types.String `tfsdk:"architecture"`
	Version      types.String `tfsdk:"version"`
}

type vmResourceDiskModel struct {
	Id         types.Int64  `tfsdk:"id"`
	SizeGb     types.Int64  `tfsdk:"size_gb"`
	TypeId     types.Int64  `tfsdk:"type_id"`
	Type_      types.String `tfsdk:"type"`
	IsBootable types.Bool   `tfsdk:"is_bootable"`
}

type vmResourceNetworkModel struct {
	Id            types.Int64  `tfsdk:"id"`
	Ip            types.String `tfsdk:"ip"`
	NetworkTypeId types.Int64  `tfsdk:"network_type_id"`
	NetworkType   types.String `tfsdk:"network_type"`
}

type vmResourceCostModel struct {
	Unit     types.String  `tfsdk:"unit"`
	Currency types.String  `tfsdk:"currency"`
	Price    types.Float64 `tfsdk:"price"`
}

func (r *vmResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *vmResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Vm resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Vm id configurable attribute",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Vm name configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
			},
			"data_center_id": schema.StringAttribute{
				MarkdownDescription: "Vm data_center_id configurable attribute",
				Required:            true,
				Optional:            false,
			},
			"os_id": schema.Int64Attribute{
				MarkdownDescription: "Vm os_id configurable attribute",
				Required:            true,
				Optional:            false,
			},
			"cloud_network_type": schema.StringAttribute{
				MarkdownDescription: "Vm cloud_network_type configurable attribute",
				Required:            true,
				Optional:            false,
			},
			"vcpu_type": schema.StringAttribute{
				MarkdownDescription: "Vm vcpu_type configurable attribute",
				Required:            true,
				Optional:            false,
			},
			"vcpu": schema.Int64Attribute{
				MarkdownDescription: "Vm vcpu configurable attribute",
				Required:            true,
				Optional:            false,
			},
			"ram_gb": schema.Int64Attribute{
				MarkdownDescription: "Vm ram_gb configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
			},
			"volume_type": schema.StringAttribute{
				MarkdownDescription: "Vm volume_type configurable attribute",
				Required:            true,
				Optional:            false,
			},
			"volume_gb": schema.Int64Attribute{
				MarkdownDescription: "Vm volume_gb configurable attribute",
				Required:            true,
				Optional:            false,
			},
			"ssh_key_id": schema.Int64Attribute{
				MarkdownDescription: "Vm ssh_key_id configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
			},
			//"security_group_id": schema.Int64Attribute{
			//	MarkdownDescription: "Vm security_group_id configurable attribute",
			//	Computed:            true,
			//	Required:            false,
			//	Optional:            true,
			//},

			"created_at": schema.StringAttribute{
				MarkdownDescription: "Vm created_at configurable attribute",
				Computed:            true,
			},
			"created_by_id": schema.Int64Attribute{
				MarkdownDescription: "Vm created_by_id configurable attribute",
				Computed:            true,
			},
			"created_by_name": schema.StringAttribute{
				MarkdownDescription: "Vm created_by_name configurable attribute",
				Computed:            true,
			},
			"modified_at": schema.StringAttribute{
				MarkdownDescription: "Vm modified_at configurable attribute",
				Computed:            true,
			},
			"modified_by_name": schema.StringAttribute{
				MarkdownDescription: "Vm modified_by_name configurable attribute",
				Computed:            true,
			},
			"modified_by_id": schema.Int64Attribute{
				MarkdownDescription: "Vm modified_by_id configurable attribute",
				Computed:            true,
			},
			"project_id": schema.Int64Attribute{
				MarkdownDescription: "Vm project_id configurable attribute",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Vm status configurable attribute",
				Computed:            true,
			},
			"cpu": schema.Int64Attribute{
				MarkdownDescription: "Vm cpu configurable attribute",
				Computed:            true,
			},
			"user_name": schema.StringAttribute{
				MarkdownDescription: "Vm user_name configurable attribute",
				Computed:            true,
			},
			"provider_": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						MarkdownDescription: "Vm provider id configurable attribute",
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Vm provider name configurable attribute",
						Computed:            true,
					},
				},
			},
			"location": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						MarkdownDescription: "Vm  location id configurable attribute",
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Vm location name configurable attribute",
						Computed:            true,
					},
				},
			},
			"data_center": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						MarkdownDescription: "Vm data center id configurable attribute",
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Vm data center name configurable attribute",
						Computed:            true,
					},
					"provider_name": schema.StringAttribute{
						MarkdownDescription: "Vm data center provider_name configurable attribute",
						Computed:            true,
					},
					"provider_id": schema.Int64Attribute{
						MarkdownDescription: "Vm data center provider_id configurable attribute",
						Computed:            true,
					},
					"location_id": schema.Int64Attribute{
						MarkdownDescription: "Vm data center location_id configurable attribute",
						Computed:            true,
					},
					"location_name": schema.StringAttribute{
						MarkdownDescription: "Vm data center location_name configurable attribute",
						Computed:            true,
					},
				},
			},
			"os": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						MarkdownDescription: "Vm os id configurable attribute",
						Computed:            true,
					},
					"family": schema.StringAttribute{
						MarkdownDescription: "Vm os family configurable attribute",
						Computed:            true,
					},
					"type": schema.StringAttribute{
						MarkdownDescription: "Vm os type configurable attribute",
						Computed:            true,
					},
					"architecture": schema.StringAttribute{
						MarkdownDescription: "Vm os architecture configurable attribute",
						Computed:            true,
					},
					"version": schema.StringAttribute{
						MarkdownDescription: "Vm os version configurable attribute",
						Computed:            true,
					},
				},
			},
			"disks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Vm disks id configurable attribute",
							Computed:            true,
						},
						"size_gb": schema.Int64Attribute{
							MarkdownDescription: "Vm disks size_gb configurable attribute",
							Computed:            true,
						},
						"type_id": schema.Int64Attribute{
							MarkdownDescription: "Vm disks type_id configurable attribute",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Vm disks type configurable attribute",
							Computed:            true,
						},
						"is_bootable": schema.BoolAttribute{
							MarkdownDescription: "Vm disks is_bootable configurable attribute",
							Computed:            true,
						},
					},
				},
			},
			"networks": schema.ListNestedAttribute{
				Computed: true,
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Vm networks id configurable attribute",
							Computed:            true,
						},
						"ip": schema.StringAttribute{
							MarkdownDescription: "Vm networks ip configurable attribute",
							Computed:            true,
						},
						"network_type_id": schema.Int64Attribute{
							MarkdownDescription: "Vm networks network_type_id configurable attribute",
							Computed:            true,
						},
						"network_type": schema.StringAttribute{
							MarkdownDescription: "Vm networks network_type configurable attribute",
							Computed:            true,
						},
					},
				},
			},
			"cost": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"unit": schema.StringAttribute{
						MarkdownDescription: "Vm cost unit configurable attribute",
						Computed:            true,
					},
					"currency": schema.StringAttribute{
						MarkdownDescription: "Vm cost currency configurable attribute",
						Computed:            true,
					},
					"price": schema.Float64Attribute{
						MarkdownDescription: "Vm cost price configurable attribute",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (r *vmResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.apiClient = client.apiClient
	r.token = client.token
}

func (r *vmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data vmResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	var vmCreateRequest emmaSdk.VmCreate
	ConvertToVmCreateRequest(data, &vmCreateRequest)
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, r.token.AccessToken)
	vm, _, err := r.apiClient.VirtualMachinesAPI.VmCreate(auth).VmCreate(vmCreateRequest).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create virtual machine, got error: %s", err))
		return
	}

	ConvertResponseToResource(ctx, &data, vm, resp.Diagnostics)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a vm resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data vmResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, r.token.AccessToken)
	vm, _, err := r.apiClient.VirtualMachinesAPI.GetVm(auth, int32(data.Id.ValueInt64())).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read virtual machine, got error: %s", err))
		return
	}

	ConvertResponseToResource(ctx, &data, vm, resp.Diagnostics)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	panic("Not implemented")
}

func (r *vmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vmResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, r.token.AccessToken)
	_, _, err := r.apiClient.VirtualMachinesAPI.VmDelete(auth, int32(data.Id.ValueInt64())).Execute()

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete virtual machine, got error: %s", err))
		return
	}
}

func ConvertToVmCreateRequest(data vmResourceModel, vmCreate *emmaSdk.VmCreate) {
	vmCreate.Name = data.Name.ValueString()
	vmCreate.DataCenterId = data.DataCenterId.ValueString()
	vmCreate.OsId = data.OsId.ValueInt64()
	vmCreate.CloudNetworkType = data.CloudNetworkType.ValueString()
	vmCreate.VCpuType = data.VcpuType.ValueString()
	vmCreate.VCpu = data.Vcpu.ValueInt64()
	vmCreate.RamGb = data.RamGb.ValueInt64()
	vmCreate.VolumeType = data.VolumeType.ValueString()
	vmCreate.VolumeGb = data.VolumeGb.ValueInt64()
	vmCreate.SshKeyId = data.SshKeyId.ValueInt64()
}

func ConvertResponseToResource(ctx context.Context, data *vmResourceModel, vm *emmaSdk.Vm, diags diag.Diagnostics) {
	data.Id = types.Int64Value(tools.ConvertToInt64(vm.Id))
	data.CreatedAt = types.StringValue(tools.ConvertToString(vm.CreatedAt))
	data.CreatedById = types.Int64Value(tools.ConvertToInt64(vm.CreatedById))
	data.CreatedByName = types.StringValue(tools.ConvertToString(vm.CreatedByName))
	data.ModifiedAt = types.StringValue(tools.ConvertToString(vm.ModifiedAt))
	data.ModifiedById = types.Int64Value(tools.ConvertToInt64(vm.ModifiedById))
	data.ModifiedByName = types.StringValue(tools.ConvertToString(vm.ModifiedByName))
	data.ProjectId = types.Int64Value(tools.ConvertToInt64(vm.ProjectId))
	data.Status = types.StringValue(tools.ConvertToString(vm.Status))
	data.UserName = types.StringValue(tools.ConvertToString(vm.UserName))
	data.Cpu = types.Int64Value(tools.ConvertToInt64(vm.Cpu))

	vmResourceDataCenter := vmResourceDataCenterModel{
		Id:           types.Int64Value(tools.ConvertToInt64(vm.DataCenter.Id)),
		Name:         types.StringValue(tools.ConvertToString(vm.DataCenter.Name)),
		ProviderName: types.StringValue(tools.ConvertToString(vm.DataCenter.ProviderName)),
		ProviderId:   types.Int64Value(tools.ConvertToInt64(vm.DataCenter.ProviderId)),
		LocationId:   types.Int64Value(tools.ConvertToInt64(vm.DataCenter.LocationId)),
		LocationName: types.StringValue(tools.ConvertToString(vm.DataCenter.LocationName)),
	}
	dataCenterObjectValue, dataCenterDiagnostic := types.ObjectValueFrom(ctx, vmResourceDataCenterModel{}.attrTypes(), vmResourceDataCenter)
	data.DataCenter = dataCenterObjectValue
	diags.Append(dataCenterDiagnostic...)

	vmResourceOs := vmResourceOsModel{
		Id:           types.Int64Value(tools.ConvertToInt64(vm.Os.Id)),
		Type:         types.StringValue(tools.ConvertToString(vm.Os.Type)),
		Family:       types.StringValue(tools.ConvertToString(vm.Os.Family)),
		Architecture: types.StringValue(tools.ConvertToString(vm.Os.Architecture)),
		Version:      types.StringValue(tools.ConvertToString(vm.Os.Version)),
	}
	osObjectValue, osDiagnostic := types.ObjectValueFrom(ctx, vmResourceOsModel{}.attrTypes(), vmResourceOs)
	data.Os = osObjectValue
	diags.Append(osDiagnostic...)

	vmResourceProvider := vmResourceProviderModel{
		Id:   types.Int64Value(tools.ConvertToInt64(vm.Provider.Id)),
		Name: types.StringValue(tools.ConvertToString(vm.Provider.Name)),
	}
	providerObjectValue, providerDiagnostic := types.ObjectValueFrom(ctx, vmResourceProviderModel{}.attrTypes(), vmResourceProvider)
	data.Provider_ = providerObjectValue
	diags.Append(providerDiagnostic...)

	vmResourceLocation := vmResourceLocationModel{
		Id:   types.Int64Value(tools.ConvertToInt64(vm.Location.Id)),
		Name: types.StringValue(tools.ConvertToString(vm.Location.Name)),
	}

	locationObjectValue, locationDiagnostic := types.ObjectValueFrom(ctx, vmResourceLocationModel{}.attrTypes(), vmResourceLocation)
	data.Location = locationObjectValue
	diags.Append(locationDiagnostic...)

	vmResourceCost := vmResourceCostModel{
		Price:    types.Float64Value(tools.ConvertToFloat64(vm.Cost.Price)),
		Currency: types.StringValue(tools.ConvertToString(vm.Cost.Currency)),
		Unit:     types.StringValue(tools.ConvertToString(vm.Cost.Unit)),
	}

	costObjectValue, costDiagnostic := types.ObjectValueFrom(ctx, vmResourceCostModel{}.attrTypes(), vmResourceCost)
	data.Cost = costObjectValue
	diags.Append(costDiagnostic...)

	var disks []vmResourceDiskModel
	for _, responseDisk := range vm.Disks {
		disk := vmResourceDiskModel{
			Id:         types.Int64Value(tools.ConvertToInt64(responseDisk.Id)),
			Type_:      types.StringValue(tools.ConvertToString(responseDisk.Type)),
			TypeId:     types.Int64Value(tools.ConvertToInt64(responseDisk.TypeId)),
			SizeGb:     types.Int64Value(tools.ConvertToInt64(responseDisk.SizeGb)),
			IsBootable: types.BoolValue(tools.ConvertToBool(responseDisk.IsBootable)),
		}
		disks = append(disks, disk)
	}
	disksListValue, disksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmResourceDiskModel{}.attrTypes()}, disks)
	data.Disks = disksListValue
	diags.Append(disksDiagnostic...)

	var networks []vmResourceNetworkModel
	for _, responseNetwork := range vm.Networks {
		network := vmResourceNetworkModel{
			Id:            types.Int64Value(tools.ConvertToInt64(responseNetwork.Id)),
			Ip:            types.StringValue(tools.ConvertToString(responseNetwork.Ip)),
			NetworkTypeId: types.Int64Value(tools.ConvertToInt64(responseNetwork.NetworkTypeId)),
			NetworkType:   types.StringValue(tools.ConvertToString(responseNetwork.NetworkType)),
		}
		networks = append(networks, network)
	}
	networksListValue, networksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmResourceNetworkModel{}.attrTypes()}, networks)
	data.Networks = networksListValue
	diags.Append(networksDiagnostic...)
}

func (o vmResourceProviderModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.Int64Type,
		"name": types.StringType,
	}
}

func (o vmResourceLocationModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.Int64Type,
		"name": types.StringType,
	}
}

func (o vmResourceDataCenterModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.Int64Type,
		"name":          types.StringType,
		"provider_name": types.StringType,
		"provider_id":   types.Int64Type,
		"location_id":   types.Int64Type,
		"location_name": types.StringType,
	}
}

func (o vmResourceOsModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":           types.Int64Type,
		"family":       types.StringType,
		"type":         types.StringType,
		"architecture": types.StringType,
		"version":      types.StringType,
	}
}

func (o vmResourceCostModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"unit":     types.StringType,
		"currency": types.StringType,
		"price":    types.Float64Type,
	}
}

func (o vmResourceDiskModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.Int64Type,
		"size_gb":     types.Int64Type,
		"type_id":     types.Int64Type,
		"type":        types.StringType,
		"is_bootable": types.BoolType,
	}
}

func (o vmResourceNetworkModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":              types.Int64Type,
		"ip":              types.StringType,
		"network_type_id": types.Int64Type,
		"network_type":    types.StringType,
	}
}
