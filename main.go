package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/labd/terraform-provider-bluestonepim/internal/provider"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"
	commit  string = "snapshot"
)

func main() {
	debugFlag := flag.Bool("debug", false, "Start provider in debug mode.")
	flag.Parse()

	fullVersion := fmt.Sprintf("%s (%s)", version, commit)

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/labd/bluestonepim",
		Debug:   *debugFlag,
	}

	err := providerserver.Serve(context.Background(), provider.New(fullVersion, *debugFlag), opts)
	if err != nil {
		log.Fatal(err)
	}
}
