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
        - NicolasFgrx
    EOT
	value    = 500
    decay    = 17
    minimum  = 50
    state    = "visible"

	flags = [{
		content = "24HIUT{Http_1s_n0t_s3cuR3}}"
	}]
	topics = [
		"Network"
	]
	tags = [
		"network",
    	"http"
	]
	hints = [{
		content = "HTTP exchanges are not ciphered."
	}, {
		content = "Content is POSTed in HTTP :)"
		cost    = 50
	}]
	files = [{
		name    = "something.txt",
		content = "I won't be really useful as a file, but I tried my best :)"
	}, {
		name       = "something-b64.txt",
		contentb64 = "SSB3b24ndCBiZSByZWFsbHkgdXNlZnVsbCBhcyBhIGZpbGUsIGJ1dCBJIHRyaWVkIG15IGJlc3QgOikK"
	}]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ctfd_challenge.http", "files.#", "2"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "flags.0.id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "hints.0.id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "files.0.id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "files.0.location"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "files.0.contentb64"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "files.1.id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "files.1.location"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.http", "files.1.content"),
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

	flags = [{
		content = "24HIUT{Http_1s_n0t_s3cuR3}}"
	}]
	topics = [
		"Network"
	]
	tags = [
		"network",
    	"http"
	]
	hints = [{
		content = "HTTP exchanges are not ciphered."
	}, {
		content = "Content is POSTed in HTTP :)"
		cost    = 50
	}]
	files = [{
		name    = "something.txt",
		content = "I won't be really useful as a file, but I tried my best :)"
	}]
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
	decay   = 17
	minimum = 50
	state   = "visible"
	requirements = {
		behavior      = "anonymized"
		prerequisites = [ctfd_challenge.http.id]
	}
  
	flags = [{
		content = "24HIUT{IcmpExfiltrationIsEasy}"
	}]
  
	topics = [
		"Network"
	]
	tags = [
		"network",
		"icmp"
	]

	hints = [{
		content = "Vous ne trouvez pas qu'il ya beaucoup de requêtes ICMP ?"
		cost    = 50
	}, {
		content = "Pour l'exo, le ttl a été modifié, tente un ` + "`ip.ttl<=20`" + `"
		cost    = 50
	}]
  
	files = [{
		name       = "icmp.pcap"
		contentb64 = "c29tZS1jb250ZW50Cg=="
	}]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ctfd_challenge.http", "files.#", "1"),
					resource.TestCheckResourceAttr("ctfd_challenge.icmp", "requirements.prerequisites.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
