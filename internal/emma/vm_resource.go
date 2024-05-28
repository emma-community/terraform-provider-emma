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
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Vm resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Vm id configurable attribute",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Vm name configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
				Validators:          []validator.String{emma.NotEmptyString{}},
			},
			"data_center_id": schema.StringAttribute{
				MarkdownDescription: "Vm data_center_id configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
				Validators:          []validator.String{emma.NotEmptyString{}},
			},
			"os_id": schema.Int64Attribute{
				MarkdownDescription: "Vm os_id configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
				Validators:          []validator.Int64{emma.PositiveInt64{}},
			},
			"cloud_network_type": schema.StringAttribute{
				MarkdownDescription: "Vm cloud_network_type configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
				Validators:          []validator.String{emma.CloudNetworkType{}},
			},
			"vcpu_type": schema.StringAttribute{
				MarkdownDescription: "Vm vcpu_type configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
				Validators:          []validator.String{emma.VCpuType{}},
			},
			"vcpu": schema.Int64Attribute{
				MarkdownDescription: "Vm vcpu configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
				Validators:          []validator.Int64{emma.PositiveInt64{}},
			},
			"ram_gb": schema.Int64Attribute{
				MarkdownDescription: "Vm ram_gb configurable attribute",
				Required:            true,
				Optional:            false,
				Validators:          []validator.Int64{emma.PositiveInt64{}},
			},
			"volume_type": schema.StringAttribute{
				MarkdownDescription: "Vm volume_type configurable attribute",
				Required:            true,
				Optional:            false,
				Validators:          []validator.String{emma.VolumeType{}},
			},
			"volume_gb": schema.Int64Attribute{
				MarkdownDescription: "Vm volume_gb configurable attribute",
				Required:            true,
				Optional:            false,
				Validators:          []validator.Int64{emma.PositiveInt64{}},
			},
			"ssh_key_id": schema.Int64Attribute{
				MarkdownDescription: "Vm ssh_key_id configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
				Validators:          []validator.Int64{emma.PositiveInt64{}},
			},
			"security_group_id": schema.Int64Attribute{
				MarkdownDescription: "Vm security_group_id configurable attribute",
				Computed:            false,
				Required:            false,
				Optional:            true,
				Validators:          []validator.Int64{emma.PositiveInt64{}},
			},

			"status": schema.StringAttribute{
				MarkdownDescription: "Vm status configurable attribute",
				Computed:            true,
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

	ConvertVmResponseToResource(ctx, &data, vm, resp.Diagnostics)

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

	ConvertVmResponseToResource(ctx, &data, vm, resp.Diagnostics)

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

	if !planData.Name.Equal(stateData.Name) || !planData.OsId.Equal(stateData.OsId) ||
		!planData.DataCenterId.Equal(stateData.DataCenterId) || !planData.VolumeType.Equal(stateData.VolumeType) ||
		!planData.CloudNetworkType.Equal(stateData.CloudNetworkType) || !planData.SshKeyId.Equal(stateData.SshKeyId) {

		var vmCreateRequest emmaSdk.VmCreate
		ConvertToVmCreateRequest(planData, &vmCreateRequest)
		vm, response, err := r.apiClient.VirtualMachinesAPI.VmCreate(auth).VmCreate(vmCreateRequest).Execute()

		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to create virtual machine, got error: %s",
					tools.ExtractErrorMessage(response)))
			return
		}

		_, response, err = r.apiClient.VirtualMachinesAPI.VmDelete(auth, tools.StringToInt32(stateData.Id.ValueString())).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to delete virtual machine, got error: %s",
					tools.ExtractErrorMessage(response)))
			return
		}

		ConvertVmResponseToResource(ctx, &stateData, vm, resp.Diagnostics)

	} else {

		if !planData.SecurityGroupId.Equal(stateData.SecurityGroupId) {
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

func ConvertEditVmHardwareResponseToResource(ctx context.Context, data *vmResourceModel, planData *vmResourceModel, vm *emmaSdk.Vm, diags diag.Diagnostics) {
	data.Status = types.StringValue(*vm.Status)

	vmResourceCost := vmResourceCostModel{
		Price:    types.Float64Value(float64(*vm.Cost.Price)),
		Currency: types.StringValue(*vm.Cost.Currency),
		Unit:     types.StringValue(*vm.Cost.Unit),
	}

	costObjectValue, costDiagnostic := types.ObjectValueFrom(ctx, vmResourceCostModel{}.attrTypes(), vmResourceCost)
	data.Cost = costObjectValue
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
	data.Disks = disksListValue
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
	data.Networks = networksListValue
	diags.Append(networksDiagnostic...)

	data.VCpu = planData.VCpu
	data.VCpuType = planData.VCpuType
	data.VolumeGb = planData.VolumeGb
	data.RamGb = planData.RamGb
}

func ConvertVmResponseToResource(ctx context.Context, data *vmResourceModel, vm *emmaSdk.Vm, diags diag.Diagnostics) {
	data.Id = types.StringValue(strconv.Itoa(int(*vm.Id)))
	data.Status = types.StringValue(*vm.Status)
	data.Name = types.StringValue(*vm.Name)

	vmResourceCost := vmResourceCostModel{
		Price:    types.Float64Value(float64(*vm.Cost.Price)),
		Currency: types.StringValue(*vm.Cost.Currency),
		Unit:     types.StringValue(*vm.Cost.Unit),
	}

	costObjectValue, costDiagnostic := types.ObjectValueFrom(ctx, vmResourceCostModel{}.attrTypes(), vmResourceCost)
	data.Cost = costObjectValue
	diags.Append(costDiagnostic...)

	var disks []vmResourceDiskModel
	for _, responseDisk := range vm.Disks {
		if *responseDisk.IsBootable {
			data.VolumeGb = types.Int64Value(int64(*responseDisk.SizeGb))
			data.VolumeType = types.StringValue(*responseDisk.Type)
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
	data.Disks = disksListValue
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
	data.Networks = networksListValue
	diags.Append(networksDiagnostic...)
	data.VCpu = types.Int64Value(int64(*vm.VCpu))
	data.VCpuType = types.StringValue(*vm.VCpuType)
	if vm.CloudNetworkType != nil {
		data.CloudNetworkType = types.StringValue(*vm.CloudNetworkType)
	}
	if vm.SecurityGroup != nil {
		data.SecurityGroupId = types.Int64Value(int64(*vm.SecurityGroup.Id))
	}
	data.RamGb = types.Int64Value(int64(*vm.RamGb))
	data.SshKeyId = types.Int64Value(int64(*vm.SshKeyId))
	data.OsId = types.Int64Value(int64(*vm.Os.Id))
	if vm.DataCenter != nil {
		data.DataCenterId = types.StringValue(*vm.DataCenter.Id)
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
