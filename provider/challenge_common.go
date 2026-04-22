package provider

import (
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/utils"
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

func GetBehavior(anon *string) types.String {
	if anon == nil {
		return BehaviorHidden // nil is hidden (default value)
	}
	switch *anon {
	case "true":
		return BehaviorAnonymized
	case "preview":
		return BehaviorPreview
	default: // "false" or anything new
		return BehaviorHidden
	}
}

func FromBehavior(bhv types.String) *string {
	switch bhv {
	case BehaviorAnonymized:
		return utils.Ptr("true")
	case BehaviorPreview:
		return utils.Ptr("preview")
	default:
		return utils.Ptr("false") // default value is hidden
	}
}
