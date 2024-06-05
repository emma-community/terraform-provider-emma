package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloudNetworkType_ValidateString_InvalidValue(t *testing.T) {
	v := CloudNetworkType{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("test")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test can contain multi-cloud, isolated or default", actualMsg)
	} else {
		assert.Fail(t, "Is valid cloud_network_type value: test")
	}
}

func TestCloudNetworkType_ValidateString_ValidValues(t *testing.T) {
	for _, validCloudNetworkTypeValue := range []string{"multi-cloud", "isolated", "default"} {
		v := CloudNetworkType{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(validCloudNetworkTypeValue)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.False(t, resp.Diagnostics.HasError(), "Is invalid cloud_network_type value: "+validCloudNetworkTypeValue)
	}
}

func TestVCpuType_ValidateString_InvalidValue(t *testing.T) {
	v := VCpuType{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("test")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test can contain shared, standard or hpc", actualMsg)
	} else {
		assert.Fail(t, "Is valid vcpu_type value: test")
	}
}

func TestVCpuType_ValidateString_ValidValues(t *testing.T) {
	for _, validVCpuTypeValue := range []string{"shared", "standard", "hpc"} {
		v := VCpuType{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(validVCpuTypeValue)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.False(t, resp.Diagnostics.HasError(), "Is invalid vcpu_type value: "+validVCpuTypeValue)
	}
}

func TestVolumeType_ValidateString_InvalidValue(t *testing.T) {
	v := VolumeType{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("test")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test can contain ssd or ssd-plus", actualMsg)
	} else {
		assert.Fail(t, "Is valid volume_type value: test")
	}
}

func TestVolumeType_ValidateString_ValidValues(t *testing.T) {
	for _, validVolumeTypeValue := range []string{"ssd", "ssd-plus"} {
		v := VolumeType{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(validVolumeTypeValue)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.False(t, resp.Diagnostics.HasError(), "Is invalid volume_type value: "+validVolumeTypeValue)
	}
}
