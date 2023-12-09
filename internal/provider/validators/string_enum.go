package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringEnumValidator valides a string value in an enumeration.
type StringEnumValidator struct {
	values []types.String
}

func NewStringEnumValidator(values []types.String) *StringEnumValidator {
	return &StringEnumValidator{
		values: values,
	}
}

var _ validator.String = (*StringEnumValidator)(nil)

func (val *StringEnumValidator) Description(ctx context.Context) string {
	return "Validates a string value in an enumeration."
}

func (val *StringEnumValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates a string value in an enumeration."
}

func (val *StringEnumValidator) ValidateString(ctx context.Context, req validator.StringRequest, res *validator.StringResponse) {
	if req.ConfigValue.IsNull() {
		return
	}

	if req.ConfigValue.IsUnknown() {
		return
	}

	for _, v := range val.values {
		if req.ConfigValue.Equal(v) {
			return
		}
	}
	res.Diagnostics.AddError(
		"StringEnumValidator Error",
		"No matching values.",
	)
}
