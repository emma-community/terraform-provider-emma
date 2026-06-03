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

var _ datasource.DataSource = &securityGroupsDataSource{}

func NewSecurityGroupsDataSource() datasource.DataSource {
	return &securityGroupsDataSource{}
}

// securityGroupsDataSource defines the data source implementation.
type securityGroupsDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// securityGroupsDataSourceModel describes the data source data model.
type securityGroupsDataSourceModel struct {
	ProjectId      types.Int64 `tfsdk:"project_id"`
	SecurityGroups types.List  `tfsdk:"security_groups"`
}

// securityGroupItemModel describes individual security group items in the list.
type securityGroupItemModel struct {
	Id                    types.Int64  `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	CreatedAt             types.String `tfsdk:"created_at"`
	CreatedByName         types.String `tfsdk:"created_by_name"`
	ModifiedAt            types.String `tfsdk:"modified_at"`
	SynchronizationStatus types.String `tfsdk:"synchronization_status"`
	RecomposingStatus     types.String `tfsdk:"recomposing_status"`
}

func (d *securityGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_groups"
}

func (d *securityGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of security groups in the Emma platform. Optionally filter by project ID.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.Int64Attribute{
				Description: "Project ID to filter security groups",
				Optional:    true,
			},
			"security_groups": schema.ListNestedAttribute{
				Description: "List of security groups",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "ID of the security group",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the security group",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Date and time when the security group was created",
							Computed:    true,
						},
						"created_by_name": schema.StringAttribute{
							Description: "Name of the user who created the security group",
							Computed:    true,
						},
						"modified_at": schema.StringAttribute{
							Description: "Date and time when the security group was last modified",
							Computed:    true,
						},
						"synchronization_status": schema.StringAttribute{
							Description: "Synchronization status of the security group rules across providers",
							Computed:    true,
						},
						"recomposing_status": schema.StringAttribute{
							Description: "Recomposing status of the security group when new VMs are added",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *securityGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *securityGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data securityGroupsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	apiReq := d.apiClient.SecurityGroupsAPI.GetSecurityGroups(auth)
	if !data.ProjectId.IsNull() {
		apiReq = apiReq.ProjectId(int32(data.ProjectId.ValueInt64()))
	}
	securityGroups, response, err := apiReq.Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read security groups, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	sgList, diags := convertSecurityGroupsToList(ctx, securityGroups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.SecurityGroups = sgList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to get attribute types for security group nested object
func (o securityGroupItemModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.Int64Type,
		"name":                   types.StringType,
		"created_at":             types.StringType,
		"created_by_name":        types.StringType,
		"modified_at":            types.StringType,
		"synchronization_status": types.StringType,
		"recomposing_status":     types.StringType,
	}
}

// convertSecurityGroupsToList converts Emma API security groups response to Terraform list
func convertSecurityGroupsToList(ctx context.Context, securityGroups []emmaSdk.SecurityGroup) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	itemModels := make([]securityGroupItemModel, 0, len(securityGroups))

	for _, sg := range securityGroups {
		item := securityGroupItemModel{}

		if sg.Id != nil {
			item.Id = types.Int64Value(int64(*sg.Id))
		} else {
			item.Id = types.Int64Null()
		}

		if sg.Name != nil {
			item.Name = types.StringValue(*sg.Name)
		} else {
			item.Name = types.StringNull()
		}

		if sg.CreatedAt != nil {
			item.CreatedAt = types.StringValue(*sg.CreatedAt)
		} else {
			item.CreatedAt = types.StringNull()
		}

		if sg.CreatedByName != nil {
			item.CreatedByName = types.StringValue(*sg.CreatedByName)
		} else {
			item.CreatedByName = types.StringNull()
		}

		if sg.ModifiedAt != nil {
			item.ModifiedAt = types.StringValue(*sg.ModifiedAt)
		} else {
			item.ModifiedAt = types.StringNull()
		}

		if sg.SynchronizationStatus != nil {
			item.SynchronizationStatus = types.StringValue(*sg.SynchronizationStatus)
		} else {
			item.SynchronizationStatus = types.StringNull()
		}

		if sg.RecomposingStatus != nil {
			item.RecomposingStatus = types.StringValue(*sg.RecomposingStatus)
		} else {
			item.RecomposingStatus = types.StringNull()
		}

		itemModels = append(itemModels, item)
	}

	sgList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: securityGroupItemModel{}.attrTypes()}, itemModels)
	diags.Append(listDiags...)

	return sgList, diags
}
