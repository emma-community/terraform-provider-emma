package emma

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"regexp"
	"slices"
)

type KubernetesResourceName struct {
	FieldName string
}

func (k KubernetesResourceName) Description(ctx context.Context) string {
	return "The name must be less than 64 characters, start with a lowercase letter, end with a lowercase alphanumeric, and use only lowercase alphanumerics and hyphens in between."
}

func (k KubernetesResourceName) MarkdownDescription(ctx context.Context) string {
	return "The name must be less than 64 characters, start with a lowercase letter, end with a lowercase alphanumeric, and use only lowercase alphanumerics and hyphens in between."
}

func (k KubernetesResourceName) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		resp.Diagnostics.AddError("Validation Error", fmt.Sprintf("%s is required", req.Path.String()))
		return
	}
	matches, _ := regexp.MatchString(`^([a-z](?:[0-9a-z-]{0,62}[0-9a-z]))$`, req.ConfigValue.ValueString())
	if !matches {
		resp.Diagnostics.AddError("Validation Error", fmt.Sprintf("Action rejected: The %s must be less than 64 characters, start with a lowercase letter, end with a lowercase alphanumeric, and use only lowercase alphanumerics and hyphens in between.", k.FieldName))
	}
}

type KubernetesResourceDomainName struct{}

func (k KubernetesResourceDomainName) Description(ctx context.Context) string {
	return "The domain name must be a valid domain name format."
}

func (k KubernetesResourceDomainName) MarkdownDescription(ctx context.Context) string {
	return "The domain name must be a valid domain name format."
}

func (k KubernetesResourceDomainName) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	matches, _ := regexp.MatchString(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`, req.ConfigValue.ValueString())
	if !matches {
		resp.Diagnostics.AddError("Validation Error", "Action rejected: Invalid domain name.")
	}
}

type UniqueField struct {
	FieldName string
}

func (v UniqueField) Description(ctx context.Context) string {
	return fmt.Sprintf("Ensures that all %s are unique.", v.FieldName)
}

func (v UniqueField) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("Ensures that all %s are unique.", v.FieldName)
}

func (v UniqueField) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var listItems []basetypes.ObjectValue
	diag := req.ConfigValue.ElementsAs(ctx, &listItems, false)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	names := make(map[string]struct{})
	for idx, node := range listItems {
		nameAttr, exists := node.Attributes()[v.FieldName]
		if !exists || nameAttr.IsNull() || nameAttr.IsUnknown() {
			resp.Diagnostics.AddError("Validation Error", fmt.Sprintf("%s[%d].%s is required", req.Path.String(), idx, v.FieldName))
			return
		}

		name := nameAttr.(types.String).ValueString()
		if _, exists := names[name]; exists {
			resp.Diagnostics.AddError("Validation Error", fmt.Sprintf("%s[%d].%s must be unique", req.Path.String(), idx, v.FieldName))
			return
		}
		names[name] = struct{}{}
	}
}

type NodeGroupPriceLimit struct{}

func (n NodeGroupPriceLimit) Description(ctx context.Context) string {
	return "The value must be between 0 and 5000 inclusive."
}

func (n NodeGroupPriceLimit) MarkdownDescription(ctx context.Context) string {
	return "The value must be between 0 and 5000 inclusive."
}

func (n NodeGroupPriceLimit) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	price := req.ConfigValue.ValueFloat64()
	if price < 0 || price > 5000 {
		resp.Diagnostics.AddError("Validation Error", "Action rejected: The allowed maximum price range for a single group is €0 - €5000.")
	}
}

type SpotPercent struct{}

func (s SpotPercent) Description(ctx context.Context) string {
	return "The value must be between 0 and 100 inclusive."
}

func (s SpotPercent) MarkdownDescription(ctx context.Context) string {
	return "The value must be between 0 and 100 inclusive."
}

func (s SpotPercent) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	spotPercent := req.ConfigValue.ValueInt64()
	if spotPercent < 0 || spotPercent > 100 {
		resp.Diagnostics.AddError("Validation Error", "Action rejected: spot_percent value must be between 0 and 100.")
	}
}

type SpotMarkup struct{}

func (s SpotMarkup) Description(ctx context.Context) string {
	return "The value must be between 0 and 100 inclusive."
}

func (s SpotMarkup) MarkdownDescription(ctx context.Context) string {
	return "The value must be between 0 and 100 inclusive."
}

func (s SpotMarkup) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	spotMarkup := req.ConfigValue.ValueFloat64()
	if spotMarkup < 0 || spotMarkup > 100 {
		resp.Diagnostics.AddError("Validation Error", "Action rejected: spot_markup value must be between 0 and 100.")
	}
}

type VCPUType struct{}

func (v VCPUType) Description(ctx context.Context) string {
	return "The value must be one of: shared, standard, hpc."
}

func (v VCPUType) MarkdownDescription(ctx context.Context) string {
	return "The value must be one of: shared, standard, hpc."
}

func (v VCPUType) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	value := req.ConfigValue.ValueString()
	validValues := []string{"shared", "standard", "hpc"}
	if !slices.Contains(validValues, value) {
		resp.Diagnostics.AddError("Validation Error", "Action rejected: configuration_priority.vcpu_type value must be one of: shared, standard, hpc.")
	}
}

type Priority struct{}

func (p Priority) Description(ctx context.Context) string {
	return "The value must be one of: low, med, high."
}

func (p Priority) MarkdownDescription(ctx context.Context) string {
	return "The value must be one of: low, med, high."
}

func (p Priority) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	value := req.ConfigValue.ValueString()
	validValues := []string{"low", "med", "high"}
	if !slices.Contains(validValues, value) {
		resp.Diagnostics.AddError("Validation Error", "Action rejected: configuration_priority.priority value must be one of: low, med, high.")
	}
}

type AutoscalingConfigValidator struct {
}

func (v AutoscalingConfigValidator) Description(ctx context.Context) string {
	return "autoscaling_configs should be valid."
}

func (v AutoscalingConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "autoscaling_configs should be valid."
}

func (v AutoscalingConfigValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	obj, diag := req.ConfigValue.ToObjectValue(ctx)
	if diag.HasError() {
		resp.Diagnostics.Append(diag.Errors()...)
		return
	}

	attrMap := obj.Attributes()
	minNodes, isMinNodesPresent := getAttributeInt64Value(ctx, attrMap, "minimum_nodes", resp)
	maxNodes, isMaxNodesPresent := getAttributeInt64Value(ctx, attrMap, "maximum_nodes", resp)
	targetNodes, isTargetNodesPresent := getAttributeInt64Value(ctx, attrMap, "target_nodes", resp)
	minVCPUs, isMinVCPUsPresent := getAttributeInt64Value(ctx, attrMap, "minimum_vcpus", resp)
	maxVCPUs, isMaxVCPUsPresent := getAttributeInt64Value(ctx, attrMap, "maximum_vcpus", resp)
	targetVCPUs, isTargetVCPUsPresent := getAttributeInt64Value(ctx, attrMap, "target_vcpus", resp)

	if resp.Diagnostics.HasError() {
		return
	}

	isAllNodesPresent := isMinNodesPresent && isMaxNodesPresent && isTargetNodesPresent
	isAllVCPUsPresent := isMinVCPUsPresent && isMaxVCPUsPresent && isTargetVCPUsPresent

	if (isAllNodesPresent && (isMinVCPUsPresent || isMaxVCPUsPresent || isTargetVCPUsPresent)) || (isAllVCPUsPresent && (isMinNodesPresent || isMaxNodesPresent || isTargetNodesPresent)) || (!isAllNodesPresent && !isAllVCPUsPresent) {
		resp.Diagnostics.AddError(
			"Validation Error",
			"Action rejected: contradicting field values: minimum_nodes, maximum_nodes, target_nodes, minimum_vcpus, maximum_vcpus, target_vcpus.",
		)
		return
	}

	if isAllNodesPresent {
		if minNodes <= 0 {
			resp.Diagnostics.AddError("Validation Error", "Action rejected: The value of minimum_nodes must be greater than 0")
		}
		if maxNodes <= 0 || maxNodes < minNodes {
			resp.Diagnostics.AddError("Validation Error", "Action rejected: The value of maximum_nodes must be greater than 0 and greater than or equal to the value of minimum_nodes")
		}
		if targetNodes < minNodes || targetNodes > maxNodes {
			resp.Diagnostics.AddError("Validation Error", "Action rejected: The value of target_nodes must be between minimum_nodes and maximum_nodes inclusive.")
		}
		return
	}

	if isAllVCPUsPresent {
		if minVCPUs <= 0 {
			resp.Diagnostics.AddError("Validation Error", "Action rejected: The value of minimum_vcpus must be greater than 0")
		}
		if maxVCPUs <= 0 || maxVCPUs < minVCPUs {
			resp.Diagnostics.AddError("Validation Error", "Action rejected: The value of maximum_vcpus must be greater than 0 and greater than or equal to the value of minimum_vcpus")
		}
		if targetVCPUs < minVCPUs || targetVCPUs > maxVCPUs {
			resp.Diagnostics.AddError("Validation Error", "Action rejected: The value of target_vcpus must be between minimum_vcpus and maximum_vcpus inclusive.")
		}
		return
	}
}

func getAttributeInt64Value(ctx context.Context, attrMap map[string]attr.Value, fieldName string, resp *validator.ObjectResponse) (int64, bool) {
	attribute, isPresent := attrMap[fieldName]
	if !isPresent || attribute.IsNull() || attribute.IsUnknown() {
		return 0, false
	}

	intValue, ok := attribute.(types.Int64)
	if !ok {
		resp.Diagnostics.AddError(
			"Validation Error",
			fmt.Sprintf("Failed to convert value for %s to int64: %v", fieldName, attribute),
		)
		return 0, false
	}

	return intValue.ValueInt64(), true
}
