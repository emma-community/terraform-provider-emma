package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotEmptyList_ValidateList_NullList(t *testing.T) {
	v := NotEmptyList{}
	var resp validator.ListResponse
	var req validator.ListRequest

	req.ConfigValue = types.ListNull(types.StringType)
	req.Path = path.Root("test")

	v.ValidateList(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test array must contain at least 1 item", actualMsg)
	} else {
		assert.Fail(t, "Is not validating null list values")
	}
}

func TestNotEmptyList_ValidateList_UnknownList(t *testing.T) {
	v := NotEmptyList{}
	var resp validator.ListResponse
	var req validator.ListRequest

	req.ConfigValue = types.ListUnknown(types.StringType)
	req.Path = path.Root("test")

	v.ValidateList(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test array must contain at least 1 item", actualMsg)
	} else {
		assert.Fail(t, "Is not validating unknown list values")
	}
}

func TestNotEmptyList_ValidateList_EmptyList(t *testing.T) {
	v := NotEmptyList{}
	var resp validator.ListResponse
	var req validator.ListRequest

	req.ConfigValue, _ = types.ListValue(types.StringType, []attr.Value{})

	req.Path = path.Root("test")

	v.ValidateList(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test array must contain at least 1 item", actualMsg)
	} else {
		assert.Fail(t, "Is not validating empty list values")
	}
}

func TestNotEmptyList_ValidateList_NotEmptyList(t *testing.T) {
	v := NotEmptyList{}
	var resp validator.ListResponse
	var req validator.ListRequest

	req.ConfigValue, _ = types.ListValue(types.StringType, []attr.Value{types.StringValue("test")})

	req.Path = path.Root("test")

	v.ValidateList(context.Background(), req, &resp)

	assert.False(t, resp.Diagnostics.HasError())
}
