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

var _ datasource.DataSource = &volumesDataSource{}

func NewVolumesDataSource() datasource.DataSource {
	return &volumesDataSource{}
}

// volumesDataSource defines the data source implementation.
type volumesDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// volumesDataSourceModel describes the data source data model.
type volumesDataSourceModel struct {
	Volumes types.List `tfsdk:"volumes"`
}

// volumeItemModel describes individual volume items in the list.
type volumeItemModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	SizeGb         types.Int64  `tfsdk:"size_gb"`
	VolumeType     types.String `tfsdk:"volume_type"`
	IsSystem       types.Bool   `tfsdk:"is_system"`
	Status         types.String `tfsdk:"status"`
	AttachedToId   types.Int64  `tfsdk:"attached_to_id"`
	ProjectId      types.Int64  `tfsdk:"project_id"`
	DataCenterId   types.String `tfsdk:"data_center_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedByName  types.String `tfsdk:"created_by_name"`
	CreatedById    types.Int64  `tfsdk:"created_by_id"`
	ModifiedAt     types.String `tfsdk:"modified_at"`
	ModifiedByName types.String `tfsdk:"modified_by_name"`
	ModifiedById   types.Int64  `tfsdk:"modified_by_id"`
}

func (d *volumesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

func (d *volumesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of all storage volumes in the project.",
		Attributes: map[string]schema.Attribute{
			"volumes": schema.ListNestedAttribute{
				Description: "List of volumes",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Volume ID",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Volume name",
							Computed:    true,
						},
						"size_gb": schema.Int64Attribute{
							Description: "Volume size in gigabytes",
							Computed:    true,
						},
						"volume_type": schema.StringAttribute{
							Description: "Volume type (e.g., ssd, hdd)",
							Computed:    true,
						},
						"is_system": schema.BoolAttribute{
							Description: "Indicates whether the volume contains the operating system",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the volume",
							Computed:    true,
						},
						"attached_to_id": schema.Int64Attribute{
							Description: "ID of the compute instance the volume is attached to",
							Computed:    true,
						},
						"project_id": schema.Int64Attribute{
							Description: "Project ID owning the volume",
							Computed:    true,
						},
						"data_center_id": schema.StringAttribute{
							Description: "Data center ID where the volume is located",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Creation timestamp",
							Computed:    true,
						},
						"created_by_name": schema.StringAttribute{
							Description: "Name of the user who created the volume",
							Computed:    true,
						},
						"created_by_id": schema.Int64Attribute{
							Description: "ID of the user who created the volume",
							Computed:    true,
						},
						"modified_at": schema.StringAttribute{
							Description: "Date and time when the volume was last modified",
							Computed:    true,
						},
						"modified_by_name": schema.StringAttribute{
							Description: "Name of the user who last modified the volume",
							Computed:    true,
						},
						"modified_by_id": schema.Int64Attribute{
							Description: "ID of the user who last modified the volume",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *volumesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *volumesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data volumesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	volumes, response, err := d.apiClient.VolumesAPI.GetVolumes(auth).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read volumes, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	volumeList, diags := convertVolumesToList(ctx, volumes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Volumes = volumeList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to get attribute types for volume nested object
func (o volumeItemModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"name":             types.StringType,
		"size_gb":          types.Int64Type,
		"volume_type":      types.StringType,
		"is_system":        types.BoolType,
		"status":           types.StringType,
		"attached_to_id":   types.Int64Type,
		"project_id":       types.Int64Type,
		"data_center_id":   types.StringType,
		"created_at":       types.StringType,
		"created_by_name":  types.StringType,
		"created_by_id":    types.Int64Type,
		"modified_at":      types.StringType,
		"modified_by_name": types.StringType,
		"modified_by_id":   types.Int64Type,
	}
}

// convertVolumesToList converts Emma API volumes response to Terraform list
func convertVolumesToList(ctx context.Context, volumes []emmaSdk.Volume) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	itemModels := make([]volumeItemModel, 0, len(volumes))

	for _, v := range volumes {
		item := volumeItemModel{}

		if v.Id != nil {
			item.Id = types.StringValue(fmt.Sprintf("%d", *v.Id))
		} else {
			item.Id = types.StringNull()
		}

		if v.Name != nil {
			item.Name = types.StringValue(*v.Name)
		} else {
			item.Name = types.StringNull()
		}

		if v.SizeGb != nil {
			item.SizeGb = types.Int64Value(int64(*v.SizeGb))
		} else {
			item.SizeGb = types.Int64Null()
		}

		if v.Type != nil {
			item.VolumeType = types.StringValue(*v.Type)
		} else {
			item.VolumeType = types.StringNull()
		}

		if v.IsSystem != nil {
			item.IsSystem = types.BoolValue(*v.IsSystem)
		} else {
			item.IsSystem = types.BoolNull()
		}

		if v.Status != nil {
			item.Status = types.StringValue(*v.Status)
		} else {
			item.Status = types.StringNull()
		}

		if v.AttachedToId != nil {
			item.AttachedToId = types.Int64Value(int64(*v.AttachedToId))
		} else {
			item.AttachedToId = types.Int64Null()
		}

		if v.ProjectId != nil {
			item.ProjectId = types.Int64Value(int64(*v.ProjectId))
		} else {
			item.ProjectId = types.Int64Null()
		}

		if v.DataCenter != nil && v.DataCenter.Id != nil {
			item.DataCenterId = types.StringValue(*v.DataCenter.Id)
		} else {
			item.DataCenterId = types.StringNull()
		}

		if v.CreatedAt != nil {
			item.CreatedAt = types.StringValue(*v.CreatedAt)
		} else {
			item.CreatedAt = types.StringNull()
		}

		if v.CreatedByName != nil {
			item.CreatedByName = types.StringValue(*v.CreatedByName)
		} else {
			item.CreatedByName = types.StringNull()
		}

		if v.CreatedById != nil {
			item.CreatedById = types.Int64Value(int64(*v.CreatedById))
		} else {
			item.CreatedById = types.Int64Null()
		}

		if v.ModifiedAt != nil {
			item.ModifiedAt = types.StringValue(*v.ModifiedAt)
		} else {
			item.ModifiedAt = types.StringNull()
		}

		if v.ModifiedByName != nil {
			item.ModifiedByName = types.StringValue(*v.ModifiedByName)
		} else {
			item.ModifiedByName = types.StringNull()
		}

		if v.ModifiedById != nil {
			item.ModifiedById = types.Int64Value(int64(*v.ModifiedById))
		} else {
			item.ModifiedById = types.Int64Null()
		}

		itemModels = append(itemModels, item)
	}

	volumeList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: volumeItemModel{}.attrTypes()}, itemModels)
	diags.Append(listDiags...)

	return volumeList, diags
}
