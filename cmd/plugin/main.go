// Package main is the entry point for the Terraform Buildkite plugin.
//
// It loads configuration, sets up logging, and coordinates plugin execution
// within Buildkite pipelines.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cultureamp/terraform-buildkite-plugin/internal/common"
	"github.com/cultureamp/terraform-buildkite-plugin/internal/plugin"
	"github.com/cultureamp/terraform-buildkite-plugin/pkg/buildkite/group"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//nolint:gochecknoglobals // Variables set at build time to provide plugin metadata.
var (
	// Version of plugin - set at build time.
	version = "dev"
	// name of the plugin - set at build time.
	name = "terraform-buildkite-plugin"
	// Commit hash of the plugin - set at build time.
	commit = "unknown"
	// date of the plugin - set at build time.
	date = "unknown"
)

// main is the entry point for the plugin.
//
// It sets up logging, loads configuration, handles test mode, and runs the plugin.
func main() {
	ctx := context.Background()

	pluginContext := &plugin.Context{
		Name:    name,
		Version: version,
		Date:    date,
		Commit:  commit,
	}

	// Configure the logger for console output with CI-friendly formatting.
	configureLogger(ctx)

	group.ClosedF("running %s version %s", pluginContext.Name, pluginContext.Version)

	log.Debug().Str("commit", pluginContext.Commit).Str("date", pluginContext.Date).Msg("Plugin metadata")

	handler := plugin.NewHandler()

	result, err := handler.Handle(ctx, pluginContext)
	if err != nil {
		group.OpenCurrent()
		log.Fatal().Err(err).Msg("Failed to handle plugin execution")
	}
	log.Info().Str("status", result.GetName()).Msg("pluging exiting with status")
	os.Exit(result.ToInt())
}

// configureLogger sets up zerolog for console output with CI-friendly formatting.
//
// It configures the logger for coloured output, omits timestamps, and attaches the context.
func configureLogger(ctx context.Context) {
	//nolint:reassign // overriding the global logger for convenience
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:             os.Stdout,
			NoColor:         false,
			PartsExclude:    []string{"time"},
			FormatFieldName: func(i any) string { return fmt.Sprintf("%s:", i) },
		},
	).With().Ctx(ctx).Logger()
	// We create the logger first and set the log level afterwards so that any logs caused by `ParseLogLevel` are properly formatted
	//nolint:reassign // overriding the global logger for convenience
	log.Logger = log.Logger.Level((common.ParseLogLevel("LOG_LEVEL", zerolog.DebugLevel)))
}
