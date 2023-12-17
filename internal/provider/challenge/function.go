// This file is related to the challenge.function attribute

package challenge

import "github.com/hashicorp/terraform-plugin-framework/types"

var (
	FunctionLinear      = types.StringValue("linear")
	FunctionLogarithmic = types.StringValue("logarithmic")
)
