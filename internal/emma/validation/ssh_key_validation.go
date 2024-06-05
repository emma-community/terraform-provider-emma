package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
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

type SshKeyName struct {
}

func (v SshKeyName) Description(ctx context.Context) string {
	return "name must be alphanumeric, hyphens, dots and underscores; length can be from 1 to 64 symbols"
}

func (v SshKeyName) MarkdownDescription(ctx context.Context) string {
	return "name must be alphanumeric, hyphens, dots and underscores; length can be from 1 to 64 symbols"
}

func (v SshKeyName) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	matches, _ := regexp.MatchString(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`, req.ConfigValue.ValueString())
	if !matches {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" must be alphanumeric, hyphens, dots and underscores; length can be from 1 to 64 symbols")
	}
}
