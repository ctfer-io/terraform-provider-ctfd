package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Hint_Lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "ctfd_challenge_standard" "example" {
	name        = "Example challenge"
	category    = "test"
	description = "Example challenge description..."
	value       = 500
}

resource "ctfd_hint" "first" {
	challenge_id = ctfd_challenge_standard.example.id
	title        = "1st"
	content      = "This is a first hint"
	cost         = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttr("ctfd_hint.first", "requirements.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ctfd_hint.first",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ctfd_challenge_standard" "example" {
	name        = "Example challenge"
	category    = "test"
	description = "Example challenge description..."
	value       = 500
}

resource "ctfd_hint" "first" {
	challenge_id = ctfd_challenge_standard.example.id
	title        = "First"
	content      = "This is a first hint"
	cost         = 1
}

resource "ctfd_hint" "second" {
	challenge_id = ctfd_challenge_standard.example.id
	content      = "This is a second hint"
	cost         = 2
	requirements = [ctfd_hint.first.id]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ctfd_hint.first", "requirements.#", "0"),
					resource.TestCheckResourceAttr("ctfd_hint.second", "requirements.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
