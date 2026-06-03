package emma

import (
	"context"
	"fmt"
	emmaSdk "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/convert"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/errors"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/state"
	"github.com/emma-community/terraform-provider-emma/tools"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"
)

var _ resource.Resource = &workflowTemplateResource{}
var _ resource.ResourceWithImportState = &workflowTemplateResource{}

func NewWorkflowTemplateResource() resource.Resource {
	return &workflowTemplateResource{}
}

type workflowTemplateResource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

type workflowTemplateResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	ContentType    types.String `tfsdk:"content_type"`
	Content        types.String `tfsdk:"content"`
	Status         types.String `tfsdk:"status"`
	ResourceType   types.String `tfsdk:"resource_type"`
	ResourceParams types.List   `tfsdk:"resource_params"`
	Tags           types.List   `tfsdk:"tags"`
	ContentParams  types.List   `tfsdk:"content_params"`
	CreatedAt      types.String `tfsdk:"created_at"`
	CreatedByName  types.String `tfsdk:"created_by_name"`
	CreatedById    types.String `tfsdk:"created_by_id"`
	ModifiedAt     types.String `tfsdk:"modified_at"`
	ModifiedByName types.String `tfsdk:"modified_by_name"`
	ModifiedById   types.String `tfsdk:"modified_by_id"`
	IsDeleted      types.Bool   `tfsdk:"is_deleted"`
	IsContentValid types.Bool   `tfsdk:"is_content_valid"`
}

type workflowTemplateResourceParamModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type workflowTemplateTagModel struct {
	TagId types.String `tfsdk:"tag_id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type workflowTemplateContentParamModel struct {
	Name         types.String `tfsdk:"name"`
	DefaultValue types.String `tfsdk:"default_value"`
	Mandatory    types.Bool   `tfsdk:"mandatory"`
}

func workflowTemplateResourceParamAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	}
}

func workflowTemplateTagAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"tag_id": types.StringType,
		"key":    types.StringType,
		"value":  types.StringType,
	}
}

func workflowTemplateContentParamAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":          types.StringType,
		"default_value": types.StringType,
		"mandatory":     types.BoolType,
	}
}

func (r *workflowTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_template"
}

func (r *workflowTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a workflow template. A workflow template is a predefined blueprint for " +
			"a workflow that includes shell commands and parameters. It serves as a starting point for creating " +
			"workflows, allowing users to quickly set up and execute common processes.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "ID of the workflow template",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Name of the workflow template",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the workflow template",
				Optional:    true,
				Computed:    false,
			},
			"content_type": schema.StringAttribute{
				Description: "Format of the content field. Only Shell is supported for now.",
				Required:    true,
			},
			"content": schema.StringAttribute{
				Description: "Content of the workflow template. For Shell contentType, includes shell commands to execute.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the workflow template. Can be PUBLISHED or UNPUBLISHED.",
				Required:    true,
			},
			"resource_type": schema.StringAttribute{
				Description: "Type of the resource, in this case COMPUTE.",
				Required:    true,
			},
			"resource_params": schema.ListNestedAttribute{
				Description: "Parameters of the resource limitation (e.g. minCpuCores, minRamSizeGb).",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the resource parameter (e.g. minCpuCores, minRamSizeGb, minVolumeSizeGb).",
							Optional:    true,
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the resource parameter.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"tags": schema.ListNestedAttribute{
				Description: "List of tags assigned to the workflow template.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"tag_id": schema.StringAttribute{
							Description: "ID of the tag.",
							Optional:    true,
							Computed:    true,
						},
						"key": schema.StringAttribute{
							Description: "Key of the tag (e.g. environment, department, project).",
							Optional:    true,
							Computed:    true,
						},
						"value": schema.StringAttribute{
							Description: "Value of the tag (e.g. production, staging, development).",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"content_params": schema.ListNestedAttribute{
				Description: "List of parameters extracted from the content. These parameters fill in template placeholders when creating a workflow from this template.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the content parameter.",
							Optional:    true,
							Computed:    true,
						},
						"default_value": schema.StringAttribute{
							Description: "Default value of the content parameter.",
							Optional:    true,
							Computed:    true,
						},
						"mandatory": schema.BoolAttribute{
							Description: "Whether the content parameter is mandatory when creating a workflow.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description:   "Date and time when the workflow template was created.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_by_name": schema.StringAttribute{
				Description:   "Name of the user who created the workflow template.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"created_by_id": schema.StringAttribute{
				Description:   "ID of the user who created the workflow template.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
	}
}

func (r *workflowTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.",
				req.ProviderData))
		return
	}
	r.apiClient = client.apiClient
	r.token = client.token
}

func (r *workflowTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data workflowTemplateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Create workflow template")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)

	createReq, diags := convertToWorkflowTemplateCreateRequest(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	template, response, err := r.apiClient.WorkflowsAPI.CreateWorkflowTemplate(auth).
		WorkflowTemplateCreate(createReq).Execute()

	if err != nil {
		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}

		resourceErr := errors.NewError("emma_workflow_template", "Create").
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	diags = convertWorkflowTemplateToResource(ctx, &data, template)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workflowTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data workflowTemplateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Read workflow template")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)

	id, err := convert.StringToInt32(data.Id)
	if err != nil {
		resourceErr := errors.NewError("emma_workflow_template", "Read").
			WithID(data.Id.ValueString()).
			WithMessage("Invalid workflow template ID format").
			WithCause(err).
			Build()
		resp.Diagnostics.AddError("Validation Error", resourceErr.Error())
		return
	}

	template, response, err := r.apiClient.WorkflowsAPI.GetWorkflowTemplate(auth, id).Execute()

	if err != nil {
		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)

			if statusCode == http.StatusNotFound {
				stateManager := state.NewStateManager(ctx)
				stateManager.RemoveFromState(resp)
				return
			}
		}

		resourceErr := errors.NewError("emma_workflow_template", "Read").
			WithID(data.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	diags := convertWorkflowTemplateToResource(ctx, &data, template)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *workflowTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData workflowTemplateResourceModel
	var stateData workflowTemplateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Update workflow template")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)

	id, err := convert.StringToInt32(stateData.Id)
	if err != nil {
		resourceErr := errors.NewError("emma_workflow_template", "Update").
			WithID(stateData.Id.ValueString()).
			WithMessage("Invalid workflow template ID format").
			WithCause(err).
			Build()
		resp.Diagnostics.AddError("Validation Error", resourceErr.Error())
		return
	}

	updateReq, diags := convertToWorkflowTemplateCreateRequest(ctx, planData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	template, response, err := r.apiClient.WorkflowsAPI.UpdateWorkflowTemplate(auth, id).
		WorkflowTemplateCreate(updateReq).Execute()

	if err != nil {
		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}

		resourceErr := errors.NewError("emma_workflow_template", "Update").
			WithID(stateData.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	diags = convertWorkflowTemplateToResource(ctx, &stateData, template)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
}

func (r *workflowTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data workflowTemplateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Delete workflow template")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)

	id, err := convert.StringToInt32(data.Id)
	if err != nil {
		resourceErr := errors.NewError("emma_workflow_template", "Delete").
			WithID(data.Id.ValueString()).
			WithMessage("Invalid workflow template ID format").
			WithCause(err).
			Build()
		resp.Diagnostics.AddError("Validation Error", resourceErr.Error())
		return
	}

	response, err := r.apiClient.WorkflowsAPI.DeleteWorkflowTemplate(auth, id).Execute()

	if err != nil {
		statusCode := 0
		apiError := ""
		if response != nil {
			statusCode = response.StatusCode
			apiError = tools.ExtractErrorMessage(response)
		}

		resourceErr := errors.NewError("emma_workflow_template", "Delete").
			WithID(data.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(apiError).
			WithMessage(errors.MapHTTPError(statusCode, apiError)).
			Build()

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}
}

func (r *workflowTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "Import workflow template")

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	r.Read(ctx, resource.ReadRequest{State: resp.State, Private: resp.Private},
		&resource.ReadResponse{State: resp.State, Private: resp.Private, Diagnostics: resp.Diagnostics})
}

func convertToWorkflowTemplateCreateRequest(ctx context.Context, data workflowTemplateResourceModel) (emmaSdk.WorkflowTemplateCreate, diag.Diagnostics) {
	var diags diag.Diagnostics
	req := emmaSdk.WorkflowTemplateCreate{
		Name:         data.Name.ValueString(),
		ContentType:  data.ContentType.ValueString(),
		Content:      data.Content.ValueString(),
		Status:       data.Status.ValueString(),
		ResourceType: data.ResourceType.ValueString(),
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		desc := data.Description.ValueString()
		req.Description = &desc
	}

	// Server rejects null lists ("must not be null") — default to empty slices.
	req.ResourceParams = make([]emmaSdk.WorkflowTemplateCreateResourceParamsInner, 0)
	if !data.ResourceParams.IsNull() && !data.ResourceParams.IsUnknown() {
		var paramModels []workflowTemplateResourceParamModel
		diags.Append(data.ResourceParams.ElementsAs(ctx, &paramModels, false)...)
		if diags.HasError() {
			return req, diags
		}
		for _, p := range paramModels {
			inner := emmaSdk.WorkflowTemplateCreateResourceParamsInner{}
			if !p.Name.IsNull() && !p.Name.IsUnknown() {
				n := p.Name.ValueString()
				inner.Name = &n
			}
			if !p.Value.IsNull() && !p.Value.IsUnknown() {
				v := p.Value.ValueString()
				inner.Value = &v
			}
			req.ResourceParams = append(req.ResourceParams, inner)
		}
	}

	req.Tags = make([]emmaSdk.WorkflowTemplateCreateTagsInner, 0)
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		var tagModels []workflowTemplateTagModel
		diags.Append(data.Tags.ElementsAs(ctx, &tagModels, false)...)
		if diags.HasError() {
			return req, diags
		}
		for _, t := range tagModels {
			inner := emmaSdk.WorkflowTemplateCreateTagsInner{}
			if !t.TagId.IsNull() && !t.TagId.IsUnknown() {
				tid := t.TagId.ValueString()
				inner.TagId = &tid
			}
			if !t.Key.IsNull() && !t.Key.IsUnknown() {
				k := t.Key.ValueString()
				inner.Key = &k
			}
			if !t.Value.IsNull() && !t.Value.IsUnknown() {
				v := t.Value.ValueString()
				inner.Value = &v
			}
			req.Tags = append(req.Tags, inner)
		}
	}

	req.ContentParams = make([]emmaSdk.WorkflowTemplateCreateContentParamsInner, 0)
	if !data.ContentParams.IsNull() && !data.ContentParams.IsUnknown() {
		var cpModels []workflowTemplateContentParamModel
		diags.Append(data.ContentParams.ElementsAs(ctx, &cpModels, false)...)
		if diags.HasError() {
			return req, diags
		}
		for _, cp := range cpModels {
			inner := emmaSdk.WorkflowTemplateCreateContentParamsInner{}
			if !cp.Name.IsNull() && !cp.Name.IsUnknown() {
				n := cp.Name.ValueString()
				inner.Name = &n
			}
			if !cp.DefaultValue.IsNull() && !cp.DefaultValue.IsUnknown() {
				dv := cp.DefaultValue.ValueString()
				inner.DefaultValue = &dv
			}
			if !cp.Mandatory.IsNull() && !cp.Mandatory.IsUnknown() {
				m := cp.Mandatory.ValueBool()
				inner.Mandatory = &m
			}
			req.ContentParams = append(req.ContentParams, inner)
		}
	}

	return req, diags
}

func convertWorkflowTemplateToResource(ctx context.Context, data *workflowTemplateResourceModel, template *emmaSdk.WorkflowTemplate) diag.Diagnostics {
	var diags diag.Diagnostics

	id := template.Id
	data.Id = convert.Int32ToString(&id)
	data.Name = types.StringValue(template.Name)
	data.ContentType = types.StringValue(template.ContentType)
	data.Content = types.StringValue(template.Content)
	data.Status = types.StringValue(template.Status)
	data.ResourceType = types.StringValue(template.ResourceType)
	data.CreatedAt = types.StringValue(template.CreatedAt)
	data.CreatedByName = types.StringValue(template.CreatedByName)
	createdById := template.CreatedById
	data.CreatedById = convert.Int32ToString(&createdById)
	data.ModifiedAt = types.StringValue(template.ModifiedAt)
	data.ModifiedByName = types.StringValue(template.ModifiedByName)
	modifiedById := template.ModifiedById
	data.ModifiedById = convert.Int32ToString(&modifiedById)
	data.IsDeleted = types.BoolValue(template.IsDeleted)

	if template.Description != nil {
		data.Description = types.StringValue(*template.Description)
	} else {
		data.Description = types.StringNull()
	}

	if template.IsContentValid != nil {
		data.IsContentValid = types.BoolValue(*template.IsContentValid)
	} else {
		data.IsContentValid = types.BoolNull()
	}

	// resource_params
	resourceParamType := types.ObjectType{AttrTypes: workflowTemplateResourceParamAttrTypes()}
	if len(template.ResourceParams) > 0 {
		paramModels := make([]workflowTemplateResourceParamModel, 0, len(template.ResourceParams))
		for _, p := range template.ResourceParams {
			m := workflowTemplateResourceParamModel{
				Name:  convert.StringPointerToString(p.Name),
				Value: convert.StringPointerToString(p.Value),
			}
			paramModels = append(paramModels, m)
		}
		listVal, listDiags := types.ListValueFrom(ctx, resourceParamType, paramModels)
		diags.Append(listDiags...)
		data.ResourceParams = listVal
	} else {
		data.ResourceParams = types.ListValueMust(resourceParamType, []attr.Value{})
	}

	// tags
	tagType := types.ObjectType{AttrTypes: workflowTemplateTagAttrTypes()}
	if len(template.Tags) > 0 {
		tagModels := make([]workflowTemplateTagModel, 0, len(template.Tags))
		for _, t := range template.Tags {
			m := workflowTemplateTagModel{
				TagId: convert.StringPointerToString(t.TagId),
				Key:   convert.StringPointerToString(t.Key),
				Value: convert.StringPointerToString(t.Value),
			}
			tagModels = append(tagModels, m)
		}
		listVal, listDiags := types.ListValueFrom(ctx, tagType, tagModels)
		diags.Append(listDiags...)
		data.Tags = listVal
	} else {
		data.Tags = types.ListValueMust(tagType, []attr.Value{})
	}

	// content_params
	contentParamType := types.ObjectType{AttrTypes: workflowTemplateContentParamAttrTypes()}
	if len(template.ContentParams) > 0 {
		cpModels := make([]workflowTemplateContentParamModel, 0, len(template.ContentParams))
		for _, cp := range template.ContentParams {
			m := workflowTemplateContentParamModel{
				Name:         convert.StringPointerToString(cp.Name),
				DefaultValue: convert.StringPointerToString(cp.DefaultValue),
			}
			if cp.Mandatory != nil {
				m.Mandatory = types.BoolValue(*cp.Mandatory)
			} else {
				m.Mandatory = types.BoolNull()
			}
			cpModels = append(cpModels, m)
		}
		listVal, listDiags := types.ListValueFrom(ctx, contentParamType, cpModels)
		diags.Append(listDiags...)
		data.ContentParams = listVal
	} else {
		data.ContentParams = types.ListValueMust(contentParamType, []attr.Value{})
	}

	return diags
}
