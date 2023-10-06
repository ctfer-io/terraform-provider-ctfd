package provider

import (
	"context"

	"github.com/opentofu/terraform-plugin-log/tflog"
)

func addSensitive(ctx context.Context, key string, value any) context.Context {
	ctx = tflog.SetField(ctx, key, value)
	return tflog.MaskFieldValuesWithFieldKeys(ctx, key)
}
