package challenge

import "github.com/opentofu/terraform-plugin-framework/datasource/schema"

func FileSubdatasourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"location": schema.StringAttribute{
			Computed: true,
		},
		"content": schema.StringAttribute{
			Computed: true,
		},
		"contentb64": schema.StringAttribute{
			Computed: true,
		},
	}
}
