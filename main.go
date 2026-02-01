package main

import (
	"context"
	"flag"
	"log"

	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/ctfer-io/ctfd",
		Debug:   debug,
	}

	ctx := context.Background()

	shutdown, err := provider.SetupOtelSDK(ctx, version)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	if err := providerserver.Serve(ctx, provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
