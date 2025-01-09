package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"strings"
	"unicode"
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

type UserPassword struct{}

func (v UserPassword) Description(ctx context.Context) string {
	return "The user_password must consist of 8 to 60 characters, including both upper- and lower-case Latin letters, digits, and symbols (|~`\"!@#$%&,.)"
}

func (v UserPassword) MarkdownDescription(ctx context.Context) string {
	return "The user_password must consist of 8 to 60 characters, including both upper- and lower-case Latin letters, digits, and symbols (|~`\"!@#$%&,.)"
}

func (v UserPassword) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	password := req.ConfigValue.ValueString()

	if len(password) < 8 || len(password) > 60 {
		resp.Diagnostics.AddError("Validation Error", "Validation error, user_password must consist of 8 to 60 characters, including both upper- and lower-case Latin letters, digits, and symbols (|~`\"!@#$%&,.).")
		return
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	specialChars := "!@#$%^&*~|=+`,_\"'\\-"

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}

	if !(hasLower && hasUpper && hasDigit && hasSpecial) {
		resp.Diagnostics.AddError("Validation Error", "Validation error, user_password must consist of 8 to 60 characters, including both upper- and lower-case Latin letters, digits, and symbols (|~`\"!@#$%&,.).")
	}
}
