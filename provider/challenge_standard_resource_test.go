package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_ChallengeStandard_Lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "ctfd_challenge_standard" "http" {
	name        = "HTTP Authentication"
	category    = "network"
	description = <<-EOT
        Oh no ! I did not see my connection was no encrypted !
        I hope no one spied me...
    EOT
	attribution = "Nicolas"
	value       = 500
    state       = "hidden"

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
					resource.TestCheckResourceAttrSet("ctfd_challenge_standard.http", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ctfd_challenge_standard.http",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ctfd_challenge_standard" "http" {
	name        = "HTTP Authentication"
	category    = "network"
	description = <<-EOT
        Oh no ! I did not see my connection was no encrypted !
        I hope no one spied me...
    EOT
	attribution = "NicolasFgrx"
	value       = 500
    state       = "visible"

	topics = [
		"Network"
	]
	tags = [
		"network",
    	"http"
	]
}

resource "ctfd_challenge_standard" "icmp" {
	name        = "Stealing data"
	category    = "network"
	description = <<-EOT
		The network administrator signaled some strange content send to a server.
		At first glance, it seems to be an internal one. Can you tell what it is ?

		(The network capture was realized out of the CTF infrastructure)
	EOT
	attribution = "NicolasFgrx"
	value       = 500

	requirements = {
		behavior      = "anonymized"
		prerequisites = [ctfd_challenge_standard.http.id]
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ctfd_challenge_standard.icmp", "requirements.prerequisites.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
