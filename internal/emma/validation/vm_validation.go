package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
)

type CloudNetworkType struct {
}

func (v CloudNetworkType) Description(ctx context.Context) string {
	return "cloud_network_type can contain multi-cloud, isolated or default"
}

func (v CloudNetworkType) MarkdownDescription(ctx context.Context) string {
	return "cloud_network_type can contain multi-cloud, isolated or default"
}

func (v CloudNetworkType) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if req.ConfigValue.ValueString() != "multi-cloud" && req.ConfigValue.ValueString() != "isolated" &&
		req.ConfigValue.ValueString() != "default" {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" can contain multi-cloud, isolated or default")
	}
}

type VCpuType struct {
}

func (v VCpuType) Description(ctx context.Context) string {
	return "vcpu_type can contain shared, standard or hpc"
}

func (v VCpuType) MarkdownDescription(ctx context.Context) string {
	return "vcpu_type can contain shared, standard or hpc"
}

func (v VCpuType) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if req.ConfigValue.ValueString() != "shared" && req.ConfigValue.ValueString() != "standard" &&
		req.ConfigValue.ValueString() != "hpc" {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" can contain shared, standard or hpc")
	}
}

type VolumeType struct {
}

func (v VolumeType) Description(ctx context.Context) string {
	return "volume_type can contain ssd or ssd-plus"
}

func (v VolumeType) MarkdownDescription(ctx context.Context) string {
	return "volume_type can contain ssd or ssd-plus"
}

func (v VolumeType) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if req.ConfigValue.ValueString() != "ssd" && req.ConfigValue.ValueString() != "ssd-plus" {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" can contain ssd or ssd-plus")
	}
}

type VmName struct {
}

func (v VmName) Description(ctx context.Context) string {
	return "name must be less than 63 characters, start with a lowercase letter, end with a lowercase alphanumeric, and use only lowercase alphanumeric and hyphens in-between"
}

func (v VmName) MarkdownDescription(ctx context.Context) string {
	return "name must be less than 63 characters, start with a lowercase letter, end with a lowercase alphanumeric, and use only lowercase alphanumeric and hyphens in-between"
}

func (v VmName) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	matches, _ := regexp.MatchString(`^([a-z](?:[0-9a-z-]{0,61}[0-9a-z]))$`, req.ConfigValue.ValueString())
	if !matches {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" must be less than 63 characters, start with a lowercase letter, end with a lowercase alphanumeric, and use only lowercase alphanumeric and hyphens in-between")
	}
}
