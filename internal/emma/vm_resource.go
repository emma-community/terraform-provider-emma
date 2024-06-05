package emma

import (
	"context"
	"fmt"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	emma "github.com/emma-community/terraform-provider-emma/internal/emma/validation"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
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
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DataCenterId     types.String `tfsdk:"data_center_id"`
	OsId             types.Int64  `tfsdk:"os_id"`
	CloudNetworkType types.String `tfsdk:"cloud_network_type"`
	VCpuType         types.String `tfsdk:"vcpu_type"`
	VCpu             types.Int64  `tfsdk:"vcpu"`
	RamGb            types.Int64  `tfsdk:"ram_gb"`
	VolumeType       types.String `tfsdk:"volume_type"`
	VolumeGb         types.Int64  `tfsdk:"volume_gb"`
	SshKeyId         types.Int64  `tfsdk:"ssh_key_id"`
	SecurityGroupId  types.Int64  `tfsdk:"security_group_id"`
	Status           types.String `tfsdk:"status"`
	Disks            types.List   `tfsdk:"disks"`
	Networks         types.List   `tfsdk:"networks"`
	Cost             types.Object `tfsdk:"cost"`
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
		Description: "This resource creates a virtual machine according to the specified parameters.\n\n" +
			"To create a virtual machine, follow these steps:\n\n" +
			"1. Select a data center using the `emma_data_center` data source. The data center determines the provider " +
			"and location of the virtual machine.\n\n" +
			"2. Select an available hardware configuration for the virtual machine.\n\n" +
			"3. Select an SSH key for the virtual machine.\n\n" +
			"4. Select an operating system using the `emma_operating_system` data source.\n\n" +
			"5. Choose one of the cloud network types: _multi-cloud_, _isolated,_ or _default_. Choose the _multi-cloud_ " +
			"network type if you need to connect compute instances from different providers.\n\n" +
			"You may choose not to specify a security group. In this case, the virtual machine will be added to the default security group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the virtual machine",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description:   "Name of the virtual machine, virtual machine will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.NotEmptyString{}},
			},
			"data_center_id": schema.StringAttribute{
				Description:   "Data center ID of the virtual machine, virtual machine will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.NotEmptyString{}},
			},
			"os_id": schema.Int64Attribute{
				Description:   "Operating system ID of the virtual machine, virtual machine will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{emma.PositiveInt64{}},
			},
			"cloud_network_type": schema.StringAttribute{
				Description:   "Cloud network type, available values: _multi-cloud_, _isolated,_ or _default_, virtual machine will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.CloudNetworkType{}},
			},
			"vcpu_type": schema.StringAttribute{
				Description: "Type of virtual Central Processing Units (vCPUs), available values: _shared_, _standard_ or _hpc_, virtual machine will be recreated after changing this value",
				Computed:    false,
				Required:    true,
				Optional:    false,
				Validators:  []validator.String{emma.VCpuType{}},
			},
			"vcpu": schema.Int64Attribute{
				Description: "Number of virtual Central Processing Units (vCPUs), the process of edit hardware will start after changing this value",
				Computed:    false,
				Required:    true,
				Optional:    false,
				Validators:  []validator.Int64{emma.PositiveInt64{}},
			},
			"ram_gb": schema.Int64Attribute{
				Description: "Capacity of the RAM in gigabytes, the process of edit hardware will start after changing this value",
				Required:    true,
				Optional:    false,
				Validators:  []validator.Int64{emma.PositiveInt64{}},
			},
			"volume_type": schema.StringAttribute{
				Description:   "Volume type of the compute instance, available values: _ssd_ or _ssd-plus_, the process of edit hardware will start after changing this value",
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.VolumeType{}},
			},
			"volume_gb": schema.Int64Attribute{
				Description: "Volume size in gigabytes, the process of edit hardware will start after changing this value",
				Required:    true,
				Optional:    false,
				Validators:  []validator.Int64{emma.PositiveInt64{}},
			},
			"ssh_key_id": schema.Int64Attribute{
				Description:   "Ssh key ID of the virtual machine, virtual machine will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{emma.PositiveInt64{}},
			},
			"security_group_id": schema.Int64Attribute{
				Description: "Security group ID of the virtual machine, the process of changing the security group will start after changing this value",
				Computed:    false,
				Required:    false,
				Optional:    true,
				Validators:  []validator.Int64{emma.PositiveInt64{}},
			},

			"status": schema.StringAttribute{
				Description: "Status of the virtual machine",
				Computed:    true,
			},
			"disks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Volume ID",
							Computed:    true,
						},
						"size_gb": schema.Int64Attribute{
							Description: "Volume size in gigabytes",
							Computed:    true,
						},
						"type_id": schema.Int64Attribute{
							Description: "ID of the volume type",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Volume type",
							Computed:    true,
						},
						"is_bootable": schema.BoolAttribute{
							Description: "Indicates whether the volume is bootable or not",
							Computed:    true,
						},
					},
				},
			},
			"networks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Network ID",
							Computed:    true,
						},
						"ip": schema.StringAttribute{
							Description: "Network IP",
							Computed:    true,
						},
						"network_type_id": schema.Int64Attribute{
							Description: "ID of the network type",
							Computed:    true,
						},
						"network_type": schema.StringAttribute{
							Description: "Network type",
							Computed:    true,
						},
					},
				},
			},
			"cost": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"unit": schema.StringAttribute{
						Description: "Cost period",
						Computed:    true,
					},
					"currency": schema.StringAttribute{
						Description: "Currency of cost",
						Computed:    true,
					},
					"price": schema.Float64Attribute{
						Description: "Cost of the virtual machine for the period",
						Computed:    true,
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
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData))
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

	tflog.Info(ctx, "Create vm")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	var vmCreateRequest emmaSdk.VmCreate
	ConvertToVmCreateRequest(data, &vmCreateRequest)
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	vm, response, err := r.apiClient.VirtualMachinesAPI.VmCreate(auth).VmCreate(vmCreateRequest).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create virtual machine, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	ConvertVmResponseToResource(ctx, &data, nil, vm, resp.Diagnostics)

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

	tflog.Info(ctx, "Read vm")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	vm, response, err := r.apiClient.VirtualMachinesAPI.GetVm(auth, tools.StringToInt32(data.Id.ValueString())).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read virtual machine, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	ConvertVmResponseToResource(ctx, &data, nil, vm, resp.Diagnostics)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *vmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData vmResourceModel
	var stateData vmResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Update vm")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)

	if !planData.SecurityGroupId.Equal(stateData.SecurityGroupId) {
		if planData.SecurityGroupId.IsUnknown() || planData.SecurityGroupId.IsNull() {
			stateData.SecurityGroupId = types.Int64Null()
		} else {
			vmId := tools.StringToInt32(stateData.Id.ValueString())
			securityGroupInstanceAdd := emmaSdk.SecurityGroupInstanceAdd{InstanceId: &vmId}
			_, response, err := r.apiClient.SecurityGroupsAPI.SecurityGroupInstanceAdd(auth,
				int32(planData.SecurityGroupId.ValueInt64())).SecurityGroupInstanceAdd(securityGroupInstanceAdd).Execute()
			if err != nil {
				resp.Diagnostics.AddError("Client Error",
					fmt.Sprintf("Unable to add virtual machine to security group, got error: %s",
						tools.ExtractErrorMessage(response)))
				return
			}
			stateData.SecurityGroupId = planData.SecurityGroupId
		}
	}

	if !planData.RamGb.Equal(stateData.RamGb) || !planData.VCpu.Equal(stateData.VCpu) ||
		!planData.VolumeGb.Equal(stateData.VolumeGb) || !planData.VCpuType.Equal(stateData.VCpuType) {

		vmActionEditHardwareRequest := emmaSdk.VmActionsRequest{}
		vmEditHardware := emmaSdk.NewVmEditHardware("edithardware", int32(planData.VCpu.ValueInt64()),
			int32(planData.RamGb.ValueInt64()), int32(planData.VolumeGb.ValueInt64()))
		vmEditHardware.VCpuType = planData.VCpuType.ValueStringPointer()
		vmActionEditHardwareRequest.VmEditHardware = vmEditHardware
		vm, response, err := r.apiClient.VirtualMachinesAPI.VmActions(auth,
			tools.StringToInt32(stateData.Id.ValueString())).VmActionsRequest(vmActionEditHardwareRequest).Execute()

		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to edit hardware of the virtual machine, got error: %s",
					tools.ExtractErrorMessage(response)))
			return
		}

		ConvertEditVmHardwareResponseToResource(ctx, &stateData, &planData, vm, resp.Diagnostics)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
}

func (r *vmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data vmResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Delete vm")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	_, response, err := r.apiClient.VirtualMachinesAPI.VmDelete(auth, tools.StringToInt32(data.Id.ValueString())).Execute()

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete virtual machine, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}
}

func (r *vmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "Import vm")

	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	r.Read(ctx, resource.ReadRequest{State: resp.State, Private: resp.Private},
		&resource.ReadResponse{State: resp.State, Private: resp.Private, Diagnostics: resp.Diagnostics})
}

func ConvertToVmCreateRequest(data vmResourceModel, vmCreate *emmaSdk.VmCreate) {
	vmCreate.Name = data.Name.ValueString()
	vmCreate.DataCenterId = data.DataCenterId.ValueString()
	vmCreate.OsId = int32(data.OsId.ValueInt64())
	vmCreate.CloudNetworkType = data.CloudNetworkType.ValueString()
	vmCreate.VCpuType = data.VCpuType.ValueString()
	vmCreate.VCpu = int32(data.VCpu.ValueInt64())
	vmCreate.RamGb = int32(data.RamGb.ValueInt64())
	vmCreate.VolumeType = data.VolumeType.ValueString()
	vmCreate.VolumeGb = int32(data.VolumeGb.ValueInt64())
	if !data.SecurityGroupId.IsUnknown() && !data.SecurityGroupId.IsNull() {
		securityGroupId := int32(data.SecurityGroupId.ValueInt64())
		vmCreate.SecurityGroupId = &securityGroupId
	}
	vmCreate.SshKeyId = int32(data.SshKeyId.ValueInt64())
}

func ConvertEditVmHardwareResponseToResource(ctx context.Context, stateData *vmResourceModel, planData *vmResourceModel, vm *emmaSdk.Vm, diags diag.Diagnostics) {
	stateData.Status = types.StringValue(*vm.Status)

	vmResourceCost := vmResourceCostModel{
		Price:    types.Float64Value(float64(*vm.Cost.Price)),
		Currency: types.StringValue(*vm.Cost.Currency),
		Unit:     types.StringValue(*vm.Cost.Unit),
	}

	costObjectValue, costDiagnostic := types.ObjectValueFrom(ctx, vmResourceCostModel{}.attrTypes(), vmResourceCost)
	stateData.Cost = costObjectValue
	diags.Append(costDiagnostic...)

	var disks []vmResourceDiskModel
	for _, responseDisk := range vm.Disks {
		disk := vmResourceDiskModel{
			Id:         types.Int64Value(int64(*responseDisk.Id)),
			Type_:      types.StringValue(*responseDisk.Type),
			TypeId:     types.Int64Value(int64(*responseDisk.TypeId)),
			SizeGb:     types.Int64Value(int64(*responseDisk.SizeGb)),
			IsBootable: types.BoolValue(*responseDisk.IsBootable),
		}
		disks = append(disks, disk)
	}
	disksListValue, disksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmResourceDiskModel{}.attrTypes()}, disks)
	stateData.Disks = disksListValue
	diags.Append(disksDiagnostic...)

	var networks []vmResourceNetworkModel
	for _, responseNetwork := range vm.Networks {
		network := vmResourceNetworkModel{
			Id:            types.Int64Value(int64(*responseNetwork.Id)),
			Ip:            types.StringPointerValue(responseNetwork.Ip),
			NetworkTypeId: types.Int64Value(int64(*responseNetwork.NetworkTypeId)),
			NetworkType:   types.StringValue(*responseNetwork.NetworkType),
		}
		networks = append(networks, network)
	}
	networksListValue, networksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmResourceNetworkModel{}.attrTypes()}, networks)
	stateData.Networks = networksListValue
	diags.Append(networksDiagnostic...)

	stateData.VCpu = planData.VCpu
	stateData.VCpuType = planData.VCpuType
	stateData.VolumeGb = planData.VolumeGb
	stateData.RamGb = planData.RamGb
}

func ConvertVmResponseToResource(ctx context.Context, stateData *vmResourceModel, planData *vmResourceModel, vm *emmaSdk.Vm, diags diag.Diagnostics) {
	stateData.Id = types.StringValue(strconv.Itoa(int(*vm.Id)))
	stateData.Status = types.StringValue(*vm.Status)
	stateData.Name = types.StringValue(*vm.Name)

	vmResourceCost := vmResourceCostModel{
		Price:    types.Float64Value(float64(*vm.Cost.Price)),
		Currency: types.StringValue(*vm.Cost.Currency),
		Unit:     types.StringValue(*vm.Cost.Unit),
	}

	costObjectValue, costDiagnostic := types.ObjectValueFrom(ctx, vmResourceCostModel{}.attrTypes(), vmResourceCost)
	stateData.Cost = costObjectValue
	diags.Append(costDiagnostic...)

	var disks []vmResourceDiskModel
	for _, responseDisk := range vm.Disks {
		if *responseDisk.IsBootable {
			stateData.VolumeGb = types.Int64Value(int64(*responseDisk.SizeGb))
			stateData.VolumeType = types.StringValue(*responseDisk.Type)
		}
		disk := vmResourceDiskModel{
			Id:         types.Int64Value(int64(*responseDisk.Id)),
			Type_:      types.StringValue(*responseDisk.Type),
			TypeId:     types.Int64Value(int64(*responseDisk.TypeId)),
			SizeGb:     types.Int64Value(int64(*responseDisk.SizeGb)),
			IsBootable: types.BoolValue(*responseDisk.IsBootable),
		}
		disks = append(disks, disk)
	}
	disksListValue, disksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmResourceDiskModel{}.attrTypes()}, disks)
	stateData.Disks = disksListValue
	diags.Append(disksDiagnostic...)

	var networks []vmResourceNetworkModel
	for _, responseNetwork := range vm.Networks {
		network := vmResourceNetworkModel{
			Id:            types.Int64Value(int64(*responseNetwork.Id)),
			Ip:            types.StringPointerValue(responseNetwork.Ip),
			NetworkTypeId: types.Int64Value(int64(*responseNetwork.NetworkTypeId)),
			NetworkType:   types.StringValue(*responseNetwork.NetworkType),
		}
		networks = append(networks, network)
	}
	networksListValue, networksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: vmResourceNetworkModel{}.attrTypes()}, networks)
	stateData.Networks = networksListValue
	diags.Append(networksDiagnostic...)
	stateData.VCpu = types.Int64Value(int64(*vm.VCpu))
	stateData.VCpuType = types.StringValue(*vm.VCpuType)
	if vm.CloudNetworkType != nil {
		stateData.CloudNetworkType = types.StringValue(*vm.CloudNetworkType)
	}
	if (planData != nil && !planData.SecurityGroupId.IsUnknown() && !planData.SecurityGroupId.IsNull()) ||
		(!stateData.SecurityGroupId.IsUnknown() && !stateData.SecurityGroupId.IsNull()) {
		stateData.SecurityGroupId = types.Int64Value(int64(*vm.SecurityGroup.Id))
	}
	stateData.RamGb = types.Int64Value(int64(*vm.RamGb))
	stateData.SshKeyId = types.Int64Value(int64(*vm.SshKeyId))
	stateData.OsId = types.Int64Value(int64(*vm.Os.Id))
	if vm.DataCenter != nil {
		stateData.DataCenterId = types.StringValue(*vm.DataCenter.Id)
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
