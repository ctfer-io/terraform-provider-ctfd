package utils

import (
	"strings"

	"github.com/opentofu/terraform-plugin-framework/types"
)

// return a null types.Int64 if pointer is nil, else its value
func ToTFInt64(i *int) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
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
