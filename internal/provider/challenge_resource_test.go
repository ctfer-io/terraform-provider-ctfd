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
resource "ctfd_challenge" "test" {
	name        = "Test Challenge"
	category    = "acc"
	description = <<-EOT
		This is some content.

		And it is multiline :o
	EOT
	value    = 500
    decay    = 17
    minimum  = 50
    state    = "visible"

	flags = [{
		content = "CTFER{acc_test}"
	}]
	topics = [
		"Acceptance"
	]
	tags = [
		"acceptance",
		"testing"
	]
	hints = [{
		content = "C'mon this is a test dude..."
	}, {
		content = "There is nothing to find here, it is just a test !"
		cost = 50
	}]
	files = [{
		name    = "something.txt",
		content = "I won't be really usefull as a file, but I tried my best :)"
	}, {
		name       = "something-b64.txt",
		contentb64 = "SSB3b24ndCBiZSByZWFsbHkgdXNlZnVsbCBhcyBhIGZpbGUsIGJ1dCBJIHRyaWVkIG15IGJlc3QgOikK"
	}]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ctfd_challenge.test", "files.#", "2"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("ctfd_challenge.test", "id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.test", "flags.0.id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.test", "hints.0.id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.test", "files.0.id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.test", "files.0.location"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.test", "files.0.contentb64"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.test", "files.1.id"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.test", "files.1.location"),
					resource.TestCheckResourceAttrSet("ctfd_challenge.test", "files.1.content"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ctfd_challenge.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "ctfd_challenge" "test" {
	name        = "Test Challenge"
	category    = "acc"
	description = <<-EOT
		This is some content.

		And it is multiline :o
	EOT
	value    = 500
    decay    = 17
    minimum  = 50
    state    = "visible"

	flags = [{
		content = "CTFER{acc_test}"
	}]
	topics = [
		"Acceptance"
	]
	tags = [
		"acceptance",
		"testing"
	]
	hints = [{
		content = "C'mon this is a test dude..."
	}, {
		content = "There is nothing to find here, it is just a test !"
		cost = 50
	}]
	files = [{
		name    = "something.txt",
		content = "I won't be really usefull as a file, but I tried my best :)"
	}]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ctfd_challenge.test", "files.#", "1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
