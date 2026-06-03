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

var _ datasource.DataSource = &acceleratorTypesDataSource{}

func NewAcceleratorTypesDataSource() datasource.DataSource {
	return &acceleratorTypesDataSource{}
}

// acceleratorTypesDataSource defines the data source implementation.
type acceleratorTypesDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// acceleratorTypesDataSourceModel describes the data source data model.
type acceleratorTypesDataSourceModel struct {
	AcceleratorTypes types.List `tfsdk:"accelerator_types"`
	IdsByName        types.Map  `tfsdk:"ids_by_name"`
}

// acceleratorTypeItemModel describes individual accelerator type items in the list.
type acceleratorTypeItemModel struct {
	Id              types.String `tfsdk:"id"`
	AcceleratorType types.String `tfsdk:"accelerator_type"`
}

func (d *acceleratorTypesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_accelerator_types"
}

func (d *acceleratorTypesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of all available GPU accelerator types that can be used with compute instances.",
		Attributes: map[string]schema.Attribute{
			"accelerator_types": schema.ListNestedAttribute{
				Description: "List of available accelerator types",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the accelerator type",
							Computed:    true,
						},
						"accelerator_type": schema.StringAttribute{
							Description: "Name of the accelerator type (e.g. NVIDIA A100, NVIDIA T4)",
							Computed:    true,
						},
					},
				},
			},
			"ids_by_name": schema.MapAttribute{
				Description: "Map from accelerator type name to its id, e.g. " +
					"`accelerator_type_id = data.emma_accelerator_types.all.ids_by_name[\"NVIDIA T4\"]`.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *acceleratorTypesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.apiClient = client.apiClient
	d.token = client.token
}

func (d *acceleratorTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data acceleratorTypesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)
	acceleratorTypes, response, err := d.apiClient.AcceleratorTypesAPI.GetAcceleratorTypes(auth).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read accelerator types, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	typeList, diags := convertAcceleratorTypesToList(ctx, acceleratorTypes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.AcceleratorTypes = typeList

	idsByName, mapDiags := buildAcceleratorIdsByName(ctx, acceleratorTypes)
	resp.Diagnostics.Append(mapDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.IdsByName = idsByName

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper function to get attribute types for accelerator type nested object
func (o acceleratorTypeItemModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"accelerator_type": types.StringType,
	}
}

// buildAcceleratorIdsByName builds a map from accelerator type name to its id.
// Items missing either field are skipped. On duplicate names the first one wins
// (in practice names are unique on the platform).
func buildAcceleratorIdsByName(ctx context.Context, acceleratorTypes []emmaSdk.AcceleratorType) (types.Map, diag.Diagnostics) {
	pairs := make(map[string]attr.Value, len(acceleratorTypes))
	for _, at := range acceleratorTypes {
		if at.Id == nil || at.AcceleratorType == nil {
			continue
		}
		name := *at.AcceleratorType
		if _, exists := pairs[name]; exists {
			continue
		}
		pairs[name] = types.StringValue(*at.Id)
	}
	return types.MapValue(types.StringType, pairs)
}

// convertAcceleratorTypesToList converts Emma API accelerator types response to Terraform list
func convertAcceleratorTypesToList(ctx context.Context, acceleratorTypes []emmaSdk.AcceleratorType) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	var itemModels []acceleratorTypeItemModel

	for _, at := range acceleratorTypes {
		item := acceleratorTypeItemModel{}

		if at.Id != nil {
			item.Id = types.StringValue(*at.Id)
		} else {
			item.Id = types.StringNull()
		}

		if at.AcceleratorType != nil {
			item.AcceleratorType = types.StringValue(*at.AcceleratorType)
		} else {
			item.AcceleratorType = types.StringNull()
		}

		itemModels = append(itemModels, item)
	}

	typeList, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: acceleratorTypeItemModel{}.attrTypes()}, itemModels)
	diags.Append(listDiags...)

	return typeList, diags
}
