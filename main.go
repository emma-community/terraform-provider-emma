package main

import (
	"context"
	"flag"
	"github.com/emma-community/terraform-provider-emma/internal/emma"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run "go generate" to format provider terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name emma -examples-dir ./examples/

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		// NOTE: This is not a typical Terraform Registry provider address,
		// such as registry.terraform.io/hashicorp/emma. This specific
		// provider address is used in these tutorials in conjunction with a
		// specific Terraform CLI configuration for manual development testing
		// of this provider.
		Address: "hashicorp.com/edu/emma",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), emma.New(), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
