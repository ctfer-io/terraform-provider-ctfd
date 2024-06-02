package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Flag_Lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "ctfd_challenge" "example" {
	name        = "Example challenge"
	category    = "test"
	description = "Example challenge description..."
	value       = 500
}

resource "ctfd_flag" "static" {
	challenge_id = ctfd_challenge.example.id
	content      = "This is a first flag"
	type         = "static"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
				// Verify dynamic values have any value set in the state.
				// resource.TestCheckResourceAttr("ctfd_flag.first", "requirements.#", "0"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ctfd_flag.static",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ctfd_challenge" "example" {
	name        = "Example challenge"
	category    = "test"
	description = "Example challenge description..."
	value       = 500
}

resource "ctfd_flag" "static" {
	challenge_id = ctfd_challenge.example.id
	content      = "This is a first flag"
	data         = "case_insensitive"
	type         = "static"
}

resource "ctfd_flag" "regex" {
	challenge_id = ctfd_challenge.example.id
	content      = "CTFER{.*}"
	type         = "regex"
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
