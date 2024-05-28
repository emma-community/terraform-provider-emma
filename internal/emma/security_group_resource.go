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
	"strings"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &securityGroupResource{}

func NewSecurityGroupResource() resource.Resource {
	return &securityGroupResource{}
}

// securityGroupResource defines the resource implementation.
type securityGroupResource struct {
	apiClient *emmaSdk.APIClient
	token     *emmaSdk.Token
}

// securityGroupResourceModel describes the resource data model.
type securityGroupResourceModel struct {
	Id                               types.String `tfsdk:"id"`
	Name                             types.String `tfsdk:"name"`
	SynchronizationStatus            types.String `tfsdk:"synchronization_status"`
	RecomposingStatus                types.String `tfsdk:"recomposing_status"`
	LastModificationErrorDescription types.String `tfsdk:"last_modification_error_description"`
	Rules                            types.List   `tfsdk:"rules"`
}

type securityGroupResourceRuleModel struct {
	Direction types.String `tfsdk:"direction"`
	Protocol  types.String `tfsdk:"protocol"`
	Ports     types.String `tfsdk:"ports"`
	IpRange   types.String `tfsdk:"ip_range"`
}

func (r *securityGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *securityGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityGroup resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "SecurityGroup id configurable attribute",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SecurityGroup name configurable attribute",
				Computed:            false,
				Required:            true,
				Optional:            false,
				Validators:          []validator.String{emma.NotEmptyString{}},
			},
			"synchronization_status": schema.StringAttribute{
				MarkdownDescription: "SecurityGroup synchronization_status configurable attribute",
				Computed:            true,
			},
			"recomposing_status": schema.StringAttribute{
				MarkdownDescription: "SecurityGroup recomposing_status configurable attribute",
				Computed:            true,
			},
			"last_modification_error_description": schema.StringAttribute{
				MarkdownDescription: "SecurityGroup last_modification_error_description configurable attribute",
				Computed:            true,
				Required:            false,
				Optional:            true,
			},
			"rules": schema.ListNestedAttribute{
				Computed:   false,
				Required:   true,
				Optional:   false,
				Validators: []validator.List{emma.NotEmptyList{}},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"direction": schema.StringAttribute{
							MarkdownDescription: "SecurityGroup rules direction configurable attribute",
							Computed:            false,
							Required:            true,
							Optional:            false,
							Validators:          []validator.String{emma.Direction{}},
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "SecurityGroup rules protocol configurable attribute",
							Computed:            false,
							Required:            true,
							Optional:            false,
							Validators:          []validator.String{emma.Protocol{}},
						},
						"ports": schema.StringAttribute{
							MarkdownDescription: "SecurityGroup rules ports configurable attribute",
							Computed:            false,
							Required:            true,
							Optional:            false,
							Validators:          []validator.String{emma.PortRange{}},
						},
						"ip_range": schema.StringAttribute{
							MarkdownDescription: "SecurityGroup rules ip_range configurable attribute",
							Computed:            false,
							Required:            true,
							Optional:            false,
							Validators:          []validator.String{emma.IpRange{}},
						},
					},
				},
			},
		},
	}
}

func (r *securityGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *securityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data securityGroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Create security group")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	var securityGroupRequest emmaSdk.SecurityGroupRequest
	ConvertToSecurityGroupRequest(ctx, data, &securityGroupRequest)
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	securityGroup, response, err := r.apiClient.SecurityGroupsAPI.SecurityGroupCreate(auth).SecurityGroupRequest(securityGroupRequest).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create security group, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	ConvertSecurityGroupResponseToResource(ctx, nil, &data, securityGroup, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *securityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data securityGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Read security group")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	securityGroup, response, err := r.apiClient.SecurityGroupsAPI.GetSecurityGroup(auth, tools.StringToInt32(data.Id.ValueString())).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read security group, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	ConvertSecurityGroupResponseToResource(ctx, nil, &data, securityGroup, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *securityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData securityGroupResourceModel
	var stateData securityGroupResourceModel

	// Read Terraform plan planData into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Update security group")

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client planData and make a call using it.
	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	securityGroup, response, err := r.apiClient.SecurityGroupsAPI.GetSecurityGroup(auth, tools.StringToInt32(stateData.Id.ValueString())).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read security group, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	defaultSecurityGroupRules := make([]emmaSdk.SecurityGroupRule, 0)
	for _, securityGroupRule := range securityGroup.Rules {
		if !*securityGroupRule.IsMutable {
			defaultSecurityGroupRules = append(defaultSecurityGroupRules, securityGroupRule)
		}
	}

	var securityGroupRequest emmaSdk.SecurityGroupRequest
	ConvertToSecurityGroupUpdateRequest(ctx, planData, &securityGroupRequest, defaultSecurityGroupRules)
	securityGroup, response, err = r.apiClient.SecurityGroupsAPI.SecurityGroupUpdate(auth, tools.StringToInt32(stateData.Id.ValueString())).SecurityGroupRequest(securityGroupRequest).Execute()

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update security group, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}

	ConvertSecurityGroupResponseToResource(ctx, &planData, &stateData, securityGroup, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save planData into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
}

func (r *securityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data securityGroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Delete security group")

	auth := context.WithValue(ctx, emmaSdk.ContextAccessToken, *r.token.AccessToken)
	i := 0
	for i < 60 {
		i++
		securityGroup, response, err := r.apiClient.SecurityGroupsAPI.GetSecurityGroup(auth, tools.StringToInt32(data.Id.ValueString())).Execute()
		if *securityGroup.SynchronizationStatus != "SYNCHRONIZED" || *securityGroup.RecomposingStatus != "RECOMPOSED" {
			time.Sleep(5 * time.Second)
			continue
		}

		securityGroupInstances, response, err := r.apiClient.SecurityGroupsAPI.SecurityGroupInstances(auth, tools.StringToInt32(data.Id.ValueString())).Execute()
		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to get security group instances, got error: %s, %s",
					tools.ExtractErrorMessage(response), err))
			return
		}

		if len(securityGroupInstances) != 0 {
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	_, response, err := r.apiClient.SecurityGroupsAPI.SecurityGroupDelete(auth, tools.StringToInt32(data.Id.ValueString())).Execute()
	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete security group, got error: %s",
				tools.ExtractErrorMessage(response)))
		return
	}
}

func (r *securityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "Import security group")

	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	r.Read(ctx, resource.ReadRequest{State: resp.State, Private: resp.Private},
		&resource.ReadResponse{State: resp.State, Private: resp.Private, Diagnostics: resp.Diagnostics})
}

func ConvertToSecurityGroupRequest(ctx context.Context, data securityGroupResourceModel, securityGroupRequest *emmaSdk.SecurityGroupRequest) {
	securityGroupRequest.Name = data.Name.ValueString()
	var rules []securityGroupResourceRuleModel
	rulesListValue, _ := data.Rules.ToListValue(ctx)
	rulesListValue.ElementsAs(ctx, &rules, false)
	var requestRules []emmaSdk.SecurityGroupRuleRequest
	for _, rule := range rules {
		requestRule := emmaSdk.SecurityGroupRuleRequest{
			Direction: rule.Direction.ValueString(),
			Protocol:  rule.Protocol.ValueString(),
			Ports:     rule.Ports.ValueString(),
			IpRange:   rule.IpRange.ValueString(),
		}
		requestRules = append(requestRules, requestRule)
	}
	securityGroupRequest.Rules = requestRules
}

func ConvertToSecurityGroupUpdateRequest(ctx context.Context, data securityGroupResourceModel,
	securityGroupRequest *emmaSdk.SecurityGroupRequest, defaultSecurityGroupRules []emmaSdk.SecurityGroupRule) {
	ConvertToSecurityGroupRequest(ctx, data, securityGroupRequest)
	defaultSecurityGroupRequestRules := make([]emmaSdk.SecurityGroupRuleRequest, 0)
	for _, defaultSecurityGroupRule := range defaultSecurityGroupRules {
		defaultSecurityGroupRequestRule := emmaSdk.NewSecurityGroupRuleRequest(*defaultSecurityGroupRule.Direction,
			*defaultSecurityGroupRule.Protocol, *defaultSecurityGroupRule.Ports, *defaultSecurityGroupRule.IpRange)
		defaultSecurityGroupRequestRules = append(defaultSecurityGroupRequestRules, *defaultSecurityGroupRequestRule)
	}
	securityGroupRequest.Rules = append(securityGroupRequest.Rules, defaultSecurityGroupRequestRules...)
}

func ConvertSecurityGroupResponseToResource(ctx context.Context, planData *securityGroupResourceModel,
	stateData *securityGroupResourceModel, securityGroupResponse *emmaSdk.SecurityGroup, diags *diag.Diagnostics) {

	stateData.Id = types.StringValue(strconv.Itoa(int(*securityGroupResponse.Id)))
	stateData.Name = types.StringValue(*securityGroupResponse.Name)
	stateData.SynchronizationStatus = types.StringValue(*securityGroupResponse.SynchronizationStatus)
	stateData.RecomposingStatus = types.StringValue(*securityGroupResponse.RecomposingStatus)
	if securityGroupResponse.LastModificationErrorDescription != nil {
		stateData.LastModificationErrorDescription = types.StringValue(*securityGroupResponse.LastModificationErrorDescription)
	} else {
		stateData.LastModificationErrorDescription = types.StringValue("")
	}
	if planData != nil {
		// since we have async security group update we store requested state
		stateData.Rules = planData.Rules
		stateData.Name = planData.Name
	} else if securityGroupResponse.Rules != nil {
		var rules []securityGroupResourceRuleModel
		rulesListValue, _ := stateData.Rules.ToListValue(ctx)
		rulesListValue.ElementsAs(ctx, &rules, false)
		ruleOrderMap := make(map[string]int)
		for idx, rule := range rules {
			ruleOrderMap[rule.Direction.ValueString()+rule.Protocol.ValueString()+rule.Ports.ValueString()+rule.IpRange.ValueString()] = idx
		}
		securityGroupRuleModels := make([]securityGroupResourceRuleModel, len(ruleOrderMap))
		for _, securityGroupRule := range securityGroupResponse.Rules {
			if securityGroupRule.IsMutable == nil || !*securityGroupRule.IsMutable {
				continue
			}
			securityGroupRuleModel := securityGroupResourceRuleModel{
				Direction: types.StringValue(*securityGroupRule.Direction),
				Protocol:  types.StringValue(*securityGroupRule.Protocol),
				Ports:     types.StringValue(*securityGroupRule.Ports),
				IpRange:   types.StringValue(*securityGroupRule.IpRange),
			}
			// to save same order as in configuration we have map, and we have 2 different checks with subnet mask and without
			if idx, ok := ruleOrderMap[*securityGroupRule.Direction+*securityGroupRule.Protocol+*securityGroupRule.Ports+*securityGroupRule.IpRange]; ok {
				securityGroupRuleModels[idx] = securityGroupRuleModel
			} else if idx1, ok1 := ruleOrderMap[*securityGroupRule.Direction+*securityGroupRule.Protocol+*securityGroupRule.Ports+stripSubnetMask(*securityGroupRule.IpRange)]; ok1 {
				securityGroupRuleModel.IpRange = types.StringValue(stripSubnetMask(securityGroupRuleModel.IpRange.ValueString()))
				securityGroupRuleModels[idx1] = securityGroupRuleModel
			} else {
				securityGroupRuleModels = append(securityGroupRuleModels, securityGroupRuleModel)
			}
		}
		rulesListValue, rulesDiagnostic := types.ListValueFrom(ctx,
			types.ObjectType{AttrTypes: securityGroupResourceRuleModel{}.attrTypes()}, securityGroupRuleModels)
		stateData.Rules = rulesListValue
		diags.Append(rulesDiagnostic...)
	}
}

func (o securityGroupResourceRuleModel) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"direction": types.StringType,
		"protocol":  types.StringType,
		"ports":     types.StringType,
		"ip_range":  types.StringType,
	}
}

func stripSubnetMask(ipRange string) string {
	if strings.Contains(ipRange, "/") {
		return strings.Split(ipRange, "/")[0]
	}
	return ipRange
}
