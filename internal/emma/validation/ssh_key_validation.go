package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type KeyType struct {
}

func (v KeyType) Description(ctx context.Context) string {
	return "key_type can contain RSA or ED25519"
}

func (v KeyType) MarkdownDescription(ctx context.Context) string {
	return "key_type can contain RSA or ED25519"
}

func (v KeyType) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if req.ConfigValue.ValueString() != "RSA" && req.ConfigValue.ValueString() != "ED25519" {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" can contain RSA or ED25519")
	}
}
