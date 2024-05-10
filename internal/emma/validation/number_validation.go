package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type PositiveInt64 struct {
}

func (v PositiveInt64) Description(ctx context.Context) string {
	return "value must be greater than 0"
}

func (v PositiveInt64) MarkdownDescription(ctx context.Context) string {
	return "value must be greater than 0"
}

func (v PositiveInt64) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if req.ConfigValue.ValueInt64() <= 0 {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" must be greater than 0")
	}
}

type PositiveFloat64 struct {
}

func (v PositiveFloat64) Description(ctx context.Context) string {
	return "value must be greater than 0"
}

func (v PositiveFloat64) MarkdownDescription(ctx context.Context) string {
	return "value must be greater than 0"
}

func (v PositiveFloat64) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	if req.ConfigValue.ValueFloat64() <= 0 {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" must be greater than 0")
	}
}
