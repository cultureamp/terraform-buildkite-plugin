package main

import (
	"context"
	"fmt"
	"os"

	"github.com/xphir/terraform-buildkite-plugin/internal/config"
	"github.com/xphir/terraform-buildkite-plugin/pkg/schema/generator"
	"github.com/xphir/terraform-buildkite-plugin/pkg/schema/schema"
)

func main() {
	ctx := context.Background()

	schema := schema.New(
		schema.WithProperties(
			&schema.PluginProperties{
				Name: "Terraform Buildkite Plugin",
				Description: `A Buildkite plugin for processing a terraform working directory.
Allowing you to perform operations such as plan & apply.
With support for looping over multiple working directories,
Open Policy Agent checks against plans &
Buildkite annotations detailing the success or failure of the the operations.`,
				Author:       "https://github.com/xphir",
				Requirements: []string{"buildkite-agent", "terraform", "opa"},
			},
		),
		schema.WithSchema(&config.Plugin{}),
	)

	if err := generator.New().GenerateSchema(ctx, schema); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}
