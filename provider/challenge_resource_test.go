package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Challenge_Lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "ctfd_challenge" "http" {
	name        = "HTTP Authentication"
	category    = "network"
	description = <<-EOT
        Oh no ! I did not see my connection was no encrypted !
        I hope no one spied me...

        Authors:
        - Nicolas
    EOT
	value    = 500
    decay    = 20
    minimum  = 50
    state    = "hidden"

	topics = [
		"Network"
	]
	tags = [
		"network"
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ctfd_challenge.http",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ctfd_challenge" "http" {
	name        = "HTTP Authentication"
	category    = "network"
	description = <<-EOT
        Oh no ! I did not see my connection was no encrypted !
        I hope no one spied me...

        Authors:
        - NicolasFgrx
    EOT
	value    = 500
    decay    = 17
    minimum  = 50
    state    = "visible"

	topics = [
		"Network"
	]
	tags = [
		"network",
    	"http"
	]
}

resource "ctfd_challenge" "icmp" {
	name        = "Stealing data"
	category    = "network"
	description = <<-EOT
		The network administrator signaled some strange content send to a server.
		At first glance, it seems to be an internal one. Can you tell what it is ?

		(The network capture was realized out of the CTF infrastructure)

		Authors:
		- NicolasFgrx
	EOT
	value   = 500

	requirements = {
		behavior      = "anonymized"
		prerequisites = [ctfd_challenge.http.id]
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ctfd_challenge.icmp", "requirements.prerequisites.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
