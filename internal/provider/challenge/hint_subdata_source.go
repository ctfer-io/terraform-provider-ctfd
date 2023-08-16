package challenge

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func hintSubdatasourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"content": schema.StringAttribute{
			Computed: true,
		},
		"cost": schema.Int64Attribute{
			Computed: true,
		},
		"requirements": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
		},
	}
}
