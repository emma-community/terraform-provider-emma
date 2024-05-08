package emma

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type NotEmptyList struct {
}

func (v NotEmptyList) Description(ctx context.Context) string {
	return "value array must contain at least 1 item"
}

func (v NotEmptyList) MarkdownDescription(ctx context.Context) string {
	return "value array must contain at least 1 item"
}

func (v NotEmptyList) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() || len(req.ConfigValue.Elements()) == 0 {
		resp.Diagnostics.AddError("Validation Error", req.Path.String()+" array must contain at least 1 item")
	}
}
