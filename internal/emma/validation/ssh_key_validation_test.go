package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeyType_ValidateString_InvalidValue(t *testing.T) {
	v := KeyType{}
	var resp validator.StringResponse
	var req validator.StringRequest

	req.ConfigValue = types.StringValue("test")
	req.Path = path.Root("test")

	v.ValidateString(context.Background(), req, &resp)

	assert.Equal(t, 1, resp.Diagnostics.ErrorsCount())
	if resp.Diagnostics.HasError() {
		actualMsg := resp.Diagnostics.Errors()[0].Detail()
		assert.Equal(t, "test can contain RSA or ED25519", actualMsg)
	} else {
		assert.Fail(t, "Is valid key_type value: test")
	}
}

func TestKeyType_ValidateString_ValidValues(t *testing.T) {
	for _, validKeyTypeValue := range []string{"RSA", "ED25519"} {
		v := KeyType{}
		var resp validator.StringResponse
		var req validator.StringRequest

		req.ConfigValue = types.StringValue(validKeyTypeValue)
		req.Path = path.Root("test")

		v.ValidateString(context.Background(), req, &resp)

		assert.False(t, resp.Diagnostics.HasError(), "Is invalid key_type value: "+validKeyTypeValue)
	}
}
