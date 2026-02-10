package provider_test

import (
	"context"
	"log"
	"testing"

	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "ctfd" {}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"ctfd": providerserver.NewProtocol6WithError(provider.New("test", nil)()),
	}
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	out, err := provider.SetupOTelSDK(ctx, "test")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := out.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	m.Run()
}
