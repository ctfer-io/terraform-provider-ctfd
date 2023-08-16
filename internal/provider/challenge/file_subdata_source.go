package challenge

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

func fileSubdatasourceAttributes() map[string]schema.Attribute {
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
