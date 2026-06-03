package emma

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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
		Description: "This resource creates a security group.\n\n" +
			"A security group refers to a set of rules that determine what network traffic is allowed to enter or leave " +
			"a particular compute instance. It acts as a virtual firewall, controlling inbound and outbound traffic " +
			"based on predefined rules.\n\n" +
			"Security groups operate based on predefined rules that allow traffic based on specified criteria, such as " +
			"source IP address, destination IP address, port number, and protocol.\n\n" +
			"When creating a security group, provide its name and a set of inbound and outbound rules. You can only " +
			"define rules that allow traffic, not deny it. All traffic is denied except for explicitly allowed traffic.\n\n" +
			"Security groups control TCP, SCTP, GRE, ESP, AH, UDP, and ICMP protocols, or all the selected protocols at once.\n\n" +
			"After creating a security group, a set of default rules is added to the security group. These rules are " +
			"immutable, and you can't edit or delete them.\n\n" +
			"All traffic in the selected protocol is allowed if the IP range in a rule is set to `0.0.0.0/0`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "ID of the security group",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Description: "Security group name",
				Computed:    false,
				Required:    true,
				Optional:    false,
				Validators:  []validator.String{emma.NotEmptyString{}, emma.SecurityGroupName{}},
			},
			"synchronization_status": schema.StringAttribute{
				Description: "Synchronization status of the security group. When you make changes in the rules the changes are propagated to the respective provider’s security groups. While this is happening the security groups have the status Synchronizing. After it is done the status changes to Synchronized. When another VM is added to the security group it will not be synchronized at first with the other VMs, therefore the status will be Desynchronized.",
				Computed:    true,
			},
			"recomposing_status": schema.StringAttribute{
				Description: "Recomposing status of the security group. When a new Virtual machine is added to the Security group it starts a synchronization process. During this process the Security group will have a Recomposing status.",
				Computed:    true,
			},
			"last_modification_error_description": schema.StringAttribute{
				Description: "Text of the error when the Security group was last edited",
				Computed:    true,
				Required:    false,
				Optional:    true,
			},
			"rules": schema.ListNestedAttribute{
				Computed:    false,
				Required:    true,
				Optional:    false,
				Validators:  []validator.List{emma.NotEmptyList{}},
				Description: "List of the inbound and outbound rules in the Security group",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"direction": schema.StringAttribute{
							Description: "Direction of the network traffic, available values: INBOUND or OUTBOUND",
							Computed:    false,
							Required:    true,
							Optional:    false,
							Validators:  []validator.String{emma.Direction{}},
						},
						"protocol": schema.StringAttribute{
							Description: "Network protocol, available values: all, TCP, SCTP, GRE, ESP, AH, UDP or ICMP",
							Computed:    false,
							Required:    true,
							Optional:    false,
							Validators:  []validator.String{emma.Protocol{}},
						},
						"ports": schema.StringAttribute{
							Description: "Allowed port or port range, available values: port number (8080), port range (1000-1005), all ports (all)",
							Computed:    false,
							Required:    true,
							Optional:    false,
							Validators:  []validator.String{emma.PortRange{}},
						},
						"ip_range": schema.StringAttribute{
							Description: "Allowed IP or IP range, available values: ip (8.8.8.8), ip range (8.8.8.8\\32), all ip addresses (0.0.0.0\\0)",
							Computed:    false,
							Required:    true,
							Optional:    false,
							Validators:  []validator.String{emma.IpRange{}},
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
		statusCode := 0
		if response != nil {
			statusCode = response.StatusCode
		}
		resourceErr := errors.NewError("emma_security_group", "Create").
			WithStatusCode(statusCode).
			WithAPIError(tools.ExtractErrorMessage(response)).
			WithMessage(errors.MapHTTPError(statusCode, tools.ExtractErrorMessage(response))).
			WithCause(err).
			Build()
		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
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
		statusCode := 0
		if response != nil {
			statusCode = response.StatusCode
		}
		
		// Handle 404 by removing from state
		if statusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		
		resourceErr := errors.NewError("emma_security_group", "Read").
			WithID(data.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(tools.ExtractErrorMessage(response)).
			WithMessage(errors.MapHTTPError(statusCode, tools.ExtractErrorMessage(response))).
			WithCause(err).
			Build()
		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
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
	
	securityGroupID := tools.StringToInt32(stateData.Id.ValueString())

	tflog.Debug(ctx, "Waiting for security group to reach RECOMPOSED state before update", map[string]interface{}{
		"security_group_id": stateData.Id.ValueString(),
	})

	// Wait for security group to reach RECOMPOSED state before update
	stateManager := state.NewStateTransitionManager(state.StateTransitionConfig{
		ResourceType: "security_group",
		ResourceID:   stateData.Id.ValueString(),
		StatusChecker: func(ctx context.Context) (string, error) {
			sg, _, err := r.apiClient.SecurityGroupsAPI.GetSecurityGroup(auth, securityGroupID).Execute()
			if err != nil {
				return "", err
			}
			if sg.RecomposingStatus == nil {
				return "", fmt.Errorf("security group recomposing status is nil")
			}
			return *sg.RecomposingStatus, nil
		},
		TargetStates:       state.SecurityGroupStableStates,
		TransitionalStates: state.SecurityGroupTransitionalStates,
		FailureStates:      state.SecurityGroupFailureStates,
		Timeout:            async.DefaultTimeout,
		PollInterval:       async.DefaultPollInterval,
	})

	if err := stateManager.WaitForStableState(auth); err != nil {
		resp.Diagnostics.AddError("State Transition Error",
			fmt.Sprintf("Security group did not reach RECOMPOSED state before update: %s", err.Error()))
		return
	}

	tflog.Info(ctx, "Security group reached RECOMPOSED state, proceeding with update", map[string]interface{}{
		"security_group_id": stateData.Id.ValueString(),
	})

	// Get current security group to preserve default rules
	securityGroup, response, err := r.apiClient.SecurityGroupsAPI.GetSecurityGroup(auth, securityGroupID).Execute()

	if err != nil {
		statusCode := 0
		if response != nil {
			statusCode = response.StatusCode
		}
		resourceErr := errors.NewError("emma_security_group", "Update").
			WithID(stateData.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(tools.ExtractErrorMessage(response)).
			WithMessage(errors.MapHTTPError(statusCode, tools.ExtractErrorMessage(response))).
			WithCause(err).
			Build()
		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
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

	// Perform update with retry on state conflicts
	retryConfig := retry.StateConflictRetryConfig()
	var lastResponse *http.Response
	var lastAPIError string
	var updatedSecurityGroup *emmaSdk.SecurityGroup

	err = retry.Retry(auth, retryConfig, func() error {
		sgResult, response, err := r.apiClient.SecurityGroupsAPI.SecurityGroupUpdate(auth, securityGroupID).SecurityGroupRequest(securityGroupRequest).Execute()

		lastResponse = response
		if err != nil {
			lastAPIError = tools.ExtractErrorMessage(response)
			statusCode := 0
			if response != nil {
				statusCode = response.StatusCode
			}

			// Check if this is a state conflict error that should be retried
			if retry.IsStateConflictError(err, statusCode, lastAPIError) {
				tflog.Warn(ctx, "Security group update failed due to state conflict, will retry", map[string]interface{}{
					"security_group_id": stateData.Id.ValueString(),
					"status_code":       statusCode,
					"error":             lastAPIError,
				})
				return err
			}

			// Non-retryable error
			return fmt.Errorf("non-retryable error: %w", err)
		}

		updatedSecurityGroup = sgResult
		return nil
	})

	if err != nil {
		resourceErr := errors.NewError("emma_security_group", "Update").
			WithID(stateData.Id.ValueString()).
			WithAPIError(lastAPIError).
			WithCause(err).
			Build()

		if lastResponse != nil {
			resourceErr.StatusCode = lastResponse.StatusCode
			resourceErr.Message = errors.MapHTTPError(lastResponse.StatusCode, lastAPIError)
		} else {
			resourceErr.Message = "Unable to update security group"
		}

		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	tflog.Info(ctx, "Security group update initiated, waiting for recomposition to complete", map[string]interface{}{
		"security_group_id": stateData.Id.ValueString(),
	})

	// Wait for security group to reach RECOMPOSED state after update
	if err := stateManager.WaitForStableState(auth); err != nil {
		tflog.Warn(ctx, "Security group did not reach RECOMPOSED state after update within timeout", map[string]interface{}{
			"security_group_id": stateData.Id.ValueString(),
			"error":             err.Error(),
		})
		// Don't fail the operation, just log the warning
		// The security group update was successful, it's just taking longer to recompose
	} else {
		tflog.Info(ctx, "Security group recomposition completed successfully", map[string]interface{}{
			"security_group_id": stateData.Id.ValueString(),
		})
	}

	ConvertSecurityGroupResponseToResource(ctx, &planData, &stateData, updatedSecurityGroup, &resp.Diagnostics)

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
	
	// Wait for security group to be synchronized and recomposed
	syncPoller := async.NewPoller(async.PollerConfig{
		Timeout:      3 * time.Minute, // 36 * 5 seconds = 3 minutes
		PollInterval: 5 * time.Second,
		StatusChecker: func(ctx context.Context) (string, error) {
			securityGroup, response, err := r.apiClient.SecurityGroupsAPI.GetSecurityGroup(auth, tools.StringToInt32(data.Id.ValueString())).Execute()
			if err != nil {
				statusCode := 0
				if response != nil {
					statusCode = response.StatusCode
				}
				return "", fmt.Errorf("failed to get security group status: %s", errors.MapHTTPError(statusCode, tools.ExtractErrorMessage(response)))
			}
			
			// Check if synchronized and recomposed
			if *securityGroup.SynchronizationStatus == "SYNCHRONIZED" && *securityGroup.RecomposingStatus == "RECOMPOSED" {
				// Also check if there are no instances attached
				securityGroupInstances, response, err := r.apiClient.SecurityGroupsAPI.SecurityGroupInstances(auth, tools.StringToInt32(data.Id.ValueString())).Execute()
				if err != nil {
					statusCode := 0
					if response != nil {
						statusCode = response.StatusCode
					}
					return "", fmt.Errorf("failed to get security group instances: %s", errors.MapHTTPError(statusCode, tools.ExtractErrorMessage(response)))
				}
				
				if len(securityGroupInstances) == 0 {
					return "READY_FOR_DELETE", nil
				}
				return "HAS_INSTANCES", nil
			}
			
			return fmt.Sprintf("%s/%s", *securityGroup.SynchronizationStatus, *securityGroup.RecomposingStatus), nil
		},
		TargetStates:  []string{"READY_FOR_DELETE"},
		FailureStates: []string{},
	})
	
	if err := syncPoller.Poll(ctx); err != nil {
		resourceErr := errors.NewError("emma_security_group", "Delete").
			WithID(data.Id.ValueString()).
			WithMessage(fmt.Sprintf("Timeout waiting for security group to be ready for deletion: %s", err.Error())).
			WithCause(err).
			Build()
		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
		return
	}

	_, response, err := r.apiClient.SecurityGroupsAPI.SecurityGroupDelete(auth, tools.StringToInt32(data.Id.ValueString())).Execute()
	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	if err != nil {
		statusCode := 0
		if response != nil {
			statusCode = response.StatusCode
		}
		resourceErr := errors.NewError("emma_security_group", "Delete").
			WithID(data.Id.ValueString()).
			WithStatusCode(statusCode).
			WithAPIError(tools.ExtractErrorMessage(response)).
			WithMessage(errors.MapHTTPError(statusCode, tools.ExtractErrorMessage(response))).
			WithCause(err).
			Build()
		resp.Diagnostics.AddError("Client Error", resourceErr.Error())
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
		rulesListValue, listDiags := stateData.Rules.ToListValue(ctx)
		diags.Append(listDiags...)
		rulesListValue.ElementsAs(ctx, &rules, false)

		// Build an ordered index: key -> position in the state slice.
		// Using the state slice length (not map length) so indices are always valid.
		ruleOrderMap := make(map[string]int, len(rules))
		for idx, rule := range rules {
			key := rule.Direction.ValueString() + rule.Protocol.ValueString() + rule.Ports.ValueString() + rule.IpRange.ValueString()
			ruleOrderMap[key] = idx
		}

		// slots mirrors the state slice; filled[i]=true means the API returned that rule.
		slots := make([]*securityGroupResourceRuleModel, len(rules))
		var appended []securityGroupResourceRuleModel

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
				slots[idx] = &securityGroupRuleModel
			} else if idx1, ok1 := ruleOrderMap[*securityGroupRule.Direction+*securityGroupRule.Protocol+*securityGroupRule.Ports+stripSubnetMask(*securityGroupRule.IpRange)]; ok1 {
				securityGroupRuleModel.IpRange = types.StringValue(stripSubnetMask(securityGroupRuleModel.IpRange.ValueString()))
				slots[idx1] = &securityGroupRuleModel
			} else {
				appended = append(appended, securityGroupRuleModel)
			}
		}

		// Collect matched state-ordered rules, skipping slots not returned by API (deleted server-side).
		securityGroupRuleModels := make([]securityGroupResourceRuleModel, 0, len(rules)+len(appended))
		for _, slot := range slots {
			if slot != nil {
				securityGroupRuleModels = append(securityGroupRuleModels, *slot)
			}
		}
		securityGroupRuleModels = append(securityGroupRuleModels, appended...)

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
