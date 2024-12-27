package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKubernetesResourceName_ValidateString(t *testing.T) {
	v := KubernetesResourceName{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("valid-name")
	v.ValidateString(context.Background(), req, &resp)
	assert.False(t, resp.Diagnostics.HasError())

	req.ConfigValue = types.StringValue("Invalid-Name")
	v.ValidateString(context.Background(), req, &resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestKubernetesResourceDomainName_ValidateString(t *testing.T) {
	v := KubernetesResourceDomainName{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("valid-sub.DoMA1N.com")
	v.ValidateString(context.Background(), req, &resp)
	assert.False(t, resp.Diagnostics.HasError())

	invalidDomains := []string{
		"invalid_domain.com",
		"invalid.domain.c",
		"invalid.net/123",
	}

	for _, invalidDomain := range invalidDomains {
		var resp validator.StringResponse
		req.ConfigValue = types.StringValue(invalidDomain)
		v.ValidateString(context.Background(), req, &resp)
		assert.True(t, resp.Diagnostics.HasError())

		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "Action rejected: Invalid domain name.", actualMsg)
	}
}

func TestUniqueField_ValidateList(t *testing.T) {
	v := UniqueField{FieldName: "name"}
	var resp validator.ListResponse
	var req validator.ListRequest

	attrTypes := map[string]attr.Type{"name": types.StringType}

	// Valid case: unique names
	req.ConfigValue = types.ListValueMust(
		types.ObjectType{AttrTypes: attrTypes},
		[]attr.Value{
			types.ObjectValueMust(attrTypes, map[string]attr.Value{"name": types.StringValue("node1")}),
			types.ObjectValueMust(attrTypes, map[string]attr.Value{"name": types.StringValue("node2")}),
		},
	)

	v.ValidateList(context.Background(), req, &resp)
	assert.False(t, resp.Diagnostics.HasError())

	// Invalid case: duplicate names
	resp = validator.ListResponse{}
	req.ConfigValue = types.ListValueMust(
		types.ObjectType{AttrTypes: attrTypes},
		[]attr.Value{
			types.ObjectValueMust(attrTypes, map[string]attr.Value{"name": types.StringValue("node1")}),
			types.ObjectValueMust(attrTypes, map[string]attr.Value{"name": types.StringValue("node1")}),
		},
	)
	v.ValidateList(context.Background(), req, &resp)
	assert.True(t, resp.Diagnostics.HasError())
	assert.Equal(t, "[1].name must be unique", resp.Diagnostics.Errors()[0].Detail())

	resp = validator.ListResponse{}
	req.ConfigValue = types.ListValueMust(
		types.ObjectType{AttrTypes: attrTypes},
		[]attr.Value{
			types.ObjectValueMust(attrTypes, map[string]attr.Value{"name": types.StringValue("node1")}),
			types.ObjectValueMust(attrTypes, map[string]attr.Value{"name": types.StringNull()}),
		},
	)
	v.ValidateList(context.Background(), req, &resp)
	assert.True(t, resp.Diagnostics.HasError())
	assert.Equal(t, "[1].name is required", resp.Diagnostics.Errors()[0].Detail())
}

func TestNodeGroupPriceLimit_ValidateFloat64(t *testing.T) {
	v := NodeGroupPriceLimit{}
	var resp validator.Float64Response
	var req validator.Float64Request

	req.ConfigValue = types.Float64Value(1000)
	v.ValidateFloat64(context.Background(), req, &resp)
	assert.False(t, resp.Diagnostics.HasError())

	invalidPrices := []float64{
		-1,
		5001,
	}
	for _, invalidPrice := range invalidPrices {
		expectedErrorMsg := "Action rejected: The allowed maximum price range for a single group is €0 - €5000."

		var resp validator.Float64Response
		req.ConfigValue = types.Float64Value(invalidPrice)
		v.ValidateFloat64(context.Background(), req, &resp)
		assert.True(t, resp.Diagnostics.HasError())

		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, expectedErrorMsg, actualMsg)
	}
}

func TestSpotPercent_ValidateString(t *testing.T) {
	v := SpotPercent{}
	var resp validator.Int64Response
	var req validator.Int64Request

	validPercents := []int64{0, 50, 100}
	for _, validPercent := range validPercents {
		req.ConfigValue = types.Int64Value(validPercent)
		v.ValidateInt64(context.Background(), req, &resp)
		assert.False(t, resp.Diagnostics.HasError())
	}

	invalidPercents := []int64{-1, 101}
	for _, invalidPercent := range invalidPercents {
		var resp validator.Int64Response
		req.ConfigValue = types.Int64Value(invalidPercent)
		v.ValidateInt64(context.Background(), req, &resp)
		assert.True(t, resp.Diagnostics.HasError())

		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "Action rejected: spot_percent value must be between 0 and 100.", actualMsg)
	}
}

func TestSpotMarkup_ValidateFloat64(t *testing.T) {
	v := SpotMarkup{}
	var resp validator.Float64Response
	var req validator.Float64Request

	validPercents := []float64{0, 50, 100}
	for _, validPercent := range validPercents {
		req.ConfigValue = types.Float64Value(validPercent)
		v.ValidateFloat64(context.Background(), req, &resp)
		assert.False(t, resp.Diagnostics.HasError())
	}

	invalidPercents := []float64{-1, 101}
	for _, invalidPercent := range invalidPercents {
		var resp validator.Float64Response
		req.ConfigValue = types.Float64Value(invalidPercent)
		v.ValidateFloat64(context.Background(), req, &resp)
		assert.True(t, resp.Diagnostics.HasError())

		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "Action rejected: spot_markup value must be between 0 and 100.", actualMsg)
	}
}

func TestVCPUType_ValidateString(t *testing.T) {
	v := VCPUType{}
	var req validator.StringRequest

	validCPUTypes := []string{"shared", "standard", "hpc"}

	for _, validCPUType := range validCPUTypes {
		var resp validator.StringResponse

		req.ConfigValue = types.StringValue(validCPUType)
		v.ValidateString(context.Background(), req, &resp)
		assert.False(t, resp.Diagnostics.HasError())
	}

	var resp validator.StringResponse

	req.ConfigValue = types.StringValue("invalid")
	v.ValidateString(context.Background(), req, &resp)
	assert.True(t, resp.Diagnostics.HasError())

	actualMsg := resp.Diagnostics.Errors()[0].Detail()
	assert.Equal(t, "Action rejected: configuration_priority.vcpu_type value must be one of: shared, standard, hpc.", actualMsg)
}

func TestPriority_ValidateString(t *testing.T) {
	v := Priority{}
	var req validator.StringRequest

	validNames := []string{"low", "med", "high"}
	for _, validName := range validNames {
		var resp validator.StringResponse

		req.ConfigValue = types.StringValue(validName)
		v.ValidateString(context.Background(), req, &resp)
		assert.False(t, resp.Diagnostics.HasError())
	}

	var resp validator.StringResponse

	req.ConfigValue = types.StringValue("invalid")
	v.ValidateString(context.Background(), req, &resp)
	assert.True(t, resp.Diagnostics.HasError())

	actualMsg := resp.Diagnostics.Errors()[0].Detail()
	assert.Equal(t, "Action rejected: configuration_priority.priority value must be one of: low, med, high.", actualMsg)
}
