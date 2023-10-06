package provider_test

import (
	"github.com/opentofu/terraform-plugin-framework/providerserver"
	"github.com/opentofu/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "ctfd" {
	username = "never gonna give you up"
	password = "never gonna let you down"
	host     = "https://example.ctfer.io"
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"ctfd": providerserver.NewProtocol6WithError(New("test")()),
	}
)
