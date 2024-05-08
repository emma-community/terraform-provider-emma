package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"strings"
)

type NotBlankString struct {
}

func (v NotBlankString) Description(ctx context.Context) string {
	return "value is blank"
}

func (v NotBlankString) MarkdownDescription(ctx context.Context) string {
	return "value is blank"
}

func (v NotBlankString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() || len(strings.TrimSpace(req.ConfigValue.ValueString())) == 0 {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" is required")
	}
}

type NotEmptyString struct {
}

func (v NotEmptyString) Description(ctx context.Context) string {
	return "value is empty"
}

func (v NotEmptyString) MarkdownDescription(ctx context.Context) string {
	return "value is empty"
}

func (v NotEmptyString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if len(strings.TrimSpace(req.ConfigValue.ValueString())) == 0 {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" must not be empty")
	}
}
