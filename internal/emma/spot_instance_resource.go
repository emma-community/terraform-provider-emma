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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
)

var _ resource.Resource = &spotInstanceResource{}

func NewSpotInstanceResource() resource.Resource {
	return &spotInstanceResource{}
}

// spotInstanceResource defines the resource implementation.
type spotInstanceResource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// spotInstanceResourceModel describes the resource data model.
type spotInstanceResourceModel struct {
	Id               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	DataCenterId     types.String  `tfsdk:"data_center_id"`
	OsId             types.Int64   `tfsdk:"os_id"`
	CloudNetworkType types.String  `tfsdk:"cloud_network_type"`
	VCpuType         types.String  `tfsdk:"vcpu_type"`
	VCpu             types.Int64   `tfsdk:"vcpu"`
	RamGb            types.Int64   `tfsdk:"ram_gb"`
	VolumeType       types.String  `tfsdk:"volume_type"`
	VolumeGb         types.Int64   `tfsdk:"volume_gb"`
	SecurityGroupId  types.Int64   `tfsdk:"security_group_id"`
	SshKeyId         types.Int64   `tfsdk:"ssh_key_id"`
	Price            types.Float64 `tfsdk:"price"`
	Status           types.String  `tfsdk:"status"`
	Disks            types.List    `tfsdk:"disks"`
	Networks         types.List    `tfsdk:"networks"`
	Cost             types.Object  `tfsdk:"cost"`
}

type spotInstanceResourceDiskModel struct {
	Id         types.Int64  `tfsdk:"id"`
	SizeGb     types.Int64  `tfsdk:"size_gb"`
	TypeId     types.Int64  `tfsdk:"type_id"`
	Type_      types.String `tfsdk:"type"`
	IsBootable types.Bool   `tfsdk:"is_bootable"`
}

type spotInstanceResourceNetworkModel struct {
	Id            types.Int64  `tfsdk:"id"`
	Ip            types.String `tfsdk:"ip"`
	NetworkTypeId types.Int64  `tfsdk:"network_type_id"`
	NetworkType   types.String `tfsdk:"network_type"`
}

type spotInstanceResourceCostModel struct {
	Unit     types.String  `tfsdk:"unit"`
	Currency types.String  `tfsdk:"currency"`
	Price    types.Float64 `tfsdk:"price"`
}

func (r *spotInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spot_instance"
}

func (r *spotInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "This resource creates a spot instance according to the specified parameters.\n\n" +
			"A Spot Instance is a specialized compute instance that allows you to access and utilize unused instance " +
			"capacity at a steeply discounted rate. Spot price is charged on an hourly basis.\n\n" +
			"To create a spot instance, follow these steps:\n\n" +
			"1. Select a data center using the `emma_data_center` data source. The data center determines the provider " +
			"and location of the spot instance.\n\n" +
			"2. Select an available hardware configuration for the spot instance.\n\n" +
			"3. Select or create an SSH key for the spot instance using the `emma_ssh_key` resource.\n\n" +
			"4. Select an operating system using the `emma_operating_system` data source.\n\n" +
			"5. Choose one of the cloud network types: _multi-cloud, isolated,_ or _default_. Choose the _multi-cloud_ " +
			"network type if you need to connect compute instances from different providers.\n\n" +
			"6. Select or create an security group for the spot instance using the `emma_security_group` resource. " +
			"You may choose not to specify a security group. In this case, the spot instance will be added to the default security group.\n\n" +
			"A `price` field of a spot instance is not required.\n\n" +
			"The spot instance market operates on a bidding system. Your specified price acts as your bid in this market. " +
			"If your bid is higher than the current spot price, your instance request will likely be fulfilled. " +
			"However, if the market price exceeds your bid, your instance may not be launched or could be terminated if already running.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of the spot instance",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description:   "Name of the spot instance, spot instance will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.NotEmptyString{}, emma.VmName{}},
			},
			"data_center_id": schema.StringAttribute{
				Description:   "Data center ID of the spot instance, spot instance will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.NotEmptyString{}},
			},
			"os_id": schema.Int64Attribute{
				Description:   "Operating system ID of the spot instance, spot instance will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{emma.PositiveInt64{}},
			},
			"cloud_network_type": schema.StringAttribute{
				Description:   "Cloud network type, available values: _multi-cloud_, _isolated,_ or _default_, spot instance will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.CloudNetworkType{}},
			},
			"vcpu_type": schema.StringAttribute{
				Description:   "Type of virtual Central Processing Units (vCPUs), available values: _shared_, _standard_ or _hpc_, spot instance will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.VCpuType{}},
			},
			"vcpu": schema.Int64Attribute{
				Description:   "Number of virtual Central Processing Units (vCPUs), spot instance will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{emma.PositiveInt64{}},
			},
			"ram_gb": schema.Int64Attribute{
				Description:   "Capacity of the RAM in gigabytes, spot instance will be recreated after changing this value",
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{emma.PositiveInt64{}},
			},
			"volume_type": schema.StringAttribute{
				Description:   "Volume type of the compute instance, available values: _ssd_ or _ssd-plus_, spot instance will be recreated after changing this value",
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{emma.VolumeType{}},
			},
			"volume_gb": schema.Int64Attribute{
				Description:   "Volume size in gigabytes, spot instance will be recreated after changing this value",
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{emma.PositiveInt64{}},
			},
			"ssh_key_id": schema.Int64Attribute{
				Description:   "Ssh key ID of the spot instance, spot instance will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
				Validators:    []validator.Int64{emma.PositiveInt64{}},
			},
			"security_group_id": schema.Int64Attribute{
				Description: "Security group ID of the spot instance, the process of changing the security group will start after changing this value",
				Computed:    false,
				Required:    false,
				Optional:    true,
				Validators:  []validator.Int64{emma.PositiveInt64{}},
			},
			"price": schema.Float64Attribute{
				Description:   "Offer price of the spot instance, spot instance will be recreated after changing this value",
				Computed:      false,
				Required:      true,
				Optional:      false,
				PlanModifiers: []planmodifier.Float64{float64planmodifier.RequiresReplace()},
				Validators:    []validator.Float64{emma.PositiveFloat64{}},
			},

			"status": schema.StringAttribute{
				Description: "Status of the spot instance",
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
						Description: "Cost of the spot instance for the period",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (r *spotInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *spotInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data spotInstanceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Create spot instance")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	var spotInstanceCreateRequest emmaSdk.SpotCreate
	ConvertToSpotInstanceCreateRequest(data, &spotInstanceCreateRequest)
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	spotInstance, response, err := r.apiClient.SpotInstancesAPI.SpotCreate(auth).SpotCreate(spotInstanceCreateRequest).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create spot machine, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	ConvertSpotInstanceResponseToResource(ctx, &data, nil, spotInstance, resp.Diagnostics)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *spotInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data spotInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Read spot instance")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	spotInstance, response, err := r.apiClient.SpotInstancesAPI.GetSpot(auth, tools.StringToInt32(data.Id.ValueString())).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read spot machine, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	ConvertSpotInstanceResponseToResource(ctx, &data, nil, spotInstance, resp.Diagnostics)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *spotInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData spotInstanceResourceModel
	var stateData spotInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Update spot instance")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)

	if !planData.SecurityGroupId.Equal(stateData.SecurityGroupId) {
		if planData.SecurityGroupId.IsUnknown() || planData.SecurityGroupId.IsNull() {
			stateData.SecurityGroupId = types.Int64Null()
		} else {
			vmId := tools.StringToInt32(stateData.Id.ValueString())
			securityGroupInstanceAdd := emmaSdk.SecurityGroupInstanceAdd{InstanceId: &vmId}
			vm, response, err := r.apiClient.SecurityGroupsAPI.SecurityGroupInstanceAdd(auth,
				int32(planData.SecurityGroupId.ValueInt64())).SecurityGroupInstanceAdd(securityGroupInstanceAdd).Execute()
			if err != nil {
				resp.Diagnostics.AddError("Client Error",
					fmt.Sprintf("Unable to add spot instance to security group, got error: %s",
						tools.ExtractErrorMessage(response)))
				return
			}
			stateData.SecurityGroupId = planData.SecurityGroupId
			ConvertSpotInstanceResponseToResource(ctx, &stateData, &planData, vm, resp.Diagnostics)
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
}

func (r *spotInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data spotInstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Delete spot instance")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	_, response, err := r.apiClient.SpotInstancesAPI.SpotDelete(auth, tools.StringToInt32(data.Id.ValueString())).Execute()

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete spot machine, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}
}

func (r *spotInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "Import spot instance")

	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	r.Read(ctx, resource.ReadRequest{State: resp.State, Private: resp.Private},
		&resource.ReadResponse{State: resp.State, Private: resp.Private, Diagnostics: resp.Diagnostics})
}

func ConvertToSpotInstanceCreateRequest(data spotInstanceResourceModel, spotInstanceCreate *emmaSdk.SpotCreate) {
	spotInstanceCreate.Name = data.Name.ValueString()
	spotInstanceCreate.DataCenterId = data.DataCenterId.ValueString()
	spotInstanceCreate.OsId = int32(data.OsId.ValueInt64())
	spotInstanceCreate.CloudNetworkType = data.CloudNetworkType.ValueString()
	spotInstanceCreate.VCpuType = data.VCpuType.ValueString()
	spotInstanceCreate.VCpu = int32(data.VCpu.ValueInt64())
	spotInstanceCreate.RamGb = int32(data.RamGb.ValueInt64())
	spotInstanceCreate.VolumeType = data.VolumeType.ValueString()
	spotInstanceCreate.VolumeGb = int32(data.VolumeGb.ValueInt64())
	if !data.SecurityGroupId.IsUnknown() && !data.SecurityGroupId.IsNull() {
		securityGroupId := int32(data.SecurityGroupId.ValueInt64())
		spotInstanceCreate.SecurityGroupId = &securityGroupId
	}
	spotInstanceCreate.SshKeyId = int32(data.SshKeyId.ValueInt64())
	spotInstanceCreate.Price = float32(data.Price.ValueFloat64())
}

func ConvertSpotInstanceResponseToResource(ctx context.Context, stateData *spotInstanceResourceModel, planData *spotInstanceResourceModel, spotInstance *emmaSdk.Vm, diags diag.Diagnostics) {
	stateData.Id = types.StringValue(strconv.Itoa(int(*spotInstance.Id)))
	stateData.Status = types.StringValue(*spotInstance.Status)
	stateData.Name = types.StringValue(*spotInstance.Name)

	if planData != nil && !planData.Price.IsUnknown() && !planData.Price.IsNull() {
		stateData.Price = planData.Price
	}

	vmResourceCost := vmResourceCostModel{
		Price:    types.Float64Value(float64(*spotInstance.Cost.Price)),
		Currency: types.StringValue(*spotInstance.Cost.Currency),
		Unit:     types.StringValue(*spotInstance.Cost.Unit),
	}

	costObjectValue, costDiagnostic := types.ObjectValueFrom(ctx, spotInstanceResourceCostModel{}.attrTypes(), vmResourceCost)
	stateData.Cost = costObjectValue
	diags.Append(costDiagnostic...)

	var disks []vmResourceDiskModel
	for _, responseDisk := range spotInstance.Disks {
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
	disksListValue, disksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: spotInstanceResourceDiskModel{}.attrTypes()}, disks)
	stateData.Disks = disksListValue
	diags.Append(disksDiagnostic...)

	var networks []vmResourceNetworkModel
	for _, responseNetwork := range spotInstance.Networks {
		network := vmResourceNetworkModel{
			Id:            types.Int64Value(int64(*responseNetwork.Id)),
			Ip:            types.StringPointerValue(responseNetwork.Ip),
			NetworkTypeId: types.Int64Value(int64(*responseNetwork.NetworkTypeId)),
			NetworkType:   types.StringValue(*responseNetwork.NetworkType),
		}
		networks = append(networks, network)
	}
	networksListValue, networksDiagnostic := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: spotInstanceResourceNetworkModel{}.attrTypes()}, networks)
	stateData.Networks = networksListValue
	diags.Append(networksDiagnostic...)
	stateData.VCpu = types.Int64Value(int64(*spotInstance.VCpu))
	stateData.VCpuType = types.StringValue(*spotInstance.VCpuType)
	if spotInstance.CloudNetworkType != nil {
		stateData.CloudNetworkType = types.StringValue(*spotInstance.CloudNetworkType)
	}
	if (planData != nil && !planData.SecurityGroupId.IsUnknown() && !planData.SecurityGroupId.IsNull()) ||
		(!stateData.SecurityGroupId.IsUnknown() && !stateData.SecurityGroupId.IsNull()) {
		stateData.SecurityGroupId = types.Int64Value(int64(*spotInstance.SecurityGroup.Id))
	}
	stateData.RamGb = types.Int64Value(int64(*spotInstance.RamGb))
	stateData.SshKeyId = types.Int64Value(int64(*spotInstance.SshKeyId))
	stateData.OsId = types.Int64Value(int64(*spotInstance.Os.Id))
	if spotInstance.DataCenter != nil {
		stateData.DataCenterId = types.StringValue(*spotInstance.DataCenter.Id)
	}
}

func (o spotInstanceResourceCostModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"unit":     types.StringType,
		"currency": types.StringType,
		"price":    types.Float64Type,
	}
}

func (o spotInstanceResourceDiskModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.Int64Type,
		"size_gb":     types.Int64Type,
		"type_id":     types.Int64Type,
		"type":        types.StringType,
		"is_bootable": types.BoolType,
	}
}

func (o spotInstanceResourceNetworkModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":              types.Int64Type,
		"ip":              types.StringType,
		"network_type_id": types.Int64Type,
		"network_type":    types.StringType,
	}
}
