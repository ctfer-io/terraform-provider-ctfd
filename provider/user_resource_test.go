package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_User_Lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "ctfd_user" "ctfer" {
	name     = "CTFer"
	email    = "ctfer-io-user@protonmail.com"
	password = "password"

	# Define as an administration account
	type     = "admin"
	verified = true
	hidden   = true
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "id"),
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "type"),
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "verified"),
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "hidden"),
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "banned"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "ctfd_user.ctfer",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // password can't be fetched from CTFd (security by design)
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ctfd_user" "ctfer" {
	name     = "CTFer"
	email    = "ctfer-io-user@protonmail.com"
	password = "password"
}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "id"),
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "type"),
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "verified"),
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "hidden"),
					resource.TestCheckResourceAttrSet("ctfd_user.ctfer", "banned"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
