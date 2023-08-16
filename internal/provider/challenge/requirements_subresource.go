package challenge

import "github.com/hashicorp/terraform-plugin-framework/types"

var (
	behaviorHidden     = types.StringValue("hidden")
	behaviorAnonymized = types.StringValue("anonymized")
)

type requirementsSubresourceModel struct {
	Behavior      types.String   `tfsdk:"behavior"`
	Prerequisites []types.String `tfsdk:"prerequisites"`
}

func getAnon(str types.String) *bool {
	switch {
	case str.Equal(behaviorHidden):
		return nil
	case str.Equal(behaviorAnonymized):
		return ptr(true)
	}
	panic("invalid anonymization value: " + str.ValueString())
}

func fromAnon(b *bool) types.String {
	if b == nil {
		return behaviorHidden
	}
	if *b {
		return behaviorAnonymized
	}
	panic("invalid anonymization value, got boolean false")
}
