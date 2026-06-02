package emma

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MutuallyExclusive validates that only one of the specified fields is set
type MutuallyExclusive struct {
	Fields []string
}

func (v MutuallyExclusive) Description(ctx context.Context) string {
	return fmt.Sprintf("Only one of %v can be specified", v.Fields)
}

func (v MutuallyExclusive) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v MutuallyExclusive) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the current field is null or unknown, no validation needed
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Count how many fields in the group are set
	setFields := []string{}
	
	// Current field is set (we're validating it)
	currentFieldName := req.Path.String()
	setFields = append(setFields, currentFieldName)

	// Check other fields in the mutually exclusive group
	for _, fieldName := range v.Fields {
		// Skip the current field
		if fieldName == currentFieldName {
			continue
		}

		// Get the value of the other field from config
		var fieldValue attr.Value
		fieldPath := path.Root(fieldName)
		diags := req.Config.GetAttribute(ctx, fieldPath, &fieldValue)
		
		if diags.HasError() {
			// Field doesn't exist or can't be accessed, skip it
			continue
		}

		// Check if the field is set (not null and not unknown)
		if !fieldValue.IsNull() && !fieldValue.IsUnknown() {
			// For string values, also check if not empty
			if strVal, ok := fieldValue.(types.String); ok {
				if strVal.ValueString() != "" {
					setFields = append(setFields, fieldName)
				}
			} else {
				setFields = append(setFields, fieldName)
			}
		}
	}

	// If more than one field is set, validation fails
	if len(setFields) > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Mutually Exclusive Fields",
			fmt.Sprintf("Only one of %v can be specified, but found: %v",
				v.Fields, strings.Join(setFields, ", ")),
		)
	}
}

// RequiresOneOf validates that at least one of the specified fields is set
type RequiresOneOf struct {
	Fields []string
}

func (v RequiresOneOf) Description(ctx context.Context) string {
	return fmt.Sprintf("At least one of %v must be specified", v.Fields)
}

func (v RequiresOneOf) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v RequiresOneOf) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// Count how many fields in the group are set
	setCount := 0

	// Check if current field is set
	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() && req.ConfigValue.ValueString() != "" {
		setCount++
	}

	// Check other fields in the group
	currentFieldName := req.Path.String()
	for _, fieldName := range v.Fields {
		// Skip the current field (already counted)
		if fieldName == currentFieldName {
			continue
		}

		// Get the value of the other field from config
		var fieldValue attr.Value
		fieldPath := path.Root(fieldName)
		diags := req.Config.GetAttribute(ctx, fieldPath, &fieldValue)
		
		if diags.HasError() {
			// Field doesn't exist or can't be accessed, skip it
			continue
		}

		// Check if the field is set (not null and not unknown)
		if !fieldValue.IsNull() && !fieldValue.IsUnknown() {
			// For string values, also check if not empty
			if strVal, ok := fieldValue.(types.String); ok {
				if strVal.ValueString() != "" {
					setCount++
				}
			} else {
				setCount++
			}
		}
	}

	// If no fields are set, validation fails
	if setCount == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Required Field Missing",
			fmt.Sprintf("At least one of %v must be specified", v.Fields),
		)
	}
}

// ValidateInt64 implementation for RequiresOneOf to support int64 fields
func (v RequiresOneOf) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	// Count how many fields in the group are set
	setCount := 0

	// Check if current field is set
	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		setCount++
	}

	// Check other fields in the group
	currentFieldName := req.Path.String()
	for _, fieldName := range v.Fields {
		// Skip the current field (already counted)
		if fieldName == currentFieldName {
			continue
		}

		// Get the value of the other field from config
		var fieldValue attr.Value
		fieldPath := path.Root(fieldName)
		diags := req.Config.GetAttribute(ctx, fieldPath, &fieldValue)
		
		if diags.HasError() {
			// Field doesn't exist or can't be accessed, skip it
			continue
		}

		// Check if the field is set (not null and not unknown)
		if !fieldValue.IsNull() && !fieldValue.IsUnknown() {
			// For string values, also check if not empty
			if strVal, ok := fieldValue.(types.String); ok {
				if strVal.ValueString() != "" {
					setCount++
				}
			} else {
				setCount++
			}
		}
	}

	// If no fields are set, validation fails
	if setCount == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Required Field Missing",
			fmt.Sprintf("At least one of %v must be specified", v.Fields),
		)
	}
}

// RequiredTogether validates that if this field is set, the companion field must also be set (and vice versa).
type RequiredTogether struct {
	CompanionField string
}

func (v RequiredTogether) Description(ctx context.Context) string {
	return fmt.Sprintf("Must be specified together with %s", v.CompanionField)
}

func (v RequiredTogether) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v RequiredTogether) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var companion attr.Value
	diags := req.Config.GetAttribute(ctx, path.Root(v.CompanionField), &companion)
	if diags.HasError() || companion.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Companion Field",
			fmt.Sprintf("%s must be specified together with %s", req.Path, v.CompanionField),
		)
	}
}

func (v RequiredTogether) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var companion attr.Value
	diags := req.Config.GetAttribute(ctx, path.Root(v.CompanionField), &companion)
	if diags.HasError() || companion.IsNull() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Missing Required Companion Field",
			fmt.Sprintf("%s must be specified together with %s", req.Path, v.CompanionField),
		)
	}
}

// ValidateInt64 implementation for MutuallyExclusive to support int64 fields
func (v MutuallyExclusive) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	// If the current field is null or unknown, no validation needed
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Count how many fields in the group are set
	setFields := []string{}
	
	// Current field is set (we're validating it)
	currentFieldName := req.Path.String()
	setFields = append(setFields, currentFieldName)

	// Check other fields in the mutually exclusive group
	for _, fieldName := range v.Fields {
		// Skip the current field
		if fieldName == currentFieldName {
			continue
		}

		// Get the value of the other field from config
		var fieldValue attr.Value
		fieldPath := path.Root(fieldName)
		diags := req.Config.GetAttribute(ctx, fieldPath, &fieldValue)
		
		if diags.HasError() {
			// Field doesn't exist or can't be accessed, skip it
			continue
		}

		// Check if the field is set (not null and not unknown)
		if !fieldValue.IsNull() && !fieldValue.IsUnknown() {
			// For string values, also check if not empty
			if strVal, ok := fieldValue.(types.String); ok {
				if strVal.ValueString() != "" {
					setFields = append(setFields, fieldName)
				}
			} else {
				setFields = append(setFields, fieldName)
			}
		}
	}

	// If more than one field is set, validation fails
	if len(setFields) > 1 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Mutually Exclusive Fields",
			fmt.Sprintf("Only one of %v can be specified, but found: %v",
				v.Fields, strings.Join(setFields, ", ")),
		)
	}
}
