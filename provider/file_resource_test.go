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
resource "ctfd_challenge_standard" "example" {
	name        = "Example challenge"
	category    = "test"
	description = "Example challenge description..."
	value       = 500
}

resource "ctfd_file" "pouet" {
	challenge_id = ctfd_challenge_standard.example.id
	name         = "pouet.txt"
	content      = "Pouet is a clown cat"
}

resource "ctfd_file" "pouet_2" {
	name         = "pouet-2.txt"
	content      = "Pouet is a clown cat, but has not challenge"
}
`,
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
resource "ctfd_challenge_standard" "example" {
	name        = "Example challenge"
	category    = "test"
	description = "Example challenge description..."
	value       = 500
}

resource "ctfd_file" "pouet" {
	challenge_id = ctfd_challenge_standard.example.id
	name         = "pouet.txt"
	content      = "Pouet the 2nd is the clowniest cat ever"
}

resource "ctfd_file" "pouet_2" {
	name         = "pouet-2.txt"
	content      = "Pouet is a clown cat, but has not challenge"
}
`,
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
