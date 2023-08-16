package challenge

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func flagSubdatasourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"content": schema.StringAttribute{
			Computed: true,
		},
		"data": schema.StringAttribute{
			Computed: true,
		},
		"type": schema.StringAttribute{
			Computed: true,
		},
	}
}
