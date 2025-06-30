package generator

import (
	"context"

	"github.com/cultureamp/terraform-buildkite-plugin/pkg/schema/handler"
	"github.com/cultureamp/terraform-buildkite-plugin/pkg/schema/schema"
	"github.com/spf13/cobra"
)

// Generator generates a plugin schema file.
type Generator interface {
	// GenerateSchema runs the schema generation logic and writes the output file.
	GenerateSchema(ctx context.Context, s schema.Config) error
}

type cobraWithRunFunc func(RunE func(cmd *cobra.Command, args []string) error) *cobra.Command

// generator implements Generator.
type generator struct {
	opts    *handler.HandleOptions
	cmd     cobraWithRunFunc
	handler handler.Handler
}

// ConfigOption configures a generator instance.
type ConfigOption func(*generator)

func WithHandler(h handler.Handler) ConfigOption {
	return func(g *generator) {
		g.handler = h
	}
}

// WithCommand sets a custom cobra command for the generator.
func WithCommand(c func(RunE func(cmd *cobra.Command, args []string) error) *cobra.Command) ConfigOption {
	return func(g *generator) {
		if c != nil {
			g.cmd = c
		}
	}
}

// defaultCommand returns a default cobra command for schema generation.
func defaultCommand(o *string) func(RunE func(cmd *cobra.Command, args []string) error) *cobra.Command {
	return func(RunE func(cmd *cobra.Command, args []string) error) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "plugin-schema-generator",
			Short: "A CLI tool to generate Buildkite plugin schemas",
			Long:  `This tool generates a Buildkite plugin schema in YAML format.`,
			RunE:  RunE,
		}
		cmd.Flags().StringVarP(
			o,
			"output",
			"o",
			"plugin.yml",
			"Output file for the generated schema",
		)
		return cmd
	}
}

// New creates a new Generator with the provided options.
func New(opts ...ConfigOption) Generator {
	g := &generator{
		opts:    &handler.HandleOptions{},
		handler: handler.New(),
	}
	for _, opt := range opts {
		opt(g)
	}
	if g.cmd == nil {
		g.cmd = defaultCommand(&g.opts.OutputFile)
	}
	return g
}

// GenerateSchema runs the cobra command to generate the schema.
func (g *generator) GenerateSchema(ctx context.Context, s schema.Config) error {
	return g.cmd(g.handler.Handle(s, g.opts)).ExecuteContext(ctx)
}
