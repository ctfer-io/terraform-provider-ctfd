package challenge

import (
	"github.com/ctfer-io/terraform-provider-ctfd/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	BehaviorHidden     = types.StringValue("hidden")
	BehaviorAnonymized = types.StringValue("anonymized")
)

type RequirementsSubresourceModel struct {
	Behavior      types.String   `tfsdk:"behavior"`
	Prerequisites []types.String `tfsdk:"prerequisites"`
}

func GetAnon(str types.String) *bool {
	switch {
	case str.Equal(BehaviorHidden):
		return nil
	case str.Equal(BehaviorAnonymized):
		return utils.Ptr(true)
	}
	panic("invalid anonymization value: " + str.ValueString())
}

func FromAnon(b *bool) types.String {
	if b == nil {
		return BehaviorHidden
	}
	if *b {
		return BehaviorAnonymized
	}
	panic("invalid anonymization value, got boolean false")
}
