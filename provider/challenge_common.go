package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	BehaviorHidden     = types.StringValue("hidden")
	BehaviorAnonymized = types.StringValue("anonymized")
	BehaviorPreview    = types.StringValue("preview")

	FunctionLinear      = types.StringValue("linear")
	FunctionLogarithmic = types.StringValue("logarithmic")
)

type RequirementsSubresourceModel struct {
	Behavior      types.String   `tfsdk:"behavior"`
	Prerequisites []types.String `tfsdk:"prerequisites"`
}
