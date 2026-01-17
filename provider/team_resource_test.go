package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Team_Lifecycle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "ctfd_user" "ctfer" {
	name     = "CTFer"
	email    = "ctfer-io-team@protonmail.com"
	password = "password"
}

resource "ctfd_user" "sa" {
	name     = "SA Bot"
	email    = "ctfer-io-bot@protonmail.com"
	password = "sa-password"
}

resource "ctfd_team" "cybercombattants" {
	name = "Les cybercombattants de l'innovation"
	email = "lucastesson@protonmail.com"
	password = "password"
	members = [
	  ctfd_user.ctfer.id,
	  ctfd_user.sa.id,
	]
	captain = ctfd_user.ctfer.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ctfd_team.cybercombattants", "id"),
					resource.TestCheckResourceAttr("ctfd_team.cybercombattants", "members.#", "2"), // 2 members for regression on #269
				),
			},
			// ImportState testing
			{
				ResourceName:            "ctfd_team.cybercombattants",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // password can't be fetched from CTFd (security by design)
			},
			// Update and Read testing (ban team)
			{
				Config: providerConfig + `
resource "ctfd_user" "ctfer" {
	name     = "CTFer"
	email    = "ctfer-io-team@protonmail.com"
	password = "new-password"
}

resource "ctfd_user" "sa" {
	name     = "SA Bot"
	email    = "ctfer-io-bot@protonmail.com"
	password = "sa-password"
}

resource "ctfd_team" "cybercombattants" {
	name = "Les cybercombattants de l'innovation"
	email = "lucastesson@protonmail.com"
	password = "password"
	banned = true
	members = [
	  ctfd_user.ctfer.id,
	  ctfd_user.sa.id,
	]
	captain = ctfd_user.ctfer.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ctfd_team.cybercombattants", "members.#", "2"),
				),
			},
		},
	})
}
