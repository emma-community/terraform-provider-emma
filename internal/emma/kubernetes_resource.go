package emma

import (
	"context"
	"fmt"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	emma "github.com/emma-community/terraform-provider-emma/internal/emma/validation"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"time"
)

var _ resource.Resource = &kubernetesResource{}

func NewKubernetesResource() resource.Resource {
	return &kubernetesResource{}
}

type kubernetesResource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

type kubernetesModel struct {
	Id                 types.Int64                 `tfsdk:"id"`
	Name               types.String                `tfsdk:"name"`
	DeploymentLocation types.String                `tfsdk:"deployment_location"`
	DomainName         types.String                `tfsdk:"domain_name"`
	WorkerNodes        []kubernetesWorkerNodeModel `tfsdk:"worker_nodes"`
	AutoscalingConfigs *[]autoscalingConfigModel   `tfsdk:"autoscaling_configs"`
}

type kubernetesWorkerNodeModel struct {
	Id            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	GeneratedName types.String `tfsdk:"generated_name"`
	DataCenterID  types.String `tfsdk:"data_center_id"`
	VCpuType      types.String `tfsdk:"vcpu_type"`
	VCpu          types.Int64  `tfsdk:"vcpu"`
	RamGb         types.Int64  `tfsdk:"ram_gb"`
	VolumeType    types.String `tfsdk:"volume_type"`
	VolumeGb      types.Int64  `tfsdk:"volume_gb"`
}

type autoscalingConfigModel struct {
	GroupName                          types.String                 `tfsdk:"group_name"`
	DataCenterId                       types.String                 `tfsdk:"data_center_id"`
	MinimumNodes                       types.Int64                  `tfsdk:"minimum_nodes"`
	MaximumNodes                       types.Int64                  `tfsdk:"maximum_nodes"`
	TargetNodes                        types.Int64                  `tfsdk:"target_nodes"`
	MinimumVCpus                       types.Int64                  `tfsdk:"minimum_vcpus"`
	MaximumVCpus                       types.Int64                  `tfsdk:"maximum_vcpus"`
	TargetVCpus                        types.Int64                  `tfsdk:"target_vcpus"`
	NodeGroupPriceLimit                types.Float64                `tfsdk:"node_group_price_limit"`
	UseOnDemandInstancesInsteadOfSpots types.Bool                   `tfsdk:"use_on_demand_instances_instead_of_spots"`
	SpotPercent                        types.Int64                  `tfsdk:"spot_percent"`
	SpotMarkup                         types.Float64                `tfsdk:"spot_markup"`
	GeneratedSpotMarkup                types.Float64                `tfsdk:"generated_spot_markup"`
	ConfigurationPriorities            []configurationPriorityModel `tfsdk:"configuration_priorities"`
}

type configurationPriorityModel struct {
	VCpuType   types.String `tfsdk:"vcpu_type"`
	VCpu       types.Int64  `tfsdk:"vcpu"`
	RamGb      types.Int64  `tfsdk:"ram_gb"`
	VolumeGb   types.Int64  `tfsdk:"volume_gb"`
	VolumeType types.String `tfsdk:"volume_type"`
	Priority   types.String `tfsdk:"priority"`
}

func (r *kubernetesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	r.apiClient = client.apiClient
	r.token = client.token
}

func (r *kubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data kubernetesModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Create kubernetes cluster")

	var kubernetesCreate emmaSdk.KubernetesCreate
	ConvertToKubernetesCreateResourceRequest(data, &kubernetesCreate)

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	kubernetesGroup, response, err := r.apiClient.KubernetesClustersAPI.CreateKubernetesCluster(auth).KubernetesCreate(kubernetesCreate).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create kubernetes cluster, got error: %s,\n %v", tools.ExtractErrorMessage(response), err))
		return
	}

	var result kubernetesModel
	ConvertKubernetesResponseToResource(&result, kubernetesGroup, &data)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *kubernetesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data kubernetesModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Read kubernetes cluster")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	kubernetes, response, err := r.apiClient.KubernetesClustersAPI.GetKubernetesCluster(auth, int32(data.Id.ValueInt64())).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read kubernetes cluster, got error: %s", tools.ExtractErrorMessage(response)))
		return
	}

	var result kubernetesModel
	ConvertKubernetesResponseToResource(&result, kubernetes, &data)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *kubernetesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData kubernetesModel
	var stateData kubernetesModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Update kubernetes cluster")

	var kubernetesUpdate emmaSdk.KubernetesUpdate
	ConvertToKubernetesUpdateResourceRequest(planData, stateData, &kubernetesUpdate)
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	_, updateHttpResponse, updateError := r.apiClient.KubernetesClustersAPI.EditKubernetesCluster(auth, int32(stateData.Id.ValueInt64())).KubernetesUpdate(kubernetesUpdate).Execute()

	if updateError != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update kubernetes cluster, got error: %s", tools.ExtractErrorMessage(updateHttpResponse)))
		return
	}

	// Update response doesn't return updated nodeGroups information, so we need to perform Get request to get updated information
	// If we perform Get request immediately after Update request, it will return old information
	time.Sleep(5 * time.Second)
	getKubernetes, getHttpResponse, getError := r.apiClient.KubernetesClustersAPI.GetKubernetesCluster(auth, int32(stateData.Id.ValueInt64())).Execute()

	if getError != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get kubernetes cluster, got error: %s", tools.ExtractErrorMessage(getHttpResponse)))
		return
	}

	var result kubernetesModel
	ConvertKubernetesResponseToResource(&result, getKubernetes, &planData)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *kubernetesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data kubernetesModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Delete kubernetes cluster")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	_, response, err := r.apiClient.KubernetesClustersAPI.DeleteKubernetesCluster(auth, int32(data.Id.ValueInt64())).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete kubernetes cluster, got error: %s", tools.ExtractErrorMessage(response)))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *kubernetesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}

func ConvertToKubernetesCreateResourceRequest(data kubernetesModel, kubernetes *emmaSdk.KubernetesCreate) {
	kubernetes.Name = data.Name.ValueString()
	kubernetes.DeploymentLocation = data.DeploymentLocation.ValueString()

	var workerNodes []emmaSdk.KubernetesCreateWorkerNodesInner
	for _, node := range data.WorkerNodes {
		workerNodes = append(workerNodes, emmaSdk.KubernetesCreateWorkerNodesInner{
			Name:         node.Name.ValueString(),
			DataCenterId: node.DataCenterID.ValueString(),
			VCpuType:     node.VCpuType.ValueString(),
			VCpu:         int32(node.VCpu.ValueInt64()),
			RamGb:        int32(node.RamGb.ValueInt64()),
			VolumeType:   node.VolumeType.ValueString(),
			VolumeGb:     int32(node.VolumeGb.ValueInt64()),
		})
	}
	kubernetes.WorkerNodes = workerNodes

	autoscalingConfigs := convertToAutoScalingConfigs(data)
	kubernetes.AutoscalingConfigs = autoscalingConfigs
}

func ConvertToKubernetesUpdateResourceRequest(planData kubernetesModel, stateData kubernetesModel, kubernetes *emmaSdk.KubernetesUpdate) {
	var workerNodes []emmaSdk.KubernetesUpdateWorkerNodesInner

	stateNodesSet := make(map[string]int64)
	for _, stateNode := range stateData.WorkerNodes {
		key := fmt.Sprintf("%s-%s-%s-%d-%d-%s-%d",
			stateNode.Name.ValueString(),
			stateNode.DataCenterID.ValueString(),
			stateNode.VCpuType.ValueString(),
			stateNode.VCpu.ValueInt64(),
			stateNode.RamGb.ValueInt64(),
			stateNode.VolumeType.ValueString(),
			stateNode.VolumeGb.ValueInt64(),
		)
		stateNodesSet[key] = stateNode.Id.ValueInt64()
	}

	for _, planNode := range planData.WorkerNodes {
		var id *int32
		key := fmt.Sprintf("%s-%s-%s-%d-%d-%s-%d",
			planNode.Name.ValueString(),
			planNode.DataCenterID.ValueString(),
			planNode.VCpuType.ValueString(),
			planNode.VCpu.ValueInt64(),
			planNode.RamGb.ValueInt64(),
			planNode.VolumeType.ValueString(),
			planNode.VolumeGb.ValueInt64(),
		)
		if stateNodeId, exists := stateNodesSet[key]; exists {
			id = tools.Int64ToInt32Pointer(stateNodeId)
		} else {
			id = nil
		}

		workerNodes = append(workerNodes, emmaSdk.KubernetesUpdateWorkerNodesInner{
			Id:           id,
			Name:         planNode.Name.ValueString(),
			DataCenterId: planNode.DataCenterID.ValueString(),
			VCpuType:     planNode.VCpuType.ValueString(),
			VCpu:         int32(planNode.VCpu.ValueInt64()),
			RamGb:        int32(planNode.RamGb.ValueInt64()),
			VolumeType:   planNode.VolumeType.ValueString(),
			VolumeGb:     int32(planNode.VolumeGb.ValueInt64()),
		})
	}

	kubernetes.WorkerNodes = workerNodes

	autoscalingConfigs := convertToAutoScalingConfigs(planData)
	kubernetes.AutoscalingConfigs = autoscalingConfigs
}

func convertToAutoScalingConfigs(data kubernetesModel) []emmaSdk.KubernetesCreateAutoscalingConfigsInner {
	if data.AutoscalingConfigs == nil {
		return nil
	}
	var autoscalingConfigs []emmaSdk.KubernetesCreateAutoscalingConfigsInner
	for _, config := range *data.AutoscalingConfigs {
		autoscalingConfig := emmaSdk.KubernetesCreateAutoscalingConfigsInner{
			GroupName:                          config.GroupName.ValueString(),
			DataCenterId:                       config.DataCenterId.ValueString(),
			UseOnDemandInstancesInsteadOfSpots: config.UseOnDemandInstancesInsteadOfSpots.ValueBool(),
			SpotMarkup:                         tools.ToFloat32PointerOrNil(config.SpotMarkup),
			SpotPercent:                        tools.ToInt32PointerOrNil(config.SpotPercent),
			NodeGroupPriceLimit:                tools.ToFloat32PointerOrNil(config.NodeGroupPriceLimit),
			MinimumNodes:                       tools.ToInt32PointerOrNil(config.MinimumNodes),
			MaximumNodes:                       tools.ToInt32PointerOrNil(config.MaximumNodes),
			TargetNodes:                        tools.ToInt32PointerOrNil(config.TargetNodes),
			MinimumVCpus:                       tools.ToInt32PointerOrNil(config.MinimumVCpus),
			MaximumVCpus:                       tools.ToInt32PointerOrNil(config.MaximumVCpus),
			TargetVCpus:                        tools.ToInt32PointerOrNil(config.TargetVCpus),
		}

		var configurationPriorities []emmaSdk.KubernetesCreateAutoscalingConfigsInnerConfigurationPrioritiesInner
		for _, priority := range config.ConfigurationPriorities {
			configurationPriorities = append(configurationPriorities, emmaSdk.KubernetesCreateAutoscalingConfigsInnerConfigurationPrioritiesInner{
				VCpuType:   tools.ToPointer(priority.VCpuType.ValueString()),
				VCpu:       tools.Int64ToInt32Pointer(priority.VCpu.ValueInt64()),
				RamGb:      tools.Int64ToInt32Pointer(priority.RamGb.ValueInt64()),
				VolumeGb:   tools.Int64ToInt32Pointer(priority.VolumeGb.ValueInt64()),
				VolumeType: tools.ToPointer(priority.VolumeType.ValueString()),
				Priority:   tools.ToPointer(priority.Priority.ValueString()),
			})
		}
		autoscalingConfig.ConfigurationPriorities = configurationPriorities

		autoscalingConfigs = append(autoscalingConfigs, autoscalingConfig)
	}

	return autoscalingConfigs
}

func ConvertKubernetesResponseToResource(result *kubernetesModel, response *emmaSdk.Kubernetes, planData *kubernetesModel) {
	if response.Id != nil {
		result.Id = types.Int64Value(int64(*response.Id))
	} else {
		result.Id = types.Int64Null()
	}

	if response.Name != nil {
		result.Name = types.StringValue(*response.Name)
	} else {
		result.Name = planData.Name
	}

	if response.DeploymentLocation != nil {
		result.DeploymentLocation = types.StringValue(*response.DeploymentLocation)
	} else {
		result.DeploymentLocation = planData.DeploymentLocation
	}

	if response.DomainName != nil {
		result.DomainName = types.StringValue(*response.DomainName)
	} else {
		result.DomainName = planData.DomainName
	}

	if len(response.NodeGroups) > 0 && len(response.NodeGroups[0].Nodes) == len(planData.WorkerNodes) {
		result.WorkerNodes = make([]kubernetesWorkerNodeModel, len(response.NodeGroups[0].Nodes))
		for i, node := range response.NodeGroups[0].Nodes {
			workerNode := kubernetesWorkerNodeModel{
				Id:            types.Int64Value(int64(*node.Id)),
				GeneratedName: types.StringValue(*node.Name),
				Name:          planData.WorkerNodes[i].Name,
				DataCenterID:  planData.WorkerNodes[i].DataCenterID,
				VCpuType:      planData.WorkerNodes[i].VCpuType,
				VCpu:          planData.WorkerNodes[i].VCpu,
				RamGb:         planData.WorkerNodes[i].RamGb,
				VolumeType:    planData.WorkerNodes[i].VolumeType,
				VolumeGb:      planData.WorkerNodes[i].VolumeGb,
			}
			result.WorkerNodes[i] = workerNode
		}
	} else {
		result.WorkerNodes = planData.WorkerNodes
	}

	if planData.AutoscalingConfigs != nil && len(*planData.AutoscalingConfigs) > 0 &&
		len(response.AutoscalingConfigs) == len(*planData.AutoscalingConfigs) {

		autoscalingConfigs := make([]autoscalingConfigModel, len(response.AutoscalingConfigs))

		for i, config := range response.AutoscalingConfigs {
			autoscalingConfig := autoscalingConfigModel{
				GroupName:                          types.StringValue(*config.GroupName),
				DataCenterId:                       types.StringValue(*config.DataCenterId),
				UseOnDemandInstancesInsteadOfSpots: types.BoolValue(*config.UseOnDemandInstancesInsteadOfSpots),
				SpotMarkup:                         (*planData.AutoscalingConfigs)[i].SpotMarkup,
			}

			autoscalingConfig.NodeGroupPriceLimit = tools.GetFloat64OrDefault(config.NodeGroupPriceLimit, (*planData.AutoscalingConfigs)[i].NodeGroupPriceLimit)
			autoscalingConfig.SpotPercent = tools.GetInt64OrDefault(config.SpotPercent, (*planData.AutoscalingConfigs)[i].SpotPercent)
			autoscalingConfig.GeneratedSpotMarkup = tools.GetFloat64OrDefault(config.SpotMarkup, (*planData.AutoscalingConfigs)[i].GeneratedSpotMarkup)
			autoscalingConfig.MinimumNodes = tools.GetInt64OrDefault(config.MinimumNodes, types.Int64Null())
			autoscalingConfig.MaximumNodes = tools.GetInt64OrDefault(config.MaximumNodes, types.Int64Null())
			autoscalingConfig.TargetNodes = tools.GetInt64OrDefault(config.TargetNodes, types.Int64Null())
			autoscalingConfig.MinimumVCpus = tools.GetInt64OrDefault(config.MinimumVCpus, types.Int64Null())
			autoscalingConfig.MaximumVCpus = tools.GetInt64OrDefault(config.MaximumVCpus, types.Int64Null())
			autoscalingConfig.TargetVCpus = tools.GetInt64OrDefault(config.TargetVCpus, types.Int64Null())

			autoscalingConfig.ConfigurationPriorities = make([]configurationPriorityModel, len(config.ConfigurationPriorities))
			for j, priority := range config.ConfigurationPriorities {
				autoscalingConfig.ConfigurationPriorities[j] = configurationPriorityModel{
					VCpuType:   types.StringValue(*priority.VCpuType),
					VCpu:       types.Int64Value(int64(*priority.VCpu)),
					RamGb:      types.Int64Value(int64(*priority.RamGb)),
					VolumeGb:   types.Int64Value(int64(*priority.VolumeGb)),
					VolumeType: types.StringValue(*priority.VolumeType),
					Priority:   types.StringValue(*priority.Priority),
				}
			}
			autoscalingConfigs[i] = autoscalingConfig
		}
		result.AutoscalingConfigs = &autoscalingConfigs
	} else {
		result.AutoscalingConfigs = planData.AutoscalingConfigs
	}
}

func (r *kubernetesResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource creates a Kubernetes cluster.\n\n" +
			"A Kubernetes cluster is a set of node machines for running containerized applications. " +
			"The cluster is managed by the Kubernetes control plane, which is responsible for maintaining the desired state of the cluster.\n\n" +
			"When creating a Kubernetes cluster, provide its name, deployment location, domain name, worker nodes configuration, " +
			"autoscaling configurations, and configuration priority settings.\n\n" +
			"After creating a Kubernetes cluster, you can manage its configuration and scaling settings.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the Kubernetes cluster",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description:   "The name of the Kubernetes cluster",
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{emma.KubernetesResourceName{FieldName: "name"}},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"deployment_location": schema.StringAttribute{
				Description:   "The deployment location of the Kubernetes cluster",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"domain_name": schema.StringAttribute{
				Description: "The domain name of the Kubernetes cluster",
				Optional:    true,
				Validators:  []validator.String{emma.KubernetesResourceDomainName{}},
			},
			"worker_nodes": schema.ListNestedAttribute{
				Description: "Worker nodes configuration",
				Required:    true,
				Validators:  []validator.List{emma.UniqueField{FieldName: "name"}},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "The ID of the worker node",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the worker node",
							Optional:    true,
							Validators:  []validator.String{emma.KubernetesResourceName{FieldName: "worker_nodes.name"}},
						},
						"generated_name": schema.StringAttribute{
							Description: "The name of the worker node generated by server",
							Computed:    true,
						},
						"data_center_id": schema.StringAttribute{
							Description: "The data center ID of the worker node",
							Required:    true,
						},
						"vcpu_type": schema.StringAttribute{
							Description: "The vCPU type of the worker node",
							Required:    true,
						},
						"vcpu": schema.Int64Attribute{
							Description: "The number of vCPUs for the worker node",
							Required:    true,
						},
						"ram_gb": schema.Int64Attribute{
							Description: "The amount of RAM in GB for the worker node",
							Required:    true,
						},
						"volume_type": schema.StringAttribute{
							Description: "The volume type for the worker node",
							Required:    true,
							Validators:  []validator.String{emma.VolumeType{}},
						},
						"volume_gb": schema.Int64Attribute{
							Description: "The volume size in GB for the worker node",
							Required:    true,
						},
					},
				},
			},
			"autoscaling_configs": schema.ListNestedAttribute{
				Description: "Autoscaling configurations",
				Optional:    true,
				Validators:  []validator.List{emma.UniqueField{FieldName: "group_name"}},
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{&emma.AutoscalingConfigValidator{}},
					Attributes: map[string]schema.Attribute{
						"group_name": schema.StringAttribute{
							Description: "The name of the autoscaling group",
							Required:    true,
							Validators:  []validator.String{emma.KubernetesResourceName{FieldName: "group_name"}},
						},
						"data_center_id": schema.StringAttribute{
							Description: "The data center ID for the autoscaling group",
							Required:    true,
						},
						"minimum_nodes": schema.Int64Attribute{
							Description: "The minimum number of nodes in the autoscaling group",
							Optional:    true,
						},
						"maximum_nodes": schema.Int64Attribute{
							Description: "The maximum number of nodes in the autoscaling group",
							Optional:    true,
						},
						"target_nodes": schema.Int64Attribute{
							Description: "The target number of nodes in the autoscaling group",
							Optional:    true,
						},
						"minimum_vcpus": schema.Int64Attribute{
							Description: "The minimum number of vCPUs in the autoscaling group",
							Optional:    true,
						},
						"maximum_vcpus": schema.Int64Attribute{
							Description: "The maximum number of vCPUs in the autoscaling group",
							Optional:    true,
						},
						"target_vcpus": schema.Int64Attribute{
							Description: "The target number of vCPUs in the autoscaling group",
							Optional:    true,
						},
						"node_group_price_limit": schema.Float64Attribute{
							Description: "The price limit for the node group",
							Optional:    true,
							Validators:  []validator.Float64{&emma.NodeGroupPriceLimit{}},
						},
						"use_on_demand_instances_instead_of_spots": schema.BoolAttribute{
							Description: "Whether to use on-demand instances instead of spot instances",
							Required:    true,
						},
						"spot_percent": schema.Int64Attribute{
							Description: "The percentage of spot instances to use",
							Optional:    true,
							Validators:  []validator.Int64{&emma.SpotPercent{}},
						},
						"spot_markup": schema.Float64Attribute{
							Description: "The markup for spot instances",
							Optional:    true,
							Validators:  []validator.Float64{&emma.SpotMarkup{}},
						},
						"generated_spot_markup": schema.Float64Attribute{
							Description: "The markup for spot instances generated by server",
							Computed:    true,
						},
						"configuration_priorities": schema.ListNestedAttribute{
							Description: "Configuration priorities settings",
							Optional:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"vcpu_type": schema.StringAttribute{
										Description: "The vCPU type for the configuration priority",
										Required:    true,
										Validators:  []validator.String{emma.VCPUType{}},
									},
									"vcpu": schema.Int64Attribute{
										Description: "The number of vCPUs for the configuration priority",
										Required:    true,
									},
									"ram_gb": schema.Int64Attribute{
										Description: "The amount of RAM in GB for the configuration priority",
										Required:    true,
									},
									"volume_gb": schema.Int64Attribute{
										Description: "The volume size in GB for the configuration priority",
										Required:    true,
									},
									"volume_type": schema.StringAttribute{
										Description: "The volume type for the worker node",
										Required:    true,
										Validators:  []validator.String{emma.VolumeType{}},
									},
									"priority": schema.StringAttribute{
										Description: "The priority level for the configuration",
										Required:    true,
										Validators:  []validator.String{emma.Priority{}},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
