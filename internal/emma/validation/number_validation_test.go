package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPositiveInt64_ValidateInt64_ZeroValue(t *testing.T) {
	v := PositiveInt64{}
	var resp validator.Int64Response
	var req validator.Int64Request

	req.ConfigValue = types.Int64Value(-1)
	req.Path = path.Root("test")

	v.ValidateInt64(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test must be greater than 0", actualMsg)
	} else {
		assert.Fail(t, "Is not validating zero int64 values")
	}
}

func TestPositiveInt64_ValidateInt64_NegativeValue(t *testing.T) {
	v := PositiveInt64{}
	var resp validator.Int64Response
	var req validator.Int64Request

	req.ConfigValue = types.Int64Value(-1)
	req.Path = path.Root("test")

	v.ValidateInt64(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test must be greater than 0", actualMsg)
	} else {
		assert.Fail(t, "Is not validating negative int64 values")
	}
}

func TestPositiveInt64_ValidateInt64_PositiveValue(t *testing.T) {
	v := PositiveInt64{}
	var resp validator.Int64Response
	var req validator.Int64Request

	req.ConfigValue = types.Int64Value(1)
	req.Path = path.Root("test")

	v.ValidateInt64(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestPositiveFloat64_ValidateFloat64_ZeroValue(t *testing.T) {
	v := PositiveFloat64{}
	var resp validator.Float64Response
	var req validator.Float64Request

	req.ConfigValue = types.Float64Value(0)
	req.Path = path.Root("test")

	v.ValidateFloat64(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test must be greater than 0", actualMsg)
	} else {
		assert.Fail(t, "Is not validating zero float64 values")
	}
}

func TestPositiveFloat64_ValidateFloat64_NegativeValue(t *testing.T) {
	v := PositiveFloat64{}
	var resp validator.Float64Response
	var req validator.Float64Request

	req.ConfigValue = types.Float64Value(-1)
	req.Path = path.Root("test")

	v.ValidateFloat64(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test must be greater than 0", actualMsg)
	} else {
		assert.Fail(t, "Is not validating negative float64 values")
	}
}

func TestPositiveFloat64_ValidateFloat64_PositiveValue(t *testing.T) {
	v := PositiveFloat64{}
	var resp validator.Float64Response
	var req validator.Float64Request

	req.ConfigValue = types.Float64Value(1)
	req.Path = path.Root("test")

	v.ValidateFloat64(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError())
}
