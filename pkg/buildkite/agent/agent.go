package agent

import (
	"context"
	"fmt"
	"os/exec"
)

type Agent interface {
	UploadPipeline(ctx context.Context, pipeline string) (*string, error)
	Annotate(ctx context.Context, opts ...AnnotateOptions) (*string, error)
	AnnotateWithTemplate(ctx context.Context, templatePath string, data any, opts ...AnnotateOptions) (*string, error)
}

type config struct {
	command CommandFn
}

// ConfigOptions allows functional options for customizing config.
type ConfigOptions func(*config)

// WithCommandFn allows injecting a custom CommandFn (e.g., for testing).
func WithCommandFn(fn CommandFn) ConfigOptions {
	return func(r *config) {
		if fn != nil {
			r.command = fn
		}
	}
}

// NewAgent creates a new instance of the Buildkite runner with the provided configuration options.
func NewAgent(opts ...ConfigOptions) Agent {
	runner := &config{
		command: exec.Command,
	}
	for _, opt := range opts {
		opt(runner)
	}
	return runner
}

// UploadPipeline allows you to upload a Buildkite pipeline configuration file.
func (a *config) UploadPipeline(ctx context.Context, pipeline string) (*string, error) {
	return a.runCommand(ctx, "buildkite-agent", "pipeline", "upload", pipeline)
}

// Annotate allows you to add annotations to the Buildkite build.
func (a *config) Annotate(ctx context.Context, opts ...AnnotateOptions) (*string, error) {
	// Set default options
	config := annotateConfig{
		style: StyleInfo, // Default style is "info"
	}

	// Apply options to the configuration
	for _, opt := range opts {
		opt(&config)
	}

	// Build the command arguments
	args := []string{"annotate", config.message, "--style", string(config.style), "--context", config.context}
	if config.artifact != "" {
		args = append(args, "--artifact", config.artifact)
	}
	if config.append {
		args = append(args, "--append ")
	}
	// Run the command using the injected function
	return a.runCommand(ctx, "buildkite-agent", args...)
}

// AnnotateWithTemplate allows you to annotate a Buildkite build using a template.
func (a *config) AnnotateWithTemplate(
	ctx context.Context,
	templatePath string,
	data any,
	opts ...AnnotateOptions, // TODO: make it so WithMessage is not a valid option here
) (*string, error) {
	// Render the template
	renderedMessage, err := a.renderTemplate(templatePath, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Add the rendered message as a MessageOption
	opts = append(opts, WithMessage(renderedMessage))

	// Use the Annotate function with the provided options
	return a.Annotate(ctx, opts...)
}
