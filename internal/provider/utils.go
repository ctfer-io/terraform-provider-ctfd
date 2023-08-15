package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// return a null types.Int64 if pointer is nil, else its value
func toTFInt64(i *int) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
}

// return a nil point if types.Int64 is null, else its value
func toInt(itf types.Int64) *int {
	if itf.IsNull() {
		return nil
	}
	i := int(itf.ValueInt64())
	return &i
}

func addSensitive(ctx context.Context, key string, value any) context.Context {
	ctx = tflog.SetField(ctx, key, value)
	return tflog.MaskFieldValuesWithFieldKeys(ctx, key)
}

func filename(location string) string {
	pts := strings.Split(location, "/")
	return pts[len(pts)-1]
}

func ptr[T any](t T) *T {
	return &t
}
