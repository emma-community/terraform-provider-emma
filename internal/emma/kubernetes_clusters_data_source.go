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

var _ datasource.DataSource = &kubernetesClustersDataSource{}

func NewKubernetesClustersDataSource() datasource.DataSource {
	return &kubernetesClustersDataSource{}
}

// kubernetesClustersDataSource defines the data source implementation.
type kubernetesClustersDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// kubernetesClustersDataSourceModel describes the data source data model.
type kubernetesClustersDataSourceModel struct {
	ProjectId           types.Int64 `tfsdk:"project_id"`
	KubernetesClusters  types.List  `tfsdk:"kubernetes_clusters"`
}

// kubernetesClusterItemModel describes individual Kubernetes cluster items in the list.
type kubernetesClusterItemModel struct {
	Id                 types.Int64  `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Status             types.String `tfsdk:"status"`
	ControlPlaneStatus types.String `tfsdk:"control_plane_status"`
	Version            types.String `tfsdk:"version"`
	DeploymentLocation types.String `tfsdk:"deployment_location"`
	K8sConnectionType  types.String `tfsdk:"k8s_connection_type"`
	DomainName         types.String `tfsdk:"domain_name"`
	CreatedAt          types.String `tfsdk:"created_at"`
	CreatedByName      types.String `tfsdk:"created_by_name"`
	CreatedById        types.Int64  `tfsdk:"created_by_id"`
	ModifiedAt         types.String `tfsdk:"modified_at"`
	ModifiedByName     types.String `tfsdk:"modified_by_name"`
	ModifiedById       types.Int64  `tfsdk:"modified_by_id"`
}

func (d *kubernetesClustersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_clusters"
}

func (d *kubernetesClustersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of Kubernetes clusters. Optionally filter by project ID.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.Int64Attribute{
				Description: "Optional project ID to filter clusters",
				Optional:    true,
			},
			"kubernetes_clusters": schema.ListNestedAttribute{
				Description: "List of Kubernetes clusters",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Kubernetes cluster ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Kubernetes cluster name",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the Kubernetes cluster",
							Computed:    true,
						},
						"control_plane_status": schema.StringAttribute{
							Description: "Control plane status",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Kubernetes cluster version",
							Computed:    true,
						},
						"deployment_location": schema.StringAttribute{
							Description: "Deployment region of the Kubernetes cluster",
							Computed:    true,
						},
						"k8s_connection_type": schema.StringAttribute{
							Description: "Kubernetes connection type",
							Computed:    true,
						},
						"domain_name": schema.StringAttribute{
							Description: "Domain attached to the Kubernetes cluster",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Date and time of the Kubernetes cluster creation",
							Computed:    true,
						},
						"created_by_name": schema.StringAttribute{
							Description: "Name of the user who created the Kubernetes cluster",
							Computed:    true,
						},
						"created_by_id": schema.Int64Attribute{
							Description: "ID of the user who created the Kubernetes cluster",
							Computed:    true,
						},
						"modified_at": schema.StringAttribute{
							Description: "Date and time when the Kubernetes cluster was last modified",
							Computed:    true,
						},
						"modified_by_name": schema.StringAttribute{
							Description: "Name of the user who last modified the Kubernetes cluster",
							Computed:    true,
						},
						"modified_by_id": schema.Int64Attribute{
							Description: "ID of the user who last modified the Kubernetes cluster",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *kubernetesClustersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *kubernetesClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data kubernetesClustersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	apiReq := d.apiClient.KubernetesClustersAPI.GetKubernetesClusters(auth)
	if !data.ProjectId.IsNull() && !data.ProjectId.IsUnknown() {
		apiReq = apiReq.ProjectId(int32(data.ProjectId.ValueInt64()))
	}

	clusters, response, err := apiReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read Kubernetes clusters, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	clusterList, diags := convertKubernetesClustersToList(ctx, clusters)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.KubernetesClusters = clusterList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to get attribute types for Kubernetes cluster nested object
func (o kubernetesClusterItemModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                   types.Int64Type,
		"name":                 types.StringType,
		"status":               types.StringType,
		"control_plane_status": types.StringType,
		"version":              types.StringType,
		"deployment_location":  types.StringType,
		"k8s_connection_type":  types.StringType,
		"domain_name":          types.StringType,
		"created_at":           types.StringType,
		"created_by_name":      types.StringType,
		"created_by_id":        types.Int64Type,
		"modified_at":          types.StringType,
		"modified_by_name":     types.StringType,
		"modified_by_id":       types.Int64Type,
	}
}

// convertKubernetesClustersToList converts Emma API Kubernetes clusters response to Terraform list
func convertKubernetesClustersToList(ctx context.Context, clusters []emmaSdk.KubernetesListResponseInner) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var itemModels []kubernetesClusterItemModel

	for _, c := range clusters {
		item := kubernetesClusterItemModel{}

		if c.Id != nil {
			item.Id = types.Int64Value(int64(*c.Id))
		} else {
			item.Id = types.Int64Null()
		}

		if c.Name != nil {
			item.Name = types.StringValue(*c.Name)
		} else {
			item.Name = types.StringNull()
		}

		if c.Status != nil {
			item.Status = types.StringValue(*c.Status)
		} else {
			item.Status = types.StringNull()
		}

		if c.ControlPlaneStatus != nil {
			item.ControlPlaneStatus = types.StringValue(*c.ControlPlaneStatus)
		} else {
			item.ControlPlaneStatus = types.StringNull()
		}

		if c.Version != nil {
			item.Version = types.StringValue(*c.Version)
		} else {
			item.Version = types.StringNull()
		}

		if c.DeploymentLocation != nil {
			item.DeploymentLocation = types.StringValue(*c.DeploymentLocation)
		} else {
			item.DeploymentLocation = types.StringNull()
		}

		if c.K8sConnectionType != nil {
			item.K8sConnectionType = types.StringValue(*c.K8sConnectionType)
		} else {
			item.K8sConnectionType = types.StringNull()
		}

		if c.DomainName != nil {
			item.DomainName = types.StringValue(*c.DomainName)
		} else {
			item.DomainName = types.StringNull()
		}

		if c.CreatedAt != nil {
			item.CreatedAt = types.StringValue(*c.CreatedAt)
		} else {
			item.CreatedAt = types.StringNull()
		}

		if c.CreatedByName != nil {
			item.CreatedByName = types.StringValue(*c.CreatedByName)
		} else {
			item.CreatedByName = types.StringNull()
		}

		if c.CreatedById != nil {
			item.CreatedById = types.Int64Value(int64(*c.CreatedById))
		} else {
			item.CreatedById = types.Int64Null()
		}

		if c.ModifiedAt != nil {
			item.ModifiedAt = types.StringValue(*c.ModifiedAt)
		} else {
			item.ModifiedAt = types.StringNull()
		}

		if c.ModifiedByName != nil {
			item.ModifiedByName = types.StringValue(*c.ModifiedByName)
		} else {
			item.ModifiedByName = types.StringNull()
		}

		if c.ModifiedById != nil {
			item.ModifiedById = types.Int64Value(int64(*c.ModifiedById))
		} else {
			item.ModifiedById = types.Int64Null()
		}

		itemModels = append(itemModels, item)
	}

	clusterList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: kubernetesClusterItemModel{}.attrTypes()}, itemModels)
	diags.Append(listDiags...)

	return clusterList, diags
}
