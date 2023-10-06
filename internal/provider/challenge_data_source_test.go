package provider_test

import (
	"testing"

	"github.com/opentofu/terraform-plugin-testing/helper/resource"
)

func TestAccChallengeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `data "ctfd_challenge" "test" {}`,
				Check:  resource.C,
			},
		},
	})
}
