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

var _ datasource.DataSource = &workflowTemplatesDataSource{}

func NewWorkflowTemplatesDataSource() datasource.DataSource {
	return &workflowTemplatesDataSource{}
}

type workflowTemplatesDataSource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

type workflowTemplatesDataSourceModel struct {
	NameLike          types.String `tfsdk:"name_like"`
	Status            types.String `tfsdk:"status"`
	WorkflowTemplates types.List   `tfsdk:"workflow_templates"`
}

type workflowTemplateItemModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ContentType    types.String `tfsdk:"content_type"`
	Content        types.String `tfsdk:"content"`
	Status         types.String `tfsdk:"status"`
	ResourceType   types.String `tfsdk:"resource_type"`
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedByName  types.String `tfsdk:"created_by_name"`
	CreatedById    types.String `tfsdk:"created_by_id"`
	ModifiedAt     types.String `tfsdk:"modified_at"`
	ModifiedByName types.String `tfsdk:"modified_by_name"`
	ModifiedById   types.String `tfsdk:"modified_by_id"`
	IsDeleted      types.Bool   `tfsdk:"is_deleted"`
	IsContentValid types.Bool   `tfsdk:"is_content_valid"`
}

func workflowTemplateItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"name":             types.StringType,
		"description":      types.StringType,
		"content_type":     types.StringType,
		"content":          types.StringType,
		"status":           types.StringType,
		"resource_type":    types.StringType,
		"created_at":       types.StringType,
		"created_by_name":  types.StringType,
		"created_by_id":    types.StringType,
		"modified_at":      types.StringType,
		"modified_by_name": types.StringType,
		"modified_by_id":   types.StringType,
		"is_deleted":       types.BoolType,
		"is_content_valid": types.BoolType,
	}
}

func (d *workflowTemplatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_templates"
}

func (d *workflowTemplatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a list of workflow templates in the Emma platform. Optionally filter by name or status.",
		Attributes: map[string]schema.Attribute{
			"name_like": schema.StringAttribute{
				Description: "Filter workflow templates by name using partial match.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Filter workflow templates by status (e.g. PUBLISHED or UNPUBLISHED).",
				Optional:    true,
			},
			"workflow_templates": schema.ListNestedAttribute{
				Description: "List of workflow templates.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "ID of the workflow template.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the workflow template.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the workflow template.",
							Computed:    true,
						},
						"content_type": schema.StringAttribute{
							Description: "Format of the content field.",
							Computed:    true,
						},
						"content": schema.StringAttribute{
							Description: "Content of the workflow template.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Status of the workflow template.",
							Computed:    true,
						},
						"resource_type": schema.StringAttribute{
							Description: "Type of the resource.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Date and time when the workflow template was created.",
							Computed:    true,
						},
						"created_by_name": schema.StringAttribute{
							Description: "Name of the user who created the workflow template.",
							Computed:    true,
						},
						"created_by_id": schema.StringAttribute{
							Description: "ID of the user who created the workflow template.",
							Computed:    true,
						},
						"modified_at": schema.StringAttribute{
							Description: "Date and time when the workflow template was last modified.",
							Computed:    true,
						},
						"modified_by_name": schema.StringAttribute{
							Description: "Name of the user who last modified the workflow template.",
							Computed:    true,
						},
						"modified_by_id": schema.StringAttribute{
							Description: "ID of the user who last modified the workflow template.",
							Computed:    true,
						},
						"is_deleted": schema.BoolAttribute{
							Description: "Indicates whether the workflow template is deleted.",
							Computed:    true,
						},
						"is_content_valid": schema.BoolAttribute{
							Description: "Indicates whether the content of the workflow template is valid.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *workflowTemplatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workflowTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data workflowTemplatesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *d.token.AccessToken)

	var allTemplates []emmaSdk.WorkflowTemplate
	const pageSize int32 = 100
	var page int32 = 0

	for {
		apiReq := d.apiClient.WorkflowsAPI.GetWorkflowTemplates(auth).
			Page(page).
			Size(pageSize)

		if !data.NameLike.IsNull() && !data.NameLike.IsUnknown() {
			apiReq = apiReq.NameLike(data.NameLike.ValueString())
		}
		if !data.Status.IsNull() && !data.Status.IsUnknown() {
			apiReq = apiReq.Status(data.Status.ValueString())
		}

		result, response, err := apiReq.Execute()
		if err != nil {
			apiErr := tools.ExtractErrorMessage(response)
			if apiErr == "" {
				apiErr = err.Error()
			}
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to read workflow templates, got error: %s", apiErr))
			return
		}

		if result != nil && len(result.Content) > 0 {
			allTemplates = append(allTemplates, result.Content...)
		}

		if result == nil || result.Last == nil || *result.Last {
			break
		}
		page++
	}

	templateList, diags := convertWorkflowTemplatesToList(ctx, allTemplates)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.WorkflowTemplates = templateList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func convertWorkflowTemplatesToList(ctx context.Context, templates []emmaSdk.WorkflowTemplate) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	itemType := types.ObjectType{AttrTypes: workflowTemplateItemAttrTypes()}

	if len(templates) == 0 {
		return types.ListValueMust(itemType, []attr.Value{}), diags
	}

	items := make([]workflowTemplateItemModel, 0, len(templates))
	for _, t := range templates {
		item := workflowTemplateItemModel{}

		id := t.Id
		item.Id = convertInt32ToStringValue(&id)
		item.Name = types.StringValue(t.Name)
		item.ContentType = types.StringValue(t.ContentType)
		item.Content = types.StringValue(t.Content)
		item.Status = types.StringValue(t.Status)
		item.ResourceType = types.StringValue(t.ResourceType)
		if t.CreatedAt != nil {
			item.CreatedAt = types.StringValue(*t.CreatedAt)
		} else {
			item.CreatedAt = types.StringNull()
		}
		item.CreatedByName = types.StringValue(t.CreatedByName)
		createdById := t.CreatedById
		item.CreatedById = convertInt32ToStringValue(&createdById)
		item.ModifiedAt = types.StringValue(t.ModifiedAt)
		item.ModifiedByName = types.StringValue(t.ModifiedByName)
		modifiedById := t.ModifiedById
		item.ModifiedById = convertInt32ToStringValue(&modifiedById)
		item.IsDeleted = types.BoolValue(t.IsDeleted)

		if t.Description != nil {
			item.Description = types.StringValue(*t.Description)
		} else {
			item.Description = types.StringNull()
		}

		if t.IsContentValid != nil {
			item.IsContentValid = types.BoolValue(*t.IsContentValid)
		} else {
			item.IsContentValid = types.BoolNull()
		}

		items = append(items, item)
	}

	listVal, listDiags := types.ListValueFrom(ctx, itemType, items)
	diags.Append(listDiags...)
	return listVal, diags
}

func convertInt32ToStringValue(v *int32) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(fmt.Sprintf("%d", *v))
}
