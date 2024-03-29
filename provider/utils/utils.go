package utils

import (
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// return a null types.Int64 if pointer is nil, else its value
func ToTFInt64(i *int) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
}

func ToTFString(str *string) types.String {
	if str == nil {
		return types.StringNull()
	}
	return types.StringValue(*str)
}

// return a nil point if types.Int64 is null, else its value
func ToInt(itf types.Int64) *int {
	if itf.IsNull() {
		return nil
	}
	i := int(itf.ValueInt64())
	return &i
}

func Filename(location string) string {
	pts := strings.Split(location, "/")
	return pts[len(pts)-1]
}

func Ptr[T any](t T) *T {
	return &t
}

// Atoi MUST only be called on trusted input as it won't
// return an error nor panic after calling `strconv.Atoi`.
func Atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
