package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotBlankString_ValidateString_NullValue(t *testing.T) {
	v := NotBlankString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringNull()
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test is required", actualMsg)
	} else {
		assert.Fail(t, "Blank is not validating null values")
	}
}

func TestNotBlankString_ValidateString_UnknownValue(t *testing.T) {
	v := NotBlankString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringUnknown()
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test is required", actualMsg)
	} else {
		assert.Fail(t, "Blank is not validating unknown values")
	}
}

func TestNotBlankString_ValidateString_EmptyValue(t *testing.T) {
	v := NotBlankString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test is required", actualMsg)
	} else {
		assert.Fail(t, "Blank is not validating empty values")
	}
}

func TestNotBlankString_ValidateString_BlankValue(t *testing.T) {
	v := NotBlankString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue(" ")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test is required", actualMsg)
	} else {
		assert.Fail(t, "Blank is not validating blank values")
	}
}

func TestNotBlankString_ValidateString_FillValue(t *testing.T) {
	v := NotBlankString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("test")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError(), "Blank must continue with filled fields")
}

func TestNotEmptyString_ValidateString_EmptyValue(t *testing.T) {
	v := NotEmptyString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test must not be empty", actualMsg)
	} else {
		assert.Fail(t, "Empty is not validating empty values")
	}
}

func TestNotEmptyString_ValidateString_BlankValue(t *testing.T) {
	v := NotEmptyString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue(" ")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test must not be empty", actualMsg)
	} else {
		assert.Fail(t, "Empty is not validating blank values")
	}
}

func TestNotEmptyString_ValidateString_FillValue(t *testing.T) {
	v := NotEmptyString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("test")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError(), "Empty must continue with filled value")
}

func TestNotEmptyString_ValidateString_NullValue(t *testing.T) {
	v := NotEmptyString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringNull()
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError(), "Empty must continue with null value")
}

func TestNotEmptyString_ValidateString_UnknownValue(t *testing.T) {
	v := NotEmptyString{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringUnknown()
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError(), "Empty must continue with unknown value")
}
