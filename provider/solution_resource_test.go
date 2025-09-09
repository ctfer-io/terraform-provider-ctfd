package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Solution_Lifecycle(t *testing.T) {
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

resource "ctfd_solution" "wu" {
	challenge_id = ctfd_challenge_standard.example.id
	content      = "Here is how to solve the challenge: ..."
	state        = "visible"
}
`,
			},
			// ImportState testing
			{
				ResourceName:      "ctfd_solution.wu",
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

resource "ctfd_solution" "wu" {
	challenge_id = ctfd_challenge_standard.example.id
	content      = "Oopsi, we disclosed it !"
	state        = "hidden"
}
`,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
