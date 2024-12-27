package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"strings"
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

func TestUserPassword_ValidateString_NotCheckedValues(t *testing.T) {
	v := UserPassword{}

	notCheckedValues := []types.String{
		types.StringNull(),
		types.StringUnknown(),
	}

	for _, value := range notCheckedValues {
		var resp validator.StringResponse
		var req validator.StringRequest
		req.ConfigValue = value
		v.ValidateString(context.Background(), req, &resp)

		assert.Equal(t, 0, resp.Diagnostics.ErrorsCount(), "Value should not be checked: '%s'", value)
		if resp.Diagnostics.HasError() {
			assert.Fail(t, "Value should not be checked: '%s'", value)
		}
	}
}

func TestUserPassword_ValidateString_InvalidValues(t *testing.T) {
	v := UserPassword{}

	invalidValues := []string{
		"",
		" ",
		"j5!Ha",
		strings.Repeat("j5!Ha", 20),
		"j5HaHaHaHa", // no special character
		"j5!hahaha",  // no upper case
		"J5!HAHAHA",  // no lower case
		"j!HaHaHaHa", // no digit

	}

	for _, value := range invalidValues {
		var resp validator.StringResponse
		var req validator.StringRequest
		req.ConfigValue = types.StringValue(value)
		v.ValidateString(context.Background(), req, &resp)

		assert.Equal(t, 1, resp.Diagnostics.ErrorsCount(), "Expected validation error for value: '%s'", value)
		if resp.Diagnostics.HasError() {
			actualMsg := resp.Diagnostics.Errors()[0].Detail()
			assert.Equal(t, "Validation error, user_password must consist of 8 to 60 characters, including both upper- and lower-case Latin letters, digits, and symbols (|~`\"!@#$%&,.).", actualMsg)
		} else {
			assert.Fail(t, "UserPassword is not validating value: '%s'", value)
		}
	}
}

func TestUserPassword_ValidateString_ValidValues(t *testing.T) {
	v := UserPassword{}

	validValues := []string{
		"j5HaHaH!",                  // minimum length
		strings.Repeat("j5!Ha", 12), // maximum length
	}
	specialChars := "!@#$%^&*~|=+`,_\"'\\-"
	for _, char := range specialChars {
		validValues = append(validValues, "j5HaHaH"+string(char))
	}

	for _, value := range validValues {
		var resp validator.StringResponse
		var req validator.StringRequest
		req.ConfigValue = types.StringValue(value)
		v.ValidateString(context.Background(), req, &resp)

		assert.Equal(t, 0, resp.Diagnostics.ErrorsCount(), "UserPassword should be valid: '%s'", value)
		if resp.Diagnostics.HasError() {
			assert.Fail(t, "UserPassword should be valid: '%s'", value)
		}
	}
}
