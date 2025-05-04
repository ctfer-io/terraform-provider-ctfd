package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Brackets_Lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "ctfd_bracket" "juniors" {
	name        = "Juniors"
	description = "Bracket for 14-25 years old."
	type        = "users"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("ctfd_bracket.juniors", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ctfd_bracket.juniors",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ctfd_bracket" "juniors" {
	name        = "Juniors"
	description = "Bracket for 14-25 years old players."
	type        = "users"
}

resource "ctfd_user" "player1" {
	name     = "player1"
	email    = "player1@ctfer.io"
	password = "password"

	bracket_id = ctfd_bracket.juniors.id
}

resource "ctfd_bracket" "seniors" {
	name        = "Seniors"
	description = "Bracket for >25 yers old players."
	type        = "teams"
}

resource "ctfd_team" "team1" {
	name = "team1"
	email = "team1@ctfer.io"
	password = "password"
	members = [
	  ctfd_user.player1.id,
	]
	captain = ctfd_user.player1.id

	bracket_id = ctfd_bracket.seniors.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ctfd_bracket.juniors", "id"),
					resource.TestCheckResourceAttrSet("ctfd_bracket.seniors", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
