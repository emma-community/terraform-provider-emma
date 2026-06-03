package emma

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/async"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/errors"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/retry"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/state"
	emma "github.com/emma-community/terraform-provider-emma/internal/emma/validation"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

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
	Id                types.String  `tfsdk:"id"`
	Name              types.String  `tfsdk:"name"`
	DataCenterId      types.String  `tfsdk:"data_center_id"`
	OsId              types.Int64   `tfsdk:"os_id"`
	CloudNetworkType  types.String  `tfsdk:"cloud_network_type"`
	VCpuType          types.String  `tfsdk:"vcpu_type"`
	VCpu              types.Int64   `tfsdk:"vcpu"`
	RamGb             types.Int64   `tfsdk:"ram_gb"`
	VolumeType        types.String  `tfsdk:"volume_type"`
	VolumeGb          types.Int64   `tfsdk:"volume_gb"`
	SshKeyId          types.Int64   `tfsdk:"ssh_key_id"`
	UserPassword      types.String  `tfsdk:"user_password"`
	SecurityGroupId   types.Int64   `tfsdk:"security_group_id"`
	AcceleratorTypeId types.String  `tfsdk:"accelerator_type_id"`
	Accelerators      types.Float64 `tfsdk:"accelerators"`
	SubnetworkId      types.String  `tfsdk:"subnetwork_id"`
	PrivateIp         types.String  `tfsdk:"private_ip"`
	IpAddressing      types.String  `tfsdk:"ip_addressing"`
	Status            types.String  `tfsdk:"status"`
	Disks             types.List    `tfsdk:"disks"`
	Networks          types.List    `tfsdk:"networks"`
	Cost              types.Object  `tfsdk:"cost"`
}

type VmResourceDiskModel struct {
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
			"3. Select or create an SSH key for the virtual machine using the `emma_ssh_key` resource.\n\n" +
			"4. Select an operating system using the `emma_operating_system` data source.\n\n" +
			"5. Choose one of the cloud network types: multi-cloud, isolated or default. Choose the multi-cloud " +
			"network type if you need to connect compute instances from different providers.\n\n" +
			"6. Select or create an security group for the virtual machine using the `emma_security_group` resource. " +
			"You may choose not to specify a security group. In this case, the virtual machine will be added to the default security group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "ID of the virtual machine",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description:   "Name of the virtual machine, virtual machine will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.NotEmptyString{}, emma.VmName{}},
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
				Description:   "Cloud network type, available values: multi-cloud, isolated or default, virtual machine will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.CloudNetworkType{}},
			},
			"vcpu_type": schema.StringAttribute{
				Description: "Type of virtual Central Processing Units (vCPUs), available values: shared, standard or hpc, virtual machine will be recreated after changing this value",
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
				Description:   "Volume type of the compute instance, available values: ssd or ssd-plus, the process of edit hardware will start after changing this value",
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
				Optional:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{emma.PositiveInt64{}},
			},
			"user_password": schema.StringAttribute{
				Description:   "User password of the virtual machine, virtual machine will be recreated after changing this value",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.UserPassword{}},
			},
			"security_group_id": schema.Int64Attribute{
				Description: "Security group ID of the virtual machine, the process of changing the security group will start after changing this value",
				Computed:    false,
				Required:    false,
				Optional:    true,
				Validators:  []validator.Int64{emma.PositiveInt64{}},
			},
			"accelerator_type_id": schema.StringAttribute{
				Description:   "GPU accelerator type ID for the virtual machine. Use the emma_accelerator_types data source to list available types and pick an id. Must be specified together with accelerators. Changing this value will recreate the virtual machine.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.RequiredTogether{CompanionField: "accelerators"}},
			},
			"accelerators": schema.Float64Attribute{
				Description:   "Number of GPU accelerators. Must be specified together with accelerator_type_id. Changing this value will recreate the virtual machine.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Float64{float64planmodifier.UseStateForUnknown(), float64planmodifier.RequiresReplace()},
				Validators:    []validator.Float64{emma.RequiredTogether{CompanionField: "accelerator_type_id"}},
			},
			"subnetwork_id": schema.StringAttribute{
				Description:   "Subnetwork ID to place the virtual machine in",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"private_ip": schema.StringAttribute{
				Description:   "Private IP address within the subnetwork range",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ip_addressing": schema.StringAttribute{
				Description:   "IP addressing mode, available values: ipv4 or dual-stack",
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
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

	// Build VmCreate request from resource model
	var vmCreateRequest emmaSdk.VmCreate
	ConvertToVmCreateRequest(data, &vmCreateRequest)

	// Call Emma API to create VM
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	vmNew, response, err := r.apiClient.VirtualMachinesAPI.VmCreate(auth).VmCreate(vmCreateRequest).Execute()

	if err != nil {
		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}

		resourceErr := errors.NewError("emma_vm", "Create").
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	ConvertVmResponseToResource(ctx, &data, nil, vmNew, resp.Diagnostics)

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

	// Extract VM ID from state
	vmId := tools.StringToInt32(data.Id.ValueString())

	// Call Emma API to get VM
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	vm, response, err := r.apiClient.VirtualMachinesAPI.GetVm(auth, vmId).Execute()

	if err != nil {
		// Handle 404 errors by removing from state
		if response != nil && response.StatusCode == 404 {
			stateManager := state.NewStateManager(ctx)
			stateManager.RemoveFromState(resp)
			return
		}

		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}

		resourceErr := errors.NewError("emma_vm", "Read").
			WithID(data.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	ConvertVmResponseToResource(ctx, &data, nil, vm, resp.Diagnostics)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func GetVolumesAsList(ctx context.Context, stateData *vmResourceModel, diagnostics diag.Diagnostics) []VmResourceDiskModel {
	var disks []VmResourceDiskModel
	diskDiagnostics := stateData.Disks.ElementsAs(ctx, &disks, false)
	if diskDiagnostics.HasError() {
		diagnostics.Append(diskDiagnostics...)
		return nil
	}
	return disks
}

func GetBootableDisk(ctx context.Context, stateData *vmResourceModel, diagnostics diag.Diagnostics) *VmResourceDiskModel {
	disks := GetVolumesAsList(ctx, stateData, diagnostics)
	if disks == nil {
		return nil
	}

	for _, disk := range disks {
		if disk.IsBootable.ValueBool() {
			return &disk
		}
	}

	return nil
}

func ResizeVolume(ctx context.Context, stateData *vmResourceModel, resp *resource.UpdateResponse, r *vmResource, volumeId int32) {
	bootableDisk := GetBootableDisk(ctx, stateData, resp.Diagnostics)
	if bootableDisk == nil {
		resp.Diagnostics.AddError("Validation Error", "Bootable disk not found")
		return
	}

	tflog.Debug(ctx, "Waiting for volume to reach stable state before resize", map[string]interface{}{
		"volume_id": bootableDisk.Id.ValueInt64(),
	})

	// Wait for volume to reach stable state before resize
	stateManager := state.NewStateTransitionManager(state.StateTransitionConfig{
		ResourceType: "volume",
		ResourceID:   fmt.Sprintf("%d", bootableDisk.Id.ValueInt64()),
		StatusChecker: func(ctx context.Context) (string, error) {
			vol, _, err := r.apiClient.VolumesAPI.GetVolume(ctx, int32(bootableDisk.Id.ValueInt64())).Execute()
			if err != nil {
				return "", err
			}
			if vol.Status == nil {
				return "", fmt.Errorf("volume status is nil")
			}
			return *vol.Status, nil
		},
		TargetStates:       state.VolumeStableStates,
		TransitionalStates: state.VolumeTransitionalStates,
		FailureStates:      state.VolumeFailureStates,
		Timeout:            async.DefaultTimeout,
		PollInterval:       async.DefaultPollInterval,
	})

	if err := stateManager.WaitForStableState(ctx); err != nil {
		resp.Diagnostics.AddError("State Transition Error",
			fmt.Sprintf("Volume did not reach stable state before resize: %s", err.Error()))
		return
	}

	tflog.Info(ctx, "Volume reached stable state, proceeding with resize", map[string]interface{}{
		"volume_id": bootableDisk.Id.ValueInt64(),
	})

	// Perform resize with retry on state conflicts
	retryConfig := retry.StateConflictRetryConfig()
	var lastResponse *http.Response
	var lastAPIError string
	var volume *emmaSdk.Volume

	err := retry.Retry(ctx, retryConfig, func() error {
		volumeEdit := emmaSdk.NewVolumeEdit("edit", volumeId)
		vol, response, err := r.apiClient.VolumesAPI.VolumeActions(ctx, int32(bootableDisk.Id.ValueInt64())).VolumeEdit(*volumeEdit).Execute()

		lastResponse = response
		volume = vol
		if err != nil {
			lastAPIError = tools.ExtractErrorMessage(response)
			statusCode := 0
			if response != nil {
				statusCode = response.StatusCode
			}

			// Check if this is a state conflict error that should be retried
			if retry.IsStateConflictError(err, statusCode, lastAPIError) {
				tflog.Warn(ctx, "Volume resize failed due to state conflict, will retry", map[string]interface{}{
					"volume_id":   bootableDisk.Id.ValueInt64(),
					"status_code": statusCode,
					"error":       lastAPIError,
				})
				return err
			}

			// Non-retryable error
			return fmt.Errorf("non-retryable error: %w", err)
		}

		return nil
	})

	if err != nil {
		statusCode := 0
		apiError := ""
		if lastResponse != nil {
			statusCode = lastResponse.StatusCode
			apiError = lastAPIError
		}

		resourceErr := errors.NewError("emma_vm", "Update").
			WithID(stateData.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(fmt.Sprintf("Unable to resize volume: %s", errors.MapHTTPError(statusCode, apiError))).
			Build()

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	tflog.Info(ctx, "Volume resize completed successfully", map[string]interface{}{
		"volume_id": bootableDisk.Id.ValueInt64(),
	})

	var updatedDisks []VmResourceDiskModel
	disks := GetVolumesAsList(ctx, stateData, resp.Diagnostics)
	if disks != nil {
		for _, disk := range disks {
			if disk.Id.ValueInt64() == int64(*volume.Id) {
				updatedDisk := VmResourceDiskModel{
					Id:         types.Int64Value(int64(*volume.Id)),
					SizeGb:     types.Int64Value(int64(*volume.SizeGb)),
					TypeId:     types.Int64Value(disk.TypeId.ValueInt64()),
					Type_:      types.StringValue(*volume.Type),
					IsBootable: types.BoolValue(*volume.IsSystem),
				}
				updatedDisks = append(updatedDisks, updatedDisk)
			} else {
				updatedDisks = append(updatedDisks, disk)
			}
		}
	}

	stateData.VolumeGb = types.Int64Value(int64(*volume.SizeGb))
	disksListValue, disksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: VmResourceDiskModel{}.attrTypes()}, updatedDisks)
	stateData.Disks = disksListValue
	resp.Diagnostics.Append(disksDiagnostic...)
}

func EditHardware(ctx context.Context, stateData *vmResourceModel, resp *resource.UpdateResponse, r *vmResource, planData *vmResourceModel) {
	vmId := tools.StringToInt32(stateData.Id.ValueString())

	tflog.Debug(ctx, "Waiting for VM to reach stable state before hardware edit", map[string]interface{}{
		"vm_id": stateData.Id.ValueString(),
	})

	// Wait for VM to reach stable state before hardware edit
	stateManager := state.NewStateTransitionManager(state.StateTransitionConfig{
		ResourceType: "vm",
		ResourceID:   stateData.Id.ValueString(),
		StatusChecker: func(ctx context.Context) (string, error) {
			vm, _, err := r.apiClient.VirtualMachinesAPI.GetVm(ctx, vmId).Execute()
			if err != nil {
				return "", err
			}
			if vm.Status == nil {
				return "", fmt.Errorf("VM status is nil")
			}
			return *vm.Status, nil
		},
		TargetStates:       state.VMStableStates,
		TransitionalStates: state.VMTransitionalStates,
		FailureStates:      state.VMFailureStates,
		Timeout:            async.DefaultTimeout,
		PollInterval:       async.DefaultPollInterval,
	})

	if err := stateManager.WaitForStableState(ctx); err != nil {
		resp.Diagnostics.AddError("State Transition Error",
			fmt.Sprintf("VM did not reach stable state before hardware edit: %s", err.Error()))
		return
	}

	tflog.Info(ctx, "VM reached stable state, proceeding with hardware edit", map[string]interface{}{
		"vm_id": stateData.Id.ValueString(),
	})

	// Perform hardware edit with retry on state conflicts
	retryConfig := retry.StateConflictRetryConfig()
	var lastResponse *http.Response
	var lastAPIError string
	var vm *emmaSdk.Vm

	err := retry.Retry(ctx, retryConfig, func() error {
		vmActionEditHardwareRequest := emmaSdk.VmActionsRequest{}

		vmEditHardware := emmaSdk.NewVmEditHardware("edithardware", int32(planData.VCpu.ValueInt64()),
			int32(planData.RamGb.ValueInt64()), int32(planData.VolumeGb.ValueInt64()))
		vmEditHardware.VCpuType = planData.VCpuType.ValueStringPointer()
		vmActionEditHardwareRequest.VmEditHardware = vmEditHardware

		vmResult, response, err := r.apiClient.VirtualMachinesAPI.VmActions(ctx, vmId).VmActionsRequest(vmActionEditHardwareRequest).Execute()

		lastResponse = response
		vm = vmResult
		if err != nil {
			lastAPIError = tools.ExtractErrorMessage(response)
			statusCode := 0
			if response != nil {
				statusCode = response.StatusCode
			}

			// Check if this is a state conflict error that should be retried
			if retry.IsStateConflictError(err, statusCode, lastAPIError) {
				tflog.Warn(ctx, "Hardware edit failed due to state conflict, will retry", map[string]interface{}{
					"vm_id":       stateData.Id.ValueString(),
					"status_code": statusCode,
					"error":       lastAPIError,
				})
				return err
			}

			// Non-retryable error
			return fmt.Errorf("non-retryable error: %w", err)
		}

		return nil
	})

	if err != nil {
		statusCode := 0
		apiError := ""
		if lastResponse != nil {
			statusCode = lastResponse.StatusCode
			apiError = lastAPIError
		}

		resourceErr := errors.NewError("emma_vm", "Update").
			WithID(stateData.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(fmt.Sprintf("Unable to edit hardware: %s", errors.MapHTTPError(statusCode, apiError))).
			Build()

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	tflog.Info(ctx, "Hardware edit completed successfully", map[string]interface{}{
		"vm_id": stateData.Id.ValueString(),
	})

	ConvertEditVmHardwareResponseToResource(ctx, stateData, planData, vm, resp.Diagnostics)
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

	// data_center_id is string like "digitalocean-sgp1", "gcp-europe-west8-a" etc...
	isDigitalOcean := strings.HasPrefix(strings.ToLower(strings.Trim(planData.DataCenterId.String(), "\"")), "digitalocean")
	hardwareChanged := !planData.RamGb.Equal(stateData.RamGb) || !planData.VCpu.Equal(stateData.VCpu) || !planData.VCpuType.Equal(stateData.VCpuType)
	volumeChanged := !planData.VolumeGb.Equal(stateData.VolumeGb)

	if !isDigitalOcean && volumeChanged && hardwareChanged {
		resp.Diagnostics.AddError("Validation Error", "Can't change volume and hardware at the same time")
		return
	}

	if planData.VolumeGb.ValueInt64() < stateData.VolumeGb.ValueInt64() {
		resp.Diagnostics.AddError("Validation Error", "Volume size cannot be decreased")
		return
	}

	tflog.Info(ctx, "Update vm")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)

	// Handle security group changes
	if !planData.SecurityGroupId.Equal(stateData.SecurityGroupId) {
		if planData.SecurityGroupId.IsUnknown() || planData.SecurityGroupId.IsNull() {
			stateData.SecurityGroupId = types.Int64Null()
		} else {
			vmId := tools.StringToInt32(stateData.Id.ValueString())

			tflog.Debug(ctx, "Waiting for VM to reach stable state before security group change", map[string]interface{}{
				"vm_id": stateData.Id.ValueString(),
			})

			// Wait for VM to reach stable state before security group change
			stateManager := state.NewStateTransitionManager(state.StateTransitionConfig{
				ResourceType: "vm",
				ResourceID:   stateData.Id.ValueString(),
				StatusChecker: func(ctx context.Context) (string, error) {
					vm, _, err := r.apiClient.VirtualMachinesAPI.GetVm(ctx, vmId).Execute()
					if err != nil {
						return "", err
					}
					if vm.Status == nil {
						return "", fmt.Errorf("VM status is nil")
					}
					return *vm.Status, nil
				},
				TargetStates:       state.VMStableStates,
				TransitionalStates: state.VMTransitionalStates,
				FailureStates:      state.VMFailureStates,
				Timeout:            async.DefaultTimeout,
				PollInterval:       async.DefaultPollInterval,
			})

			if err := stateManager.WaitForStableState(ctx); err != nil {
				resp.Diagnostics.AddError("State Transition Error",
					fmt.Sprintf("VM did not reach stable state before security group change: %s", err.Error()))
				return
			}

			tflog.Info(ctx, "VM reached stable state, proceeding with security group change", map[string]interface{}{
				"vm_id": stateData.Id.ValueString(),
			})

			// Perform security group change with retry on state conflicts
			retryConfig := retry.StateConflictRetryConfig()
			var lastResponse *http.Response
			var lastAPIError string
			var vm *emmaSdk.Vm

			err := retry.Retry(ctx, retryConfig, func() error {
				securityGroupInstanceAdd := emmaSdk.SecurityGroupInstanceAdd{InstanceId: &vmId}
				vmResult, response, err := r.apiClient.SecurityGroupsAPI.SecurityGroupInstanceAdd(ctx,
					int32(planData.SecurityGroupId.ValueInt64())).SecurityGroupInstanceAdd(securityGroupInstanceAdd).Execute()

				lastResponse = response
				vm = vmResult
				if err != nil {
					lastAPIError = tools.ExtractErrorMessage(response)
					statusCode := 0
					if response != nil {
						statusCode = response.StatusCode
					}

					// Check if this is a state conflict error that should be retried
					if retry.IsStateConflictError(err, statusCode, lastAPIError) {
						tflog.Warn(ctx, "Security group change failed due to state conflict, will retry", map[string]interface{}{
							"vm_id":       stateData.Id.ValueString(),
							"status_code": statusCode,
							"error":       lastAPIError,
						})
						return err
					}

					// Non-retryable error
					return fmt.Errorf("non-retryable error: %w", err)
				}

				return nil
			})

			if err != nil {
				statusCode := 0
				apiError := ""
				if lastResponse != nil {
					statusCode = lastResponse.StatusCode
					apiError = lastAPIError
				}

				resourceErr := errors.NewError("emma_vm", "Update").
					WithID(stateData.Id.ValueString()).
					WithStatusCode(statusCode).
					WithAPIError(apiError).
					WithMessage(fmt.Sprintf("Unable to add VM to security group: %s", errors.MapHTTPError(statusCode, apiError))).
					Build()

				resp.Diagnostics.AddError("Client Error", resourceErr.Error())
				return
			}

			tflog.Info(ctx, "Security group change completed successfully", map[string]interface{}{
				"vm_id": stateData.Id.ValueString(),
			})

			ConvertVmResponseToResource(ctx, &stateData, &planData, vm, resp.Diagnostics)
		}
	}

	if !isDigitalOcean && volumeChanged {
		ResizeVolume(auth, &stateData, resp, r, int32(planData.VolumeGb.ValueInt64()))
	}

	if hardwareChanged || (isDigitalOcean && volumeChanged) {
		EditHardware(auth, &stateData, resp, r, &planData)
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

	vmId := tools.StringToInt32(data.Id.ValueString())
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	_, response, err := r.apiClient.VirtualMachinesAPI.VmDelete(auth, vmId).Execute()

	if err != nil {
		// Handle 404 errors as successful deletion (idempotent)
		if response != nil && response.StatusCode == 404 {
			// VM already deleted, treat as success
			return
		}

		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}

		resourceErr := errors.NewError("emma_vm", "Delete").
			WithID(data.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
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
		vmCreate.SecurityGroupId = tools.ToPointer(int32(data.SecurityGroupId.ValueInt64()))
	}
	if !data.UserPassword.IsUnknown() && !data.UserPassword.IsNull() {
		vmCreate.UserPassword = tools.ToPointer(data.UserPassword.ValueString())
	}
	if !data.SshKeyId.IsUnknown() && !data.SshKeyId.IsNull() {
		vmCreate.SshKeyId = tools.ToInt32PointerOrNil(data.SshKeyId)
	}
	if !data.AcceleratorTypeId.IsUnknown() && !data.AcceleratorTypeId.IsNull() {
		vmCreate.AcceleratorTypeId = tools.ToPointer(data.AcceleratorTypeId.ValueString())
	}
	if !data.Accelerators.IsUnknown() && !data.Accelerators.IsNull() {
		vmCreate.Accelerators = tools.ToPointer(float32(data.Accelerators.ValueFloat64()))
	}
	if !data.SubnetworkId.IsUnknown() && !data.SubnetworkId.IsNull() {
		vmCreate.SubnetworkId = tools.ToPointer(data.SubnetworkId.ValueString())
	}
	if !data.PrivateIp.IsUnknown() && !data.PrivateIp.IsNull() {
		vmCreate.PrivateIp = tools.ToPointer(data.PrivateIp.ValueString())
	}
	if !data.IpAddressing.IsUnknown() && !data.IpAddressing.IsNull() {
		vmCreate.IpAddressing = tools.ToPointer(data.IpAddressing.ValueString())
	}
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

	var disks []VmResourceDiskModel
	for _, responseDisk := range vm.Disks {
		disk := VmResourceDiskModel{
			Id:         types.Int64Value(int64(*responseDisk.Id)),
			Type_:      types.StringValue(*responseDisk.Type),
			TypeId:     types.Int64Value(int64(*responseDisk.TypeId)),
			SizeGb:     types.Int64Value(int64(*responseDisk.SizeGb)),
			IsBootable: types.BoolValue(*responseDisk.IsBootable),
		}
		disks = append(disks, disk)
	}
	disksListValue, disksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: VmResourceDiskModel{}.attrTypes()}, disks)
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

	var disks []VmResourceDiskModel
	for _, responseDisk := range vm.Disks {
		if *responseDisk.IsBootable {
			stateData.VolumeGb = types.Int64Value(int64(*responseDisk.SizeGb))
			stateData.VolumeType = types.StringValue(*responseDisk.Type)
		}
		disk := VmResourceDiskModel{
			Id:         types.Int64Value(int64(*responseDisk.Id)),
			Type_:      types.StringValue(*responseDisk.Type),
			TypeId:     types.Int64Value(int64(*responseDisk.TypeId)),
			SizeGb:     types.Int64Value(int64(*responseDisk.SizeGb)),
			IsBootable: types.BoolValue(*responseDisk.IsBootable),
		}
		disks = append(disks, disk)
	}
	disksListValue, disksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: VmResourceDiskModel{}.attrTypes()}, disks)
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

	if planData != nil && !planData.SecurityGroupId.IsUnknown() && !planData.SecurityGroupId.IsNull() {
		stateData.SecurityGroupId = planData.SecurityGroupId
	} else if !stateData.SecurityGroupId.IsUnknown() && !stateData.SecurityGroupId.IsNull() {
		if vm.SecurityGroup != nil && vm.SecurityGroup.Id != nil {
			stateData.SecurityGroupId = types.Int64Value(int64(*vm.SecurityGroup.Id))
		}
	}

	stateData.RamGb = types.Int64Value(int64(*vm.RamGb))
	stateData.OsId = types.Int64Value(int64(*vm.Os.Id))
	if vm.DataCenter != nil {
		stateData.DataCenterId = types.StringValue(*vm.DataCenter.Id)
	}
	if vm.UserPassword != nil {
		stateData.UserPassword = types.StringValue(*vm.UserPassword)
	}
	if vm.SshKeyId != nil {
		stateData.SshKeyId = types.Int64Value(int64(*vm.SshKeyId))
	}

	// Read GPU accelerator fields from API response
	if vm.Accelerator != nil {
		if vm.Accelerator.AcceleratorTypeId != nil {
			stateData.AcceleratorTypeId = types.StringValue(*vm.Accelerator.AcceleratorTypeId)
		} else {
			stateData.AcceleratorTypeId = types.StringNull()
		}
		if vm.Accelerator.Accelerators != nil {
			stateData.Accelerators = types.Float64Value(float64(*vm.Accelerator.Accelerators))
		} else {
			stateData.Accelerators = types.Float64Null()
		}
	} else if planData != nil {
		stateData.AcceleratorTypeId = planData.AcceleratorTypeId
		stateData.Accelerators = planData.Accelerators
	} else {
		stateData.AcceleratorTypeId = types.StringNull()
		stateData.Accelerators = types.Float64Null()
	}

	// Preserve create-only fields from plan
	if planData != nil {
		stateData.SubnetworkId = planData.SubnetworkId
		stateData.PrivateIp = planData.PrivateIp
		stateData.IpAddressing = planData.IpAddressing
	}
}

func (o vmResourceCostModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"unit":     types.StringType,
		"currency": types.StringType,
		"price":    types.Float64Type,
	}
}

func (o VmResourceDiskModel) attrTypes() map[string]attr.Type {
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

// ConvertVmNewResponseToResource wraps ConvertVmResponseToResource for VmNew type
func ConvertVmNewResponseToResource(ctx context.Context, stateData *vmResourceModel, planData *vmResourceModel, vmNew *emmaSdk.VmNew, diags diag.Diagnostics) {
	// Convert VmNew to Vm by copying fields
	vm := &emmaSdk.Vm{
		Id:               vmNew.Id,
		Name:             vmNew.Name,
		Status:           vmNew.Status,
		VCpu:             vmNew.VCpu,
		VCpuType:         vmNew.VCpuType,
		CloudNetworkType: vmNew.CloudNetworkType,
		RamGb:            vmNew.RamGb,
		SshKeyId:         vmNew.SshKeyId,
		UserPassword:     vmNew.UserPassword,
	}

	// Convert nested objects - only copy fields that exist in both types
	if vmNew.Provider != nil {
		vm.Provider = &emmaSdk.VmProvider{
			Id:   vmNew.Provider.Id,
			Name: vmNew.Provider.Name,
			// Type field doesn't exist in VmProvider in v0.0.10
		}
	}

	if vmNew.Location != nil {
		vm.Location = &emmaSdk.VmLocation{
			Id:   vmNew.Location.Id,
			Name: vmNew.Location.Name,
			// Country field doesn't exist in VmLocation in v0.0.10
		}
	}

	if vmNew.DataCenter != nil {
		vm.DataCenter = &emmaSdk.VmDataCenter{
			Id:   vmNew.DataCenter.Id,
			Name: vmNew.DataCenter.Name,
		}
	}

	if vmNew.Os != nil {
		vm.Os = &emmaSdk.VmOs{
			Id: vmNew.Os.Id,
			// Name field doesn't exist in VmOs in v0.0.10
		}
	}

	if vmNew.SecurityGroup != nil {
		vm.SecurityGroup = &emmaSdk.VmSecurityGroup{
			Id:   vmNew.SecurityGroup.Id,
			Name: vmNew.SecurityGroup.Name,
		}
	}

	if vmNew.Subnetwork != nil {
		vm.Subnetwork = &emmaSdk.VmSubnetwork{
			Id:   vmNew.Subnetwork.Id,
			Name: vmNew.Subnetwork.Name,
		}
	}

	if vmNew.Cost != nil {
		vm.Cost = &emmaSdk.VmCost{
			Currency: vmNew.Cost.Currency,
			Price:    vmNew.Cost.Price,
			Unit:     vmNew.Cost.Unit,
		}
	}

	// Convert disks
	if vmNew.Disks != nil {
		vm.Disks = make([]emmaSdk.VmDisksInner, len(vmNew.Disks))
		for i, disk := range vmNew.Disks {
			vm.Disks[i] = emmaSdk.VmDisksInner{
				Id:         disk.Id,
				SizeGb:     disk.SizeGb,
				TypeId:     disk.TypeId,
				Type:       disk.Type,
				IsBootable: disk.IsBootable,
			}
		}
	}

	// Convert networks
	if vmNew.Networks != nil {
		vm.Networks = make([]emmaSdk.VmNetworksInner, len(vmNew.Networks))
		for i, network := range vmNew.Networks {
			vm.Networks[i] = emmaSdk.VmNetworksInner{
				Id:            network.Id,
				Ip:            network.Ip,
				NetworkTypeId: network.NetworkTypeId,
				NetworkType:   network.NetworkType,
			}
		}
	}

	// Call the existing conversion function
	ConvertVmResponseToResource(ctx, stateData, planData, vm, diags)
}
