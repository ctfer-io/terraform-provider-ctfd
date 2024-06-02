package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_File_Lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
#resource "ctfd_challenge" "example" {
#	name        = "Example challenge"
#	category    = "test"
#	description = "Example challenge description..."
#	value       = 500
#}

resource "ctfd_file" "pouet" {
	challenge_id = "1"
	name         = "pouet.txt"
	content      = "Pouet is a clown cat"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
				// Verify dynamic values have any value set in the state.
				// resource.TestCheckResourceAttr("ctfd_flag.first", "requirements.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ctfd_file.pouet",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
#resource "ctfd_challenge" "example" {
#	name        = "Example challenge"
#	category    = "test"
#	description = "Example challenge description..."
#	value       = 500
#}

resource "ctfd_file" "pouet" {
	challenge_id = "1"
	name         = "pouet.txt"
	content      = "Pouet the 2nd is the clowniest cat ever"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
				// resource.TestCheckResourceAttr("ctfd_flag.first", "requirements.#", "0"),
				// resource.TestCheckResourceAttr("ctfd_flag.second", "requirements.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
