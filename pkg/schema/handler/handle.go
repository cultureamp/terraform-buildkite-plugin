package handler

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xphir/terraform-buildkite-plugin/pkg/schema/caller"
	"github.com/xphir/terraform-buildkite-plugin/pkg/schema/schema"
)

type HandleOptions struct {
	OutputFile string `validate:"required,extension=yaml yml"`
}

type Handler interface {
	Handle(schema schema.Config, opts *HandleOptions) func(cmd *cobra.Command, args []string) error
}

type handle struct {
	caller caller.Caller
}

// ConfigOption configures a generator instance.
type ConfigOption func(*handle)

// WithCaller sets a custom caller for the handler.
func WithCaller(c caller.Caller) ConfigOption {
	return func(g *handle) {
		g.caller = c
	}
}

// New creates a new Generator with the provided options.
func New(opts ...ConfigOption) Handler {
	g := &handle{
		caller: caller.New(),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// HandleCommand returns a cobra RunE function that generates the plugin schema file.
// It validates options, generates the schema, and writes the output file.
// The output and status messages are written to the command's configured stdout.
func (h *handle) Handle(s schema.Config, opts *HandleOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		// Validate CLI options before proceeding.
		if err := ValidateOptions(opts); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Generating plugin schema to %s\n", opts.OutputFile)

		pluginSchema, err := s.GeneratePluginSchema()
		if err != nil {
			return fmt.Errorf("error generating plugin schema: %w", err)
		}

		callerPath, err := h.caller.CallPath()
		if err != nil {
			return fmt.Errorf("failed to determine caller path: %w", err)
		}

		if err = pluginSchema.WriteFile(opts.OutputFile, callerPath); err != nil {
			return fmt.Errorf("error writing schema to file: %w", err)
		}

		fmt.Fprintf(out, "✅ Plugin schema successfully generated and saved to %s\n", opts.OutputFile)
		return nil
	}
}
